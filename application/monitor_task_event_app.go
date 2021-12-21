package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/common"
	sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"
	"time"
)

type MonitorTaskEventApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

func (app *MonitorTaskEventApplication) Query(req *types.MonitorTaskEventQueryRequest) (int64, []types.MonitorTaskEventQueryResponse, error) {
	length, data, err := app.repository.MonitorTaskEventRepository.Select(req)
	if err != nil || data == nil || len(data) == 0 {
		return 0, []types.MonitorTaskEventQueryResponse{}, err
	}

	taskIds := make([]int64, 0)
	userIds := make([]int64, 0)
	for _, item := range data {
		taskIds = append(taskIds, item.TaskId)
		if item.DealUser != nil {
			userIds = append(userIds, *item.DealUser)
		}
	}
	taskMap, err := app.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)
	if err != nil {
		return 0, nil, err
	}

	userMap := make(map[int64]sysEntity.SysUser, 0)
	users, err := app.repository.AlertGroupUserRepository.SelectUsersByUserIds(userIds)
	if err != nil {
		return 0, nil, err
	}
	for _, user := range users {
		userMap[user.Id] = user
	}

	// 整理数据
	result := make([]types.MonitorTaskEventQueryResponse, 0)
	for _, item := range data {
		task, taskOk := taskMap[item.TaskId]
		var dealUserName string
		var taskName string
		if taskOk {
			taskName = task.TaskName
		}

		//  处理人信息
		if item.DealUser != nil {
			if user, userOk := userMap[*item.DealUser]; userOk {
				dealUserName = user.Name
			}
		}
		result = append(result, types.MonitorTaskEventQueryResponse{MonitorTaskEvent: item, TaskName: taskName, DealUserName: dealUserName})
	}
	return length, result, err
}

// DealEvent 处理事件
// 事件表：设置处理人，处理事件
// 检查表：设置状态为处理中Deal
func (app *MonitorTaskEventApplication) DealEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	// 查出task下全部待处理的事件
	events, err := app.repository.MonitorTaskEventRepository.SelectByTaskId(req.TaskId)
	if err != nil {
		return err
	}

	if events == nil || len(events) == 0 {
		return errors.New("任务下待处理的事件不存在")
	}

	// 除最后一条外都设置为误报，状态完成
	if len(events) > 1 {
		eventIds := make([]int64, 0)
		for _, item := range events {
			if item.Id != eventId {
				eventIds = append(eventIds, item.Id)
			}
		}

		// 判断是否错误
		if err := app.repository.MonitorTaskEventRepository.BatchModifyByEvents(eventIds, &entity.MonitorTaskEvent{
			DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
			CompleteTime: &common.LocalTime{Time: time.Now()}}); err != nil {
			return err
		}
	}
	return app.repository.MonitorTaskEventRepository.DealEvent(eventId, req)
}

// CompleteEvent 完成事件
// 时间表：更新事件经过，完成时间
// 检查表：更新报警状态，处理状态为Normal
func (app *MonitorTaskEventApplication) CompleteEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return app.repository.MonitorTaskEventRepository.CompleteEvent(eventId, req)
}
