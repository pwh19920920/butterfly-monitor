package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type MonitorTaskEventRepository interface {
	// FindEventJob 获取需要执行的job
	FindEventJob() ([]entity.MonitorTaskEvent, error)

	// Create 创建
	Create(monitorTaskEvent *entity.MonitorTaskEvent) error

	// Modify 更新
	Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error

	// ModifyByEvent 批量更新
	ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error

	// BatchModifyByEvents 批量更新
	BatchModifyByEvents(eventIds []int64, monitorTaskEvent *entity.MonitorTaskEvent) error

	// SelectByTaskId 查询
	SelectByTaskId(taskId int64) ([]entity.MonitorTaskEvent, error)

	// Select 查询
	Select(req *types.MonitorTaskEventQueryRequest) (int64, []entity.MonitorTaskEvent, error)

	// DealEvent 事件处理中
	DealEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error

	// CompleteEvent 事件完成
	CompleteEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error
}
