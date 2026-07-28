package application

import (
	"context"
	"errors"

	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
)

// MonitorTaskEventApplication 告警事件应用服务
type MonitorTaskEventApplication struct {
	sequence     *snowflake.Node
	repository   *persistence.Repository
	alertConfApp *AlertConfApplication
}

// NewMonitorTaskEventApplication 创建告警事件应用服务
func NewMonitorTaskEventApplication(
	sequence *snowflake.Node,
	repository *persistence.Repository,
	alertConfApp *AlertConfApplication,
) MonitorTaskEventApplication {
	return MonitorTaskEventApplication{
		sequence:     sequence,
		repository:   repository,
		alertConfApp: alertConfApp,
	}
}

// Query 分页查询事件，补 TaskName / DealUserName
func (app *MonitorTaskEventApplication) Query(ctx context.Context, req *types.MonitorTaskEventQueryRequest) (int64, []types.MonitorTaskEventQueryResponse, error) {
	total, data, err := app.repository.MonitorTaskEventRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "MonitorTaskEventRepository.Select() happen error for", err)
		return total, nil, err
	}

	// 收集 taskId / dealUser
	taskIds := make([]int64, 0)
	taskIdSet := make(map[int64]bool)
	userIds := make([]int64, 0)
	userIdSet := make(map[int64]bool)
	for _, item := range data {
		if !taskIdSet[item.TaskId] {
			taskIdSet[item.TaskId] = true
			taskIds = append(taskIds, item.TaskId)
		}
		if item.DealUser != nil && !userIdSet[*item.DealUser] {
			userIdSet[*item.DealUser] = true
			userIds = append(userIds, *item.DealUser)
		}
	}

	taskMap, err := app.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)
	if err != nil {
		return total, nil, err
	}

	userMap := make(map[int64]entity.SysUser)
	if len(userIds) > 0 {
		users, err := app.repository.AlertGroupUserRepository.SelectUsersByUserIds(userIds)
		if err != nil {
			return total, nil, err
		}
		for _, u := range users {
			userMap[u.Id] = u
		}
	}

	result := make([]types.MonitorTaskEventQueryResponse, 0, len(data))
	for _, item := range data {
		resp := types.MonitorTaskEventQueryResponse{MonitorTaskEvent: item}
		if t, ok := taskMap[item.TaskId]; ok {
			resp.TaskName = t.TaskName
		}
		if item.DealUser != nil {
			if u, ok := userMap[*item.DealUser]; ok {
				resp.DealUserName = u.Name
			}
		}
		result = append(result, resp)
	}
	return total, result, nil
}

// DealEvent 处理事件
func (app *MonitorTaskEventApplication) DealEvent(ctx context.Context, eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	if req.TaskId == 0 {
		// 从事件反查
		event, err := app.repository.MonitorTaskEventRepository.GetById(eventId)
		if err != nil {
			return err
		}
		if event == nil {
			return errors.New("事件不存在")
		}
		req.TaskId = event.TaskId
	}
	return app.repository.MonitorTaskEventRepository.DealEvent(eventId, req)
}

// CompleteEvent 完成事件
func (app *MonitorTaskEventApplication) CompleteEvent(ctx context.Context, eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	event, err := app.repository.MonitorTaskEventRepository.GetById(eventId)
	if err != nil {
		return err
	}
	if event == nil {
		return errors.New("事件不存在")
	}
	if req.TaskId == 0 {
		req.TaskId = event.TaskId
	}
	return app.repository.MonitorTaskEventRepository.CompleteEvent(eventId, req)
}

// IgnoreEvent 忽略事件：仅置当前事件为 Ignore；该 alert 无剩余 Pending 时恢复 alert 正常
func (app *MonitorTaskEventApplication) IgnoreEvent(ctx context.Context, eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	event, err := app.repository.MonitorTaskEventRepository.GetById(eventId)
	if err != nil {
		return err
	}
	if event == nil {
		return errors.New("事件不存在")
	}
	// 仅待处理事件可忽略
	if event.DealStatus != entity.MonitorTaskEventDealStatusPending {
		return errors.New("仅待处理事件可忽略")
	}
	if req.TaskId == 0 {
		req.TaskId = event.TaskId
	}
	if req.AlertId == 0 {
		req.AlertId = event.AlertId
	}
	return app.repository.MonitorTaskEventRepository.IgnoreEvent(eventId, req)
}

// Count 统计
func (app *MonitorTaskEventApplication) Count(ctx context.Context) (int64, error) {
	c, err := app.repository.MonitorTaskEventRepository.Count()
	if err != nil || c == nil {
		return 0, err
	}
	return *c, nil
}

// CountByStatus 按处理状态统计事件数
func (app *MonitorTaskEventApplication) CountByStatus(ctx context.Context) (map[entity.MonitorTaskEventDealStatus]int64, error) {
	return app.repository.MonitorTaskEventRepository.CountByStatus()
}

// CountByLevel 按告警级别统计事件数
func (app *MonitorTaskEventApplication) CountByLevel(ctx context.Context) (map[int32]int64, error) {
	return app.repository.MonitorTaskEventRepository.CountByLevel()
}

// SelectRecent 查询最近 N 条事件
func (app *MonitorTaskEventApplication) SelectRecent(limit int) ([]entity.MonitorTaskEvent, error) {
	return app.repository.MonitorTaskEventRepository.SelectRecent(limit)
}

// HomeCount 首页统计
func (app *MonitorTaskEventApplication) HomeCount(ctx context.Context) (*types.MonitorHomeCountResponse, error) {
	// 由 all_app 中的聚合方法实现，这里保留占位
	return nil, nil
}
