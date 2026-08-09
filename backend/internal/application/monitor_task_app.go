package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/infrastructure/support"
	"butterfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// MonitorTaskApplication 监控任务应用服务
type MonitorTaskApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaHandler
	commonMap      *CommonMapApplication
}

// NewMonitorTaskApplication 创建监控任务应用服务
func NewMonitorTaskApplication(sequence *snowflake.Node, repository *persistence.Repository, grafanaHandler *support.GrafanaHandler, commonMap *CommonMapApplication) MonitorTaskApplication {
	return MonitorTaskApplication{sequence: sequence, repository: repository, grafanaHandler: grafanaHandler, commonMap: commonMap}
}

// Query 分页查询（附带规则运行态 taskAlertStatus）
func (app *MonitorTaskApplication) Query(ctx context.Context, req *types.MonitorTaskQueryRequest) (int64, []types.MonitorTaskListItem, error) {
	total, data, err := app.repository.MonitorTaskRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "MonitorTaskRepository.Select() happen error for", err)
		return total, nil, err
	}
	items := make([]types.MonitorTaskListItem, 0, len(data))
	if len(data) == 0 {
		return total, items, nil
	}

	taskIds := make([]int64, 0, len(data))
	for _, t := range data {
		taskIds = append(taskIds, t.Id)
	}
	alerts, err := app.repository.MonitorTaskAlertRepository.BatchGetByTaskIds(taskIds)
	if err != nil {
		logger.Error(ctx, "MonitorTaskAlertRepository.BatchGetByTaskIds() happen error for", err)
		return total, nil, err
	}

	alertStatusMap := make(map[int64]entity.MonitorTaskAlertStatus, len(alerts))
	firstFlagMap := make(map[int64]*common.LocalTime, len(alerts))
	for _, a := range alerts {
		// 软删除的配置视为不存在
		if a.Deleted == common.DeletedTrue {
			continue
		}
		alertStatusMap[a.TaskId] = a.AlertStatus
		firstFlagMap[a.TaskId] = a.FirstFlagTime
	}
	for _, t := range data {
		status := entity.MonitorTaskAlertStatusNormal
		if s, ok := alertStatusMap[t.Id]; ok {
			status = s
		}
		items = append(items, types.MonitorTaskListItem{
			MonitorTask:     t,
			TaskAlertStatus: status,
			FirstFlagTime:   firstFlagMap[t.Id],
		})
	}
	return total, items, nil
}

// GetById 按主键查询详情（含 dashboard 关联与告警规则）
func (app *MonitorTaskApplication) GetById(ctx context.Context, id int64) (*types.MonitorTaskQueryResponse, error) {
	task, err := app.repository.MonitorTaskRepository.GetById(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}
	return app.coverToQueryResponse(task)
}

// GetByTaskKey 按任务标识查询任务本体（用于 dataPush 把 taskKey 转 taskId）
func (app *MonitorTaskApplication) GetByTaskKey(ctx context.Context, taskKey string) (*entity.MonitorTask, error) {
	return app.repository.MonitorTaskRepository.SelectByTaskKey(taskKey)
}

func (app *MonitorTaskApplication) coverToQueryResponse(task *entity.MonitorTask) (*types.MonitorTaskQueryResponse, error) {
	resp := &types.MonitorTaskQueryResponse{MonitorTask: *task}

	if task.ExecParams != "" {
		var execParams types.MonitorTaskExecParams
		if err := json.Unmarshal([]byte(task.ExecParams), &execParams); err == nil {
			resp.TaskExecParams = execParams
		}
	}

	dashboardTasks, err := app.repository.MonitorDashboardTaskRepository.FindByTaskId(task.Id)
	if err != nil {
		return nil, err
	}
	dashboards := make([]string, 0, len(dashboardTasks))
	for _, item := range dashboardTasks {
		dashboards = append(dashboards, strconv.FormatInt(item.DashboardId, 10))
	}
	resp.Dashboards = dashboards

	alert, err := app.repository.MonitorTaskAlertRepository.GetByTaskId(task.Id)
	if err != nil {
		return nil, err
	}
	// 无告警配置视为正常
	resp.TaskAlertStatus = entity.MonitorTaskAlertStatusNormal
	if alert != nil {
		alertReq := types.MonitorTaskAlertCreateRequest{MonitorTaskAlert: *alert}
		if alert.Params != "" {
			var checkParams []entity.MonitorAlertCheckParams
			if err := json.Unmarshal([]byte(alert.Params), &checkParams); err == nil {
				alertReq.CheckParams = checkParams
			}
		}
		if alert.AlertChannels != "" {
			alertReq.AlertChannels = strings.Split(alert.AlertChannels, ",")
		}
		if alert.AlertGroups != "" {
			alertReq.AlertGroups = strings.Split(alert.AlertGroups, ",")
		}
		resp.TaskAlert = alertReq
		resp.TaskAlertStatus = alert.AlertStatus
	}
	return resp, nil
}

// needAlertGroups 所选通道是否需要报警分组。
// 规则：未选通道默认需要；只要存在非 Webhook 通道就需要；全为 Webhook 则不需要。
func (app *MonitorTaskApplication) needAlertGroups(channelIds []string) (bool, error) {
	if len(channelIds) == 0 {
		return true, nil
	}
	need := false
	hasValid := false
	for _, idStr := range channelIds {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		ch, err := app.repository.AlertChannelRepository.GetById(id)
		if err != nil {
			return true, err
		}
		if ch == nil {
			continue
		}
		hasValid = true
		if ch.Type != entity.AlertChannelTypeWebhook {
			need = true
			break
		}
	}
	// 通道 id 全部无效时保守要求分组
	if !hasValid {
		return true, nil
	}
	return need, nil
}

// resolveRelatedMetrics 把关联任务 ID（逗号分隔字符串）解析为 Grafana 面板取数信息。
// 关联任务的实时/样本曲线会叠加到主任务面板；无效 ID 静默跳过，不阻断主流程。
func (app *MonitorTaskApplication) resolveRelatedMetrics(ctx context.Context, relatedIds string) []support.RelatedMetric {
	if strings.TrimSpace(relatedIds) == "" {
		return nil
	}
	ids := make([]int64, 0)
	for _, idStr := range strings.Split(relatedIds, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	tasks, err := app.repository.MonitorTaskRepository.SelectByIds(ids)
	if err != nil {
		logger.WarnFormat(ctx, "resolveRelatedMetrics: 查询关联任务失败 ids=%v: %v", ids, err)
		return nil
	}
	metrics := make([]support.RelatedMetric, 0, len(tasks))
	for _, t := range tasks {
		metrics = append(metrics, support.RelatedMetric{
			TaskKey:  t.TaskKey,
			TaskName: t.TaskName,
			Sampled:  t.Sampled == entity.MonitorSampledStatusOpen,
		})
	}
	return metrics
}

// Create 创建任务：校验 TaskKey 唯一 + Grafana 加 panel + 事务保存
func (app *MonitorTaskApplication) Create(ctx context.Context, req *types.MonitorTaskCreateRequest) error {
	if err := app.validateTaskInput(ctx, req); err != nil {
		return err
	}

	exist, err := app.repository.MonitorTaskRepository.SelectByTaskKey(req.TaskKey)
	if err != nil {
		return err
	}
	if exist != nil {
		return errors.New("任务key已存在")
	}

	taskId := app.sequence.Generate().Int64()
	now := &common.LocalTime{Time: time.Now()}

	task := req.MonitorTask
	task.Id = taskId
	task.PreExecuteTime = now
	task.PreSampleTime = now
	if err := app.marshalExecParamsAndNormalizeTask(&task, &req.TaskExecParams); err != nil {
		return err
	}

	dashboardIds, err := req.GetDashboardIds()
	if err != nil {
		return err
	}
	dashboardTasks := make([]entity.MonitorDashboardTask, 0, len(dashboardIds))
	for i, dashboardId := range dashboardIds {
		dashboardTasks = append(dashboardTasks, entity.MonitorDashboardTask{
			BaseEntity:  common.BaseEntity{Id: app.sequence.Generate().Int64()},
			TaskId:      taskId,
			DashboardId: dashboardId,
			Sort:        int32(i + 1),
		})
	}

	taskAlert, err := app.buildAlertForTask(req, taskId, task.DataType, true)
	if err != nil {
		return err
	}
	if taskAlert == nil {
		task.AlertStatus = entity.MonitorAlertStatusClose
	}

	if app.grafanaHandler != nil && len(dashboardIds) > 0 {
		// 聚合任务的数据写入 taskKey_valueCol（多维度多 metric），不适合单一 realtime panel，跳过同步
		if task.DataType != entity.DataTypeAggregate {
			dashboards, err := app.repository.MonitorDashboardRepository.SelectByIds(dashboardIds)
			if err != nil {
				return err
			}
			sampled := task.Sampled == entity.MonitorSampledStatusOpen
			related := app.resolveRelatedMetrics(ctx, task.RelatedTaskIds)
			for _, d := range dashboards {
				if strings.TrimSpace(d.Uid) == "" {
					return fmt.Errorf("大盘[%s]缺少 Grafana uid，无法同步 panel", d.Name)
				}
				if err := app.grafanaHandler.AddPanel(d.Uid, task.TaskKey, task.TaskName, sampled, related); err != nil {
					logger.Error(ctx, "Grafana AddPanel failed", err)
					return fmt.Errorf("同步 Grafana panel 失败(dashboard=%s, taskKey=%s): %w", d.Name, task.TaskKey, err)
				}
			}
		}
	}

	return app.repository.MonitorTaskRepository.Save(&task, dashboardTasks, taskAlert)
}

// Modify 修改任务
func (app *MonitorTaskApplication) Modify(ctx context.Context, req *types.MonitorTaskCreateRequest) error {
	if req.Id == 0 {
		return errors.New("id 不能为空")
	}
	if err := app.validateTaskInput(ctx, req); err != nil {
		return err
	}
	old, err := app.repository.MonitorTaskRepository.GetById(req.Id)
	if err != nil {
		return err
	}
	if old == nil {
		return errors.New("任务不存在")
	}

	// 聚合任务被下钻依赖时，只允许修改任务名称，其余字段保持原样，不碰 dashboard/alert/Grafana
	if old.DataType == entity.DataTypeAggregate {
		n, err := app.repository.MonitorTaskRepository.CountDrilldownBySourceTaskId(req.Id, false)
		if err != nil {
			return err
		}
		if n > 0 {
			return app.repository.MonitorTaskRepository.UpdateById(req.Id, &entity.MonitorTask{
				TaskName: req.MonitorTask.TaskName,
			})
		}
	}

	// 聚合 → 非聚合：若仍有下钻依赖，禁止改类型，避免下钻悬空
	if old.DataType == entity.DataTypeAggregate && req.DataType != entity.DataTypeAggregate {
		if err := app.ensureNoDrilldownDependents(ctx, req.Id); err != nil {
			return err
		}
	}

	task := req.MonitorTask
	if err := app.marshalExecParamsAndNormalizeTask(&task, &req.TaskExecParams); err != nil {
		return err
	}

	oldDashboards, err := app.repository.MonitorDashboardTaskRepository.FindByTaskId(req.Id)
	if err != nil {
		return err
	}
	oldSet := make(map[int64]bool)
	for _, item := range oldDashboards {
		oldSet[item.DashboardId] = true
	}

	dashboardIds, err := req.GetDashboardIds()
	if err != nil {
		return err
	}
	newSet := make(map[int64]bool)
	dashboardTasks := make([]entity.MonitorDashboardTask, 0, len(dashboardIds))
	for i, dashboardId := range dashboardIds {
		newSet[dashboardId] = true
		dashboardTasks = append(dashboardTasks, entity.MonitorDashboardTask{
			BaseEntity:  common.BaseEntity{Id: app.sequence.Generate().Int64()},
			TaskId:      req.Id,
			DashboardId: dashboardId,
			Sort:        int32(i + 1),
		})
	}

	if err := app.syncGrafanaPanels(ctx, &task, old.Sampled, oldSet, newSet); err != nil {
		return err
	}

	taskAlert, err := app.buildAlertForTask(req, req.Id, task.DataType, false)
	if err != nil {
		return err
	}
	if taskAlert == nil {
		task.AlertStatus = entity.MonitorAlertStatusClose
	}

	return app.repository.MonitorTaskRepository.UpdateTaskAndDashboardTaskAndAlertById(req.Id, &task, dashboardTasks, taskAlert)
}

// marshalExecParamsAndNormalizeTask 将 ExecParams JSON 序列化写入 task，
// 同时完成 dataType 兜底与聚合任务的强制字段清空。Create 与 Modify 共用。
func (app *MonitorTaskApplication) marshalExecParamsAndNormalizeTask(task *entity.MonitorTask, execParams *types.MonitorTaskExecParams) error {
	execBytes, err := json.Marshal(execParams)
	if err != nil {
		return err
	}
	task.ExecParams = string(execBytes)
	// dataType 兜底：仅聚合任务需显式置为 Aggregate(2)；其余归 Normal(1)
	if task.DataType != entity.DataTypeAggregate {
		task.DataType = entity.DataTypeNormal
	} else {
		// 聚合任务只收集不告警不采样，强制关闭；且不支持关联任务/监控分组
		task.AlertStatus = entity.MonitorAlertStatusClose
		task.Sampled = entity.MonitorSampledStatusClose
		task.RelatedTaskIds = ""
		task.MonitorGroup = ""
		task.PromoSensitive = entity.PromoSensitiveOff
	}
	return nil
}

// syncGrafanaPanels 根据编辑前后面板集合的差集，分别对 Grafana 面板做 新增/更新/删除。
// 仅在 grafanaHandler 非 nil 且非聚合任务时执行，聚合任务无单一实时 panel，跳过同步。
func (app *MonitorTaskApplication) syncGrafanaPanels(
	ctx context.Context,
	task *entity.MonitorTask,
	oldSampled entity.MonitorSampledStatus,
	oldSet, newSet map[int64]bool,
) error {
	if app.grafanaHandler == nil {
		return nil
	}
	// 聚合任务的数据写入 taskKey_valueCol（多维度多 metric），不适合单一 realtime panel，跳过同步
	if task.DataType == entity.DataTypeAggregate {
		return nil
	}

	allIds := make([]int64, 0)
	for id := range oldSet {
		allIds = append(allIds, id)
	}
	for id := range newSet {
		if !oldSet[id] {
			allIds = append(allIds, id)
		}
	}
	dashboards, err := app.repository.MonitorDashboardRepository.SelectByIds(allIds)
	if err != nil {
		return err
	}
	dashboardMap := make(map[int64]entity.MonitorDashboard)
	for _, d := range dashboards {
		dashboardMap[d.Id] = d
	}
	// 编辑表单不管理 sampled，请求体里 Sampled 恒为 0；
	// Grafana 同步须以数据库旧值为准，否则保存任务会把面板的样本展示误关
	sampled := oldSampled == entity.MonitorSampledStatusOpen
	related := app.resolveRelatedMetrics(ctx, task.RelatedTaskIds)

	// 处理旧面板：仍在面板集合内则更新，已移除则删除
	for id := range oldSet {
		d, ok := dashboardMap[id]
		if !ok {
			continue
		}
		if strings.TrimSpace(d.Uid) == "" {
			return fmt.Errorf("大盘[%s]缺少 Grafana uid，无法同步 panel", d.Name)
		}
		if newSet[id] {
			if err := app.grafanaHandler.ModifyDashBoardPanel(d.Uid, task.TaskKey, task.TaskName, sampled, related, false, false, true); err != nil {
				logger.Error(ctx, "Grafana ModifyDashBoardPanel(update) failed", err)
				return fmt.Errorf("同步 Grafana panel 失败(dashboard=%s, taskKey=%s): %w", d.Name, task.TaskKey, err)
			}
		} else {
			if err := app.grafanaHandler.ModifyDashBoardPanel(d.Uid, task.TaskKey, task.TaskName, sampled, related, false, true, false); err != nil {
				logger.Error(ctx, "Grafana ModifyDashBoardPanel(delete) failed", err)
				return fmt.Errorf("删除 Grafana panel 失败(dashboard=%s, taskKey=%s): %w", d.Name, task.TaskKey, err)
			}
		}
	}
	// 处理新增面板
	for id := range newSet {
		if oldSet[id] {
			continue
		}
		d, ok := dashboardMap[id]
		if !ok {
			continue
		}
		if strings.TrimSpace(d.Uid) == "" {
			return fmt.Errorf("大盘[%s]缺少 Grafana uid，无法同步 panel", d.Name)
		}
		if err := app.grafanaHandler.ModifyDashBoardPanel(d.Uid, task.TaskKey, task.TaskName, sampled, related, true, false, false); err != nil {
			logger.Error(ctx, "Grafana ModifyDashBoardPanel(add) failed", err)
			return fmt.Errorf("同步 Grafana panel 失败(dashboard=%s, taskKey=%s): %w", d.Name, task.TaskKey, err)
		}
	}
	return nil
}

// validateTaskInput 创建/修改共用的前置校验：
// 结构校验 + 下钻引用 + 报警配置可选对称校验 + Webhook 清分组。
func (app *MonitorTaskApplication) validateTaskInput(ctx context.Context, req *types.MonitorTaskCreateRequest) error {
	if err := req.ValidateForCreate(); err != nil {
		return err
	}
	if err := app.validateDrilldownSource(ctx, req); err != nil {
		return err
	}
	needGroups, err := app.needAlertGroups(req.TaskAlert.AlertChannels)
	if err != nil {
		return err
	}
	if err := req.TaskAlert.ValidateOptional(needGroups); err != nil {
		return err
	}
	if !needGroups {
		req.TaskAlert.AlertGroups = nil
	}
	return nil
}

// buildAlertForTask 为任务构建/补全 Alarm 记录，返回 alert 或 nil。
// isCreate=true 时所有字段强制赋值；=false 时仅补齐零值字段。
// 聚合任务或未填写告警配置时返回 nil，由调用方关闭 task.AlertStatus。
func (app *MonitorTaskApplication) buildAlertForTask(req *types.MonitorTaskCreateRequest, taskId int64, dataType entity.MonitorTaskDataType, isCreate bool) (*entity.MonitorTaskAlert, error) {
	if dataType == entity.DataTypeAggregate || !req.TaskAlert.HasAlertConfig() {
		return nil, nil
	}
	alert := common.Ptr(req.TaskAlert.MonitorTaskAlert)
	alert.TaskId = taskId
	if alert.Id == 0 {
		alert.Id = app.sequence.Generate().Int64()
	}
	now := &common.LocalTime{Time: time.Now()}
	if isCreate || alert.AlertStatus == 0 {
		alert.AlertStatus = entity.MonitorTaskAlertStatusNormal
	}
	if isCreate || alert.DealStatus == 0 {
		alert.DealStatus = entity.MonitorTaskAlertDealStatusNormal
	}
	if isCreate || alert.FirstFlagTime == nil {
		alert.FirstFlagTime = now
	}
	if isCreate || alert.PreCheckTime == nil {
		alert.PreCheckTime = now
	}
	if len(req.TaskAlert.CheckParams) > 0 {
		paramsBytes, err := json.Marshal(req.TaskAlert.CheckParams)
		if err != nil {
			return nil, err
		}
		alert.Params = string(paramsBytes)
	}
	// 始终覆盖通道/分组；纯 Webhook 时 AlertGroups 为空，可清掉历史分组
	alert.AlertChannels = strings.Join(req.TaskAlert.AlertChannels, ",")
	alert.AlertGroups = strings.Join(req.TaskAlert.AlertGroups, ",")
	return alert, nil
}

// ModifyAlertStatus 修改告警开关
// 开启时必须已有告警配置（t_monitor_task_alert 存在且含检测规则/通道等）
func (app *MonitorTaskApplication) ModifyAlertStatus(ctx context.Context, id int64, status entity.MonitorAlertStatus) error {
	task, err := app.repository.MonitorTaskRepository.GetById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	// 聚合任务只收集不告警，禁止开启告警开关
	if task.DataType == entity.DataTypeAggregate && status == entity.MonitorAlertStatusOpen {
		return errors.New("聚合任务不支持告警，无法开启告警开关")
	}
	if status == entity.MonitorAlertStatusOpen {
		alert, err := app.repository.MonitorTaskAlertRepository.GetByTaskId(id)
		if err != nil {
			return err
		}
		if alert == nil {
			return errors.New("该任务尚未配置告警规则，请先编辑任务补充告警配置后再开启")
		}
		// 配置需基本可用：有检测参数 + 检查周期/持续时长
		if strings.TrimSpace(alert.Params) == "" || alert.Params == "[]" || alert.Params == "null" {
			return errors.New("该任务告警配置不完整（缺少异常检测规则），请先编辑任务补充后再开启")
		}
		if alert.TimeSpan <= 0 || alert.Duration <= 0 {
			return errors.New("该任务告警配置不完整（检查间隔/持续时间无效），请先编辑任务补充后再开启")
		}
		if strings.TrimSpace(alert.AlertChannels) == "" {
			return errors.New("该任务告警配置不完整（缺少报警通道），请先编辑任务补充后再开启")
		}
	}
	return app.repository.MonitorTaskRepository.UpdateAlertStatusById(id, status)
}

// ModifyTaskStatus 修改任务开关
// 关闭聚合任务前检查是否仍有开启中的下钻依赖：关停后源 metric 不再更新，下钻会持续空值。
// 开启下钻任务前检查源聚合任务必须已开启。
func (app *MonitorTaskApplication) ModifyTaskStatus(ctx context.Context, id int64, status entity.MonitorTaskStatus) error {
	task, err := app.repository.MonitorTaskRepository.GetById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	if status == entity.MonitorTaskStatusClose {
		if task.DataType == entity.DataTypeAggregate {
			n, err := app.repository.MonitorTaskRepository.CountDrilldownBySourceTaskId(id, true)
			if err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("无法关闭聚合任务：仍有 %d 个下钻任务处于开启状态，请先关闭下钻任务后再操作", n)
			}
		}
	}

	// 下钻任务开启前：源聚合任务必须处于开启状态
	if status == entity.MonitorTaskStatusOpen && task.TaskType != nil && *task.TaskType == entity.TaskTypeDrilldown {
		if err := app.verifyDrilldownSourceOpen(task, id); err != nil {
			return fmt.Errorf("无法开启下钻任务：%w", err)
		}
	}

	// 开启任务时重置上次采集时间与上次样本时间，避免冷启动
	if status == entity.MonitorTaskStatusOpen {
		now := time.Now()
		return app.repository.MonitorTaskRepository.UpdateById(id, &entity.MonitorTask{
			TaskStatus:     status,
			PreExecuteTime: &common.LocalTime{Time: now},
			PreSampleTime:  &common.LocalTime{Time: now},
		})
	}

	return app.repository.MonitorTaskRepository.UpdateTaskStatusById(id, status)
}

// verifyDrilldownSourceOpen 下钻任务开启时验证源聚合任务必须处于开启状态。
func (app *MonitorTaskApplication) verifyDrilldownSourceOpen(task *entity.MonitorTask, drilldownId int64) error {
	var execParams types.MonitorTaskExecParams
	if task.ExecParams != "" {
		if err := json.Unmarshal([]byte(task.ExecParams), &execParams); err != nil {
			return fmt.Errorf("解析下钻任务 %d 的 execParams 失败: %w", drilldownId, err)
		}
	}
	if execParams.SourceTaskId == nil || *execParams.SourceTaskId == 0 {
		return errors.New("下钻任务未配置 sourceTaskId，无法开启")
	}
	src, err := app.repository.MonitorTaskRepository.GetById(*execParams.SourceTaskId)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("下钻依赖的聚合任务 %d 不存在或已删除", *execParams.SourceTaskId)
	}
	if src.TaskStatus != entity.MonitorTaskStatusOpen {
		return fmt.Errorf("下钻依赖的聚合任务「%s」未开启，请先开启聚合任务后再开启下钻任务", src.TaskName)
	}
	return nil
}

// Delete 软删除任务，同步清理关联告警规则。
// 聚合任务若仍被下钻引用则拒绝，避免源消失后下钻静默取空。
func (app *MonitorTaskApplication) Delete(ctx context.Context, id int64) error {
	task, err := app.repository.MonitorTaskRepository.GetById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	if task.DataType == entity.DataTypeAggregate {
		if err := app.ensureNoDrilldownDependents(ctx, id); err != nil {
			return fmt.Errorf("无法删除聚合任务：%w", err)
		}
	}

	// 同步软删除关联告警规则，避免产生孤儿记录
	alert, _ := app.repository.MonitorTaskAlertRepository.GetByTaskId(id)
	if alert != nil {
		if err := app.repository.MonitorTaskAlertRepository.SoftDeleteAlert(alert.Id); err != nil {
			logger.ErrorFormat(ctx, "删除任务时清理告警规则失败 taskId=%d alertId=%d: %v", id, alert.Id, err)
		}
	}

	return app.repository.MonitorTaskRepository.Delete(id)
}

// ModifySampled 修改样本展示开关
func (app *MonitorTaskApplication) ModifySampled(ctx context.Context, id int64, status entity.MonitorSampledStatus) error {
	task, err := app.repository.MonitorTaskRepository.GetById(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	// 聚合任务不生成样本面板，切换样本开关无意义，直接返回
	if task.DataType == entity.DataTypeAggregate {
		return nil
	}
	if err := app.repository.MonitorTaskRepository.UpdateSampledById(id, status); err != nil {
		return err
	}
	if app.grafanaHandler != nil {
		dashboards, err := app.repository.MonitorDashboardTaskRepository.FindMonitorDashBoardsByTaskId(id)
		if err != nil {
			return err
		}
		sampled := status == entity.MonitorSampledStatusOpen
		// 关联任务 ID 为任务一级字段，切换样本开关时需一并解析以重建面板
		related := app.resolveRelatedMetrics(ctx, task.RelatedTaskIds)
		for _, d := range dashboards {
			if strings.TrimSpace(d.Uid) == "" {
				return fmt.Errorf("大盘[%s]缺少 Grafana uid，无法同步 panel", d.Name)
			}
			if err := app.grafanaHandler.ModifyDashBoardPanel(d.Uid, task.TaskKey, task.TaskName, sampled, related, false, false, true); err != nil {
				logger.Error(ctx, "Grafana ModifyDashBoardPanel(sampled) failed", err)
				return fmt.Errorf("同步 Grafana panel 失败(dashboard=%s, taskKey=%s): %w", d.Name, task.TaskKey, err)
			}
		}
	}
	return nil
}

// Count 统计
func (app *MonitorTaskApplication) Count(ctx context.Context) (int64, error) {
	c, err := app.repository.MonitorTaskRepository.Count()
	if err != nil || c == nil {
		return 0, err
	}
	return *c, nil
}

// validateDrilldownSource 校验下钻任务的源聚合引用：
// sourceTaskId 必须存在、未删除，且 DataType=Aggregate；
// 若指定 queryMetric，须落在源任务 valueColumns 中。
func (app *MonitorTaskApplication) validateDrilldownSource(ctx context.Context, req *types.MonitorTaskCreateRequest) error {
	if req.TaskType == nil || *req.TaskType != entity.TaskTypeDrilldown {
		return nil
	}
	srcId := req.TaskExecParams.SourceTaskId
	if srcId == nil || *srcId == 0 {
		// ValidateTypedParams 已拦，这里兜底
		return errors.New("下钻任务必须指定 sourceTaskId")
	}
	// 禁止自引用（修改场景下 id 已有值）
	if req.Id != 0 && req.Id == *srcId {
		return errors.New("下钻任务不能依赖自身")
	}
	src, err := app.repository.MonitorTaskRepository.GetById(*srcId)
	if err != nil {
		return err
	}
	if src == nil {
		return fmt.Errorf("下钻依赖的聚合任务 %d 不存在或已删除", *srcId)
	}
	if src.DataType != entity.DataTypeAggregate {
		return fmt.Errorf("下钻依赖的任务「%s」不是聚合任务，无法作为源", src.TaskName)
	}

	// queryMetric 可选；指定时校验落在源 valueColumns
	qm := strings.TrimSpace(req.TaskExecParams.QueryMetric)
	if qm == "" {
		return nil
	}
	var srcParams types.MonitorTaskExecParams
	if src.ExecParams != "" {
		if err := json.Unmarshal([]byte(src.ExecParams), &srcParams); err != nil {
			return fmt.Errorf("解析源聚合任务 execParams 失败: %w", err)
		}
	}
	for _, col := range srcParams.ValueColumns {
		if col == qm {
			return nil
		}
	}
	return fmt.Errorf("queryMetric=%s 不在源聚合任务「%s」的 valueColumns 中", qm, src.TaskName)
}

// ensureNoDrilldownDependents 确认没有下钻依赖指定聚合任务。
// 用于：改类型 / 关停 / 删除 聚合任务前的统一反查。
func (app *MonitorTaskApplication) ensureNoDrilldownDependents(ctx context.Context, sourceTaskId int64) error {
	n, err := app.repository.MonitorTaskRepository.CountDrilldownBySourceTaskId(sourceTaskId, false)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("仍有 %d 个下钻任务依赖此聚合任务，请先删除或改挂下钻任务后再操作", n)
	}
	return nil
}

// PreviewAggregate 聚合预览：临时执行多行查询，返回结果列名，供前端勾选 label/value 维度。
// 不落库、不影响采集；复用 CommandHandler.ExecuteMultiRows 透传当前填写的数据源/SQL/URL。
func (app *MonitorTaskApplication) PreviewAggregate(ctx context.Context, req *types.MonitorTaskPreviewRequest) (*types.MonitorTaskPreviewResponse, error) {
	var taskType int32
	if req.TaskType != nil {
		taskType = int32(*req.TaskType)
	}
	cmd, ok := app.commonMap.GetCommandHandler(ctx, taskType)
	if !ok {
		return nil, fmt.Errorf("未找到任务类型 %d 对应的数据源处理器", taskType)
	}
	execBytes, err := json.Marshal(req.ExecParams)
	if err != nil {
		return nil, err
	}
	// 构造临时任务：仅注入多行取数所需字段（Command + ExecParams）
	tmp := entity.MonitorTask{
		Command:    req.Command,
		ExecParams: string(execBytes),
	}
	// 预览查询加超时保护，防止慢 SQL 长时间占住连接（D-003）
	previewCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	rows, err := cmd.ExecuteMultiRows(previewCtx, tmp)
	if err != nil {
		return nil, fmt.Errorf("预览查询失败: %w", err)
	}
	if len(rows) == 0 {
		return &types.MonitorTaskPreviewResponse{Columns: []string{}}, nil
	}
	// 取首行列名；不同行列名应一致，取并集
	colSet := make(map[string]bool)
	columns := make([]string, 0)
	for _, row := range rows {
		for col := range row.Columns {
			if !colSet[col] {
				colSet[col] = true
				columns = append(columns, col)
			}
		}
	}

	// 聚合查询要求每列都有明确别名：无别名聚合函数（COUNT(*)/SUM(x) 等）或空列名
	// 会导致 tag/metric 命名错乱，提前拦截并提示用户用 AS 取别名
	if badCols := common.FindUnnamedAggColumns(columns); len(badCols) > 0 {
		return nil, fmt.Errorf(
			"预览列名不规范：%s 缺少别名，聚合函数/表达式列必须用 AS 取别名（如 COUNT(*) AS cnt）",
			strings.Join(badCols, "、"),
		)
	}

	return &types.MonitorTaskPreviewResponse{Columns: columns}, nil
}
