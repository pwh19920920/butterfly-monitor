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

	// 全部 Pending 事件：按监控分组聚合，供关联告警使用
	groupIdForEventsMap, err := job.buildMonitorGroupIdForEventListMap(taskMap)
	if err != nil {
		logger.ErrorFormat(ctx, "buildMonitorGroupIdForEventListMap fail: %v", err)
		// 关联失败不阻断主告警发送
		groupIdForEventsMap = map[int64][]entity.MonitorTaskEvent{}
	}

	// 通道缓存：同一通道多桶复用
	channelCache := make(map[int64]*entity.AlertChannel)

	// 分桶：有用户组 → (用户组集合, 渠道)；无用户组 → (渠道)
	type bucket struct {
		channel       *entity.AlertChannel
		users         []entity.SysUser
		items         []common.AlertTemplateItem
		relationTexts map[string]string // 关联告警按链路 PathKey 去重后的文案
		eventIDs      map[int64]bool
	}
	bucketMap := make(map[string]*bucket)

	// 每个 event 的关联告警链路（与聚合分组无关，按 monitor_group 链路独立计算）
	relationAlertsMap := make(map[int64][]RelationAlert, len(events))

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
		relationAlerts := job.getEventRelationParams(task, taskMap, groupIdForEventsMap)
		relationAlertsMap[event.Id] = relationAlerts

		// 汇总用户一次，本 event 各桶共用
		groups := splitCSV(alert.AlertGroups)
		channels := splitCSV(alert.AlertChannels)
		users := make([]entity.SysUser, 0)
		seenUids := make(map[int64]bool)
		for _, gIdStr := range groups {
			gId, _ := parseInt64(gIdStr)
			if gId == 0 {
				continue
			}
			gus, _ := job.repository.AlertGroupUserRepository.SelectByGroupId(gId)
			uids := make([]int64, 0, len(gus))
			for _, gu := range gus {
				if !seenUids[gu.UserId] {
					seenUids[gu.UserId] = true
					uids = append(uids, gu.UserId)
				}
			}
			if us, err := job.repository.AlertGroupUserRepository.SelectUsersByUserIds(uids); err == nil {
				users = append(users, us...)
			}
		}

		// 分组键规范化：用户组排序后的 id 串（无用户组则为空）
		groupKey := ""
		if len(groups) > 0 {
			sorted := make([]string, 0, len(groups))
			sorted = append(sorted, groups...)
			sort.Strings(sorted)
			groupKey = strings.Join(sorted, ",")
		}

		for _, chIdStr := range channels {
			chId, _ := parseInt64(chIdStr)
			if chId == 0 {
				continue
			}
			ch, ok := channelCache[chId]
			if !ok {
				c, err := job.repository.AlertChannelRepository.GetById(chId)
				if err != nil || c == nil {
					continue
				}
				channelCache[chId] = c
				ch = c
			}
			// 桶键：有组用 组+渠道，无组只用 渠道
			var key string
			if groupKey == "" {
				key = fmt.Sprintf("c:%d", chId)
			} else {
				key = fmt.Sprintf("g:%s|c:%d", groupKey, chId)
			}
			bk, exists := bucketMap[key]
			if !exists {
				bk = &bucket{channel: ch, eventIDs: make(map[int64]bool), relationTexts: make(map[string]string)}
				// 用户组桶才注入 users；无组桶 users 为空（Webhook 不依赖收件人）
				if groupKey != "" {
					bk.users = users
				}
				bucketMap[key] = bk
			} else if groupKey != "" && len(bk.users) == 0 && len(users) > 0 {
				// 同键桶理论 users 相同，兜底补齐
				bk.users = users
			}
			bk.items = append(bk.items, common.AlertTemplateItem{
				TaskName: task.TaskName,
				HitRule:  event.AlertMsg,
			})
			bk.eventIDs[event.Id] = true
			// 关联告警按链路 PathKey 去重，整条消息共享一份
			for _, rel := range relationAlerts {
				if _, ok := bk.relationTexts[rel.PathKey]; !ok {
					bk.relationTexts[rel.PathKey] = rel.Text
				}
			}
		}
	}

	// 逐桶发送：每桶一条消息
	// 关联告警与 items 平级：整条消息一份（按链路去重）
	sentEventIDs := make(map[int64]bool)
	for _, bk := range bucketMap {
		ch := bk.channel
		h, ok := job.commonMap.GetChannelHandler(ctx, ch.Handler)
		if !ok {
			continue
		}
		// 收集本桶关联告警文案：生成阶段已按"含自身 route"为每个告警任务产出一条链路，
		// 同一物理链路的嵌套短链路（祖先前缀）需被最长链路覆盖，只保留最长一条
		type relItem struct {
			pathKey string
			text    string
		}
		relList := make([]relItem, 0, len(bk.relationTexts))
		for pk, text := range bk.relationTexts {
			if text != "" {
				relList = append(relList, relItem{pathKey: pk, text: text})
			}
		}
		// 规范化为带尾斜杠的 /seg/seg/，按段做前缀比较，避免 /1/ 误判为 /10/ 前缀
		norm := func(pk string) string {
			pk = strings.Trim(pk, "/")
			if pk == "" {
				return ""
			}
			return "/" + pk + "/"
		}
		// 按链路深度降序：长链路优先保留，输出顺序也跟着稳定
		sort.Slice(relList, func(i, j int) bool {
			return len(norm(relList[i].pathKey)) > len(norm(relList[j].pathKey))
		})
		relationTexts := make([]string, 0, len(relList))
		keptNorm := make([]string, 0, len(relList))
		for _, it := range relList {
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
				relationTexts = append(relationTexts, it.text)
				if normThis != "" {
					keptNorm = append(keptNorm, normThis)
				}
			}
		}
		tpl := conf.ResolveTemplate(ch.Template, ch.Handler)
		msg := ""
		if tpl != "" {
			if rendered, err := common.RenderAlertTemplateMulti(tpl, bk.items, relationTexts); err == nil {
				msg = rendered
			}
		}
		if msg == "" {
			// 无模板：按行拼接 HitRule + 关联告警
			lines := make([]string, 0, len(bk.items))
			for _, it := range bk.items {
				lines = append(lines, it.TaskName+"："+it.HitRule)
			}
			for _, rel := range relationTexts {
				lines = append(lines, "关联预警："+rel)
			}
			msg = strings.Join(lines, "\n")
		}
		if err := h.DispatchMessage(*ch, bk.users, msg); err == nil {
			for id := range bk.eventIDs {
				sentEventIDs[id] = true
			}
		}
	}

	// 推进成功发送的事件 NextAlertTime（各 event 按自身翻倍进度）
	for _, event := range events {
		if !sentEventIDs[event.Id] {
			continue
		}
		n := event.AlertCount
		span := conf.AlertSpan
		if span <= 0 {
			span = 300
		}
		shift := n
		if shift > 5 {
			shift = 5
		}
		span = span * (1 << shift)
		if span > 1800 {
			span = 1800
		}
		now := time.Now()
		// 只推进调度字段，不改 AlertMsg（保持原始 hitRule）
		if err := job.repository.MonitorTaskEventRepository.Modify(event.Id, &entity.MonitorTaskEvent{
			PreAlertTime:  &common.LocalTime{Time: now},
			NextAlertTime: &common.LocalTime{Time: now.Add(time.Duration(span) * time.Second)},
			AlertCount:    n + 1,
		}); err != nil {
			// 告警事件未推进：下次仍按旧 NextAlertTime 调度，可能重复发送，必须留痕
			logger.ErrorFormat(ctx, "event Modify fail eventId=%d: %v", event.Id, err)
		}
	}
	return "execute complete"
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
		monitorGroup, err := job.repository.MonitorGroupRepository.GetById(groupId)
		if err != nil || monitorGroup == nil {
			continue
		}

		// route 形如 /1/2/3/，保留自身节点，得到根 -> ... -> 自身的完整链路 id 序列
		route := strings.TrimSpace(monitorGroup.Route)
		if route == "" {
			continue
		}
		// 子（末节点）在前、根在后：split 后反转
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
		if len(chainIds) == 0 {
			continue
		}

		// 链路上每个分组节点：收集该分组下所有 Pending 告警任务名，聚成一段
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
					// 跨分组链路层去重计数
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
