package job

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorEventCheckJob 告警事件通知发送任务
type MonitorEventCheckJob struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	xxlExec    xxl.Executor
	alertConf  application.AlertConfApplication
	commonMap  *application.CommonMapApplication
}

func (job *MonitorEventCheckJob) RegisterExecJob() {
	if job.xxlExec == nil {
		return
	}
	job.xxlExec.RegTask("eventCheck", job.eventCheck)
}

// sendBucket 分桶：按 (用户组集合, 通道) 聚合事件，一桶一条消息
type sendBucket struct {
	channel       *entity.AlertChannel
	users         []entity.SysUser
	items         []common.AlertTemplateItem
	relationTexts map[string]string // 关联告警按链路 PathKey 去重后的文案
	eventIDs      map[int64]bool
}

func (job *MonitorEventCheckJob) eventCheck(ctx context.Context, _ *xxl.RunReq) string {
	events, err := job.repository.MonitorTaskEventRepository.FindEventJob()
	if err != nil || len(events) == 0 {
		return "execute complete"
	}
	conf, err := job.alertConf.Cover2AlertConf(ctx)
	if err != nil {
		return "exec failure conf"
	}

	// 全量任务：关联告警需要按 monitor_group 反查
	allTasks, err := job.repository.MonitorTaskRepository.SelectAll()
	if err != nil {
		logger.ErrorFormat(ctx, "SelectAll tasks fail: %v", err)
		return "exec failure tasks"
	}
	taskMap := make(map[int64]entity.MonitorTask, len(allTasks))
	for _, t := range allTasks {
		taskMap[t.Id] = t
	}

	// 全部 Pending 事件按监控分组聚合，供关联告警使用（失败不阻断主告警发送）
	groupIdForEventsMap, err := job.buildMonitorGroupIdForEventListMap(taskMap)
	if err != nil {
		logger.ErrorFormat(ctx, "buildMonitorGroupIdForEventListMap fail: %v", err)
		groupIdForEventsMap = map[int64][]entity.MonitorTaskEvent{}
	}

	channelCache := make(map[int64]*entity.AlertChannel)
	bucketMap := make(map[string]*sendBucket)
	relationAlertsMap := make(map[int64][]RelationAlert, len(events))
	// 缓存分组→用户映射：同一告警分组被多个事件共享时免重复查库
	groupUserCache := make(map[int64][]int64)
	userCache := make(map[int64]entity.SysUser)

	// 预加载全量监控分组，避免 getEventRelationParams 逐事件调用 GetById 产生 N+1
	allGroupIds := collectMonitorGroupIds(taskMap)
	allGroups, _ := job.repository.MonitorGroupRepository.SelectByIds(allGroupIds)
	groupMap := make(map[int64]entity.MonitorGroup, len(allGroups))
	for _, g := range allGroups {
		groupMap[g.Id] = g
	}

	for _, event := range events {
		task, ok := taskMap[event.TaskId]
		if !ok {
			// 兜底再查一次
			t, err := job.repository.MonitorTaskRepository.GetById(event.TaskId)
			if err != nil || t == nil {
				continue
			}
			task = *t
			taskMap[task.Id] = task
		}
		alert, err := job.repository.MonitorTaskAlertRepository.GetByTaskId(event.TaskId)
		if err != nil || alert == nil {
			continue
		}

		// 关联告警：沿 monitor_group 祖先链路收集该链路上所有告警任务（按链路去重）
		relationAlertsMap[event.Id] = job.getEventRelationParams(task, taskMap, groupIdForEventsMap, groupMap)

		// 汇总用户一次，本 event 各桶共用（命中分组缓存则免查库）
		users := job.collectAlertUsers(ctx, alert.AlertGroups, groupUserCache, userCache)
		groupKey := normalizeGroupKey(splitCSV(alert.AlertGroups))

		for _, chIdStr := range splitCSV(alert.AlertChannels) {
			chId, _ := parseInt64(chIdStr)
			if chId == 0 {
				continue
			}
			ch := job.resolveChannel(ctx, chId, channelCache)
			if ch == nil {
				continue
			}
			bk := job.ensureBucket(bucketMap, bucketKey(groupKey, chId), ch, users, groupKey != "")
			bk.items = append(bk.items, common.AlertTemplateItem{
				TaskName: task.TaskName,
				HitRule:  event.AlertMsg,
			})
			bk.eventIDs[event.Id] = true
			// 关联告警按链路 PathKey 去重，整条消息共享一份
			for _, rel := range relationAlertsMap[event.Id] {
				if _, ok := bk.relationTexts[rel.PathKey]; !ok {
					bk.relationTexts[rel.PathKey] = rel.Text
				}
			}
		}
	}

	// 逐桶发送：每桶一条消息，关联告警与 items 平级
	sentEventIDs := job.dispatchBuckets(ctx, conf, bucketMap)
	// 推进成功发送的事件 NextAlertTime（各 event 按自身翻倍进度）
	job.advanceSentEvents(ctx, events, sentEventIDs, conf)
	return "execute complete"
}

// collectAlertUsers 汇总告警分组下的用户（跨组去重，批量查）。
// groupUserCache 缓存「分组 id → 用户 id 列表」，userCache 缓存「用户 id → 用户实体」，
// 多个事件共享同一告警分组/用户时免去重复查询 t_alert_group_user / t_sys_user。
func (job *MonitorEventCheckJob) collectAlertUsers(
	ctx context.Context,
	alertGroupsCSV string,
	groupUserCache map[int64][]int64,
	userCache map[int64]entity.SysUser,
) []entity.SysUser {
	groups := splitCSV(alertGroupsCSV)

	// 第一遍：收集本集合涉及的分组用户 id（未命中缓存的批量查，已命中的直接用）
	needUids := make([]int64, 0)
	seenUids := make(map[int64]bool)
	for _, gIdStr := range groups {
		gId, _ := parseInt64(gIdStr)
		if gId == 0 {
			continue
		}
		ids, ok := groupUserCache[gId]
		if !ok {
			gus, _ := job.repository.AlertGroupUserRepository.SelectByGroupId(gId)
			ids = make([]int64, 0, len(gus))
			for _, gu := range gus {
				ids = append(ids, gu.UserId)
			}
			groupUserCache[gId] = ids
		}
		for _, uid := range ids {
			if !seenUids[uid] {
				seenUids[uid] = true
				needUids = append(needUids, uid)
			}
		}
	}

	// 第二遍：用户详情，未缓存的批量补一次
	missing := make([]int64, 0)
	for _, uid := range needUids {
		if _, ok := userCache[uid]; !ok {
			missing = append(missing, uid)
		}
	}
	if len(missing) > 0 {
		if us, err := job.repository.AlertGroupUserRepository.SelectUsersByUserIds(missing); err == nil {
			for _, u := range us {
				userCache[u.Id] = u
			}
		}
	}

	users := make([]entity.SysUser, 0, len(needUids))
	for _, uid := range needUids {
		if u, ok := userCache[uid]; ok {
			users = append(users, u)
		}
	}
	return users
}

// resolveChannel 按通道 id 取通道实体，命中缓存优先。
func (job *MonitorEventCheckJob) resolveChannel(ctx context.Context, chId int64, cache map[int64]*entity.AlertChannel) *entity.AlertChannel {
	if ch, ok := cache[chId]; ok {
		return ch
	}
	c, err := job.repository.AlertChannelRepository.GetById(chId)
	if err != nil || c == nil {
		return nil
	}
	cache[chId] = c
	return c
}

// normalizeGroupKey 用户组排序后的 id 串（无用户组则为空），作为分桶维度之一。
func normalizeGroupKey(groups []string) string {
	if len(groups) == 0 {
		return ""
	}
	sorted := make([]string, len(groups))
	copy(sorted, groups)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

// bucketKey 分桶键：有组用 组+渠道，无组只用渠道。
func bucketKey(groupKey string, chId int64) string {
	if groupKey == "" {
		return fmt.Sprintf("c:%d", chId)
	}
	return fmt.Sprintf("g:%s|c:%d", groupKey, chId)
}

// ensureBucket 获取或创建桶。有用户组的桶才注入 users；无组桶 users 为空（Webhook 不依赖收件人）。
func (job *MonitorEventCheckJob) ensureBucket(bucketMap map[string]*sendBucket, key string, ch *entity.AlertChannel, users []entity.SysUser, hasGroup bool) *sendBucket {
	bk, exists := bucketMap[key]
	if !exists {
		bk = &sendBucket{
			channel:       ch,
			eventIDs:      make(map[int64]bool),
			relationTexts: make(map[string]string),
		}
		if hasGroup {
			bk.users = users
		}
		bucketMap[key] = bk
	} else if hasGroup && len(bk.users) == 0 && len(users) > 0 {
		// 同键桶理论 users 相同，兜底补齐
		bk.users = users
	}
	return bk
}

// dispatchBuckets 逐桶渲染并发送，返回成功发送的 event id 集合。
func (job *MonitorEventCheckJob) dispatchBuckets(ctx context.Context, conf *types.AlertConfObject, bucketMap map[string]*sendBucket) map[int64]bool {
	sentEventIDs := make(map[int64]bool)
	for _, bk := range bucketMap {
		h, ok := job.commonMap.GetChannelHandler(ctx, bk.channel.Handler)
		if !ok {
			continue
		}
		msg := renderBucketMessage(conf, bk)
		if err := h.DispatchMessage(*bk.channel, bk.users, msg); err == nil {
			for id := range bk.eventIDs {
				sentEventIDs[id] = true
			}
		}
	}
	return sentEventIDs
}

// renderBucketMessage 渲染桶消息：模板优先，无模板降级为按行拼接。
// 关联告警链路做祖先覆盖去重——同一物理链路的嵌套短链路被最长链路覆盖，只保留最长一条。
func renderBucketMessage(conf *types.AlertConfObject, bk *sendBucket) string {
	relationTexts := dedupRelationTexts(bk.relationTexts)

	tpl := conf.ResolveTemplate(bk.channel.Template, bk.channel.Handler)
	if tpl != "" {
		if rendered, err := common.RenderAlertTemplateMulti(tpl, bk.items, relationTexts); err == nil {
			return rendered
		}
	}
	// 无模板：按行拼接 HitRule + 关联告警
	lines := make([]string, 0, len(bk.items)+len(relationTexts))
	for _, it := range bk.items {
		lines = append(lines, it.TaskName+"："+it.HitRule)
	}
	for _, rel := range relationTexts {
		lines = append(lines, "关联预警："+rel)
	}
	return strings.Join(lines, "\n")
}

// relationItem 关联告警链路项（PathKey 用于前缀覆盖去重）
type relationItem struct {
	pathKey string
	text    string
}

// dedupRelationTexts 关联告警链路祖先覆盖去重：
// 规范化为带尾斜杠的 /seg/seg/ 后按段做前缀比较，避免 /1/ 误判为 /10/ 前缀；
// 长链路优先保留，祖先链路（更短、被最长链路包含）丢弃。
func dedupRelationTexts(relationTexts map[string]string) []string {
	list := make([]relationItem, 0, len(relationTexts))
	for pk, text := range relationTexts {
		if text != "" {
			list = append(list, relationItem{pathKey: pk, text: text})
		}
	}
	norm := func(pk string) string {
		pk = strings.Trim(pk, "/")
		if pk == "" {
			return ""
		}
		return "/" + pk + "/"
	}
	// 按链路深度降序：长链路优先保留，输出顺序稳定
	sort.Slice(list, func(i, j int) bool {
		return len(norm(list[i].pathKey)) > len(norm(list[j].pathKey))
	})
	out := make([]string, 0, len(list))
	keptNorm := make([]string, 0, len(list))
	for _, it := range list {
		normThis := norm(it.pathKey)
		covered := false
		// 已保留链路中若存在 this 的后代（更深路径以 this 为前缀），this 作为祖先被覆盖，丢弃
		for _, kn := range keptNorm {
			if normThis != "" && kn != normThis && strings.HasPrefix(kn, normThis) {
				covered = true
				break
			}
		}
		if !covered {
			out = append(out, it.text)
			if normThis != "" {
				keptNorm = append(keptNorm, normThis)
			}
		}
	}
	return out
}

// nextAlertSpan 计算下次告警间隔（翻倍退避，上限 types.DefaultMaxAlertSpan）。
// shift = min(AlertCount, maxAlertShift)，间隔 = AlertSpan * 2^shift。
func nextAlertSpan(conf *types.AlertConfObject, alertCount int32) int64 {
	span := conf.AlertSpan
	shift := int64(alertCount)
	maxShift := conf.MaxAlertShift
	if shift > maxShift {
		shift = maxShift
	}
	span = span * (1 << shift)
	if span > types.DefaultMaxAlertSpan {
		span = types.DefaultMaxAlertSpan
	}
	return span
}

// advanceSentEvents 推进成功发送事件的 NextAlertTime（按各自翻倍进度），只改调度字段不改 AlertMsg。
func (job *MonitorEventCheckJob) advanceSentEvents(ctx context.Context, events []entity.MonitorTaskEvent, sentEventIDs map[int64]bool, conf *types.AlertConfObject) {
	now := time.Now()
	for _, event := range events {
		if !sentEventIDs[event.Id] {
			continue
		}
		span := nextAlertSpan(conf, event.AlertCount)
		if err := job.repository.MonitorTaskEventRepository.Modify(event.Id, &entity.MonitorTaskEvent{
			PreAlertTime:  &common.LocalTime{Time: now},
			NextAlertTime: &common.LocalTime{Time: now.Add(time.Duration(span) * time.Second)},
			AlertCount:    event.AlertCount + 1,
		}); err != nil {
			// 告警事件未推进：下次仍按旧 NextAlertTime 调度，可能重复发送，必须留痕
			logger.ErrorFormat(ctx, "event Modify fail eventId=%d: %v", event.Id, err)
		}
	}
}

// buildMonitorGroupIdForEventListMap 将 Pending 事件按任务的 monitor_group 聚合
// key=监控分组 id，value=该分组下所有 Pending 事件
func (job *MonitorEventCheckJob) buildMonitorGroupIdForEventListMap(taskMap map[int64]entity.MonitorTask) (map[int64][]entity.MonitorTaskEvent, error) {
	groupIdForEventsMap := make(map[int64][]entity.MonitorTaskEvent)
	eventAll, err := job.repository.MonitorTaskEventRepository.FindPendingEventAll()
	if err != nil {
		return groupIdForEventsMap, err
	}
	for _, event := range eventAll {
		task, ok := taskMap[event.TaskId]
		if !ok || strings.TrimSpace(task.MonitorGroup) == "" {
			continue
		}
		for _, groupStr := range splitCSV(task.MonitorGroup) {
			groupId, _ := parseInt64(groupStr)
			if groupId == 0 {
				continue
			}
			groupIdForEventsMap[groupId] = append(groupIdForEventsMap[groupId], event)
		}
	}
	return groupIdForEventsMap, nil
}

// RelationAlert 一条关联告警链路
type RelationAlert struct {
	PathKey string // 链路标识：祖先分组 id 序列，用于同链路去重
	Text    string // 渲染文案：[近分组] 任务 -> [远分组] 任务
}

// getEventRelationParams 计算关联告警链路
// 规则：任务所在每个分组节点定义一条链路（根 -> ... -> 该节点，含自身）。
// 链路上每个分组节点下所有 Pending 告警任务聚成一段（同分组多任务用顿号连），
// 段之间用 " -> " 串接（子在前、根在后），跳过无告警的节点。
// 链路上至少两个不同告警任务（跨分组去重计数）才输出——单点孤立链路剔除，
// 即任务挂在根节点或链路上只有自己告警时不产生关联告警。
func (job *MonitorEventCheckJob) getEventRelationParams(
	task entity.MonitorTask,
	taskMap map[int64]entity.MonitorTask,
	groupIdForEventsMap map[int64][]entity.MonitorTaskEvent,
	groupMap map[int64]entity.MonitorGroup,
) []RelationAlert {
	relations := make([]RelationAlert, 0)
	if strings.TrimSpace(task.MonitorGroup) == "" {
		return relations
	}

	// 一个任务可挂多个分组；每个分组对应一条链路
	for _, groupStr := range splitCSV(task.MonitorGroup) {
		groupId, _ := parseInt64(groupStr)
		if groupId == 0 {
			continue
		}
		monitorGroup, ok := groupMap[groupId]
		if !ok {
			continue
		}

		// route 形如 /1/2/3/，保留自身节点，得到根 -> ... -> 自身的完整链路 id 序列
		route := strings.TrimSpace(monitorGroup.Route)
		if route == "" {
			continue
		}
		// 子（末节点）在前、根在后：split 后反转
		chainIds := parseChainIds(route)
		if len(chainIds) == 0 {
			continue
		}

		// 链路上每个分组节点：收集该分组下所有 Pending 告警任务名，聚成一段
		segments, taskNodeCount := job.collectChainSegments(taskMap, groupIdForEventsMap, chainIds)
		// 链路上至少两个不同告警任务才算关联告警（单点孤立链路剔除，含自身根节点情形）
		if taskNodeCount < 2 {
			continue
		}
		relations = append(relations, RelationAlert{
			PathKey: strings.Trim(route, "/"),
			Text:    strings.Join(segments, " -> "),
		})
	}
	return relations
}

// parseChainIds 从 route（/1/2/3/）解析祖先链路 id 序列，子在前、根在后。
func parseChainIds(route string) []int64 {
	parts := strings.Split(strings.Trim(route, "/"), "/")
	chainIds := make([]int64, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "" {
			continue
		}
		id, _ := parseInt64(parts[i])
		if id != 0 {
			chainIds = append(chainIds, id)
		}
	}
	return chainIds
}

// collectChainSegments 收集链路各分组节点的告警任务名段，返回 (段列表, 跨分组去重任务数)。
func (job *MonitorEventCheckJob) collectChainSegments(
	taskMap map[int64]entity.MonitorTask,
	groupIdForEventsMap map[int64][]entity.MonitorTaskEvent,
	chainIds []int64,
) ([]string, int) {
	segments := make([]string, 0, len(chainIds))
	// 整条链路上不同告警任务的总数（跨分组去重），用于"至少两个节点"判定
	taskSeen := make(map[int64]bool)
	taskNodeCount := 0
	for _, gid := range chainIds {
		events, ok := groupIdForEventsMap[gid]
		if !ok || len(events) == 0 {
			continue
		}
		seen := make(map[int64]bool)
		names := make([]string, 0)
		for _, ev := range events {
			if seen[ev.TaskId] {
				continue
			}
			seen[ev.TaskId] = true
			if rel, ok := taskMap[ev.TaskId]; ok && rel.TaskName != "" {
				names = append(names, rel.TaskName)
				if !taskSeen[ev.TaskId] {
					taskSeen[ev.TaskId] = true
					taskNodeCount++
				}
			}
		}
		if len(names) > 0 {
			segments = append(segments, strings.Join(names, "、"))
		}
	}
	return segments, taskNodeCount
}

// collectMonitorGroupIds 从任务集合中收集所有唯一的监控分组 ID。
// 用于预加载，避免 getEventRelationParams 逐事件调用 GetById 产生 N+1。
func collectMonitorGroupIds(taskMap map[int64]entity.MonitorTask) []int64 {
	seen := make(map[int64]bool)
	ids := make([]int64, 0)
	for _, task := range taskMap {
		for _, gStr := range splitCSV(task.MonitorGroup) {
			id, _ := parseInt64(gStr)
			if id != 0 && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}
