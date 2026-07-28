package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type MonitorTaskEventRepository interface {
	FindEventJob() ([]entity.MonitorTaskEvent, error)
	FindPendingEventAll() ([]entity.MonitorTaskEvent, error)
	Create(monitorTaskEvent *entity.MonitorTaskEvent) error
	Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error
	ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error
	BatchModifyByEvents(eventIds []int64, monitorTaskEvent *entity.MonitorTaskEvent) error
	SelectByTaskId(taskId int64) ([]entity.MonitorTaskEvent, error)
	Select(req *types.MonitorTaskEventQueryRequest) (int64, []entity.MonitorTaskEvent, error)
	GetById(id int64) (*entity.MonitorTaskEvent, error)
	DealEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error
	CompleteEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error
	IgnoreEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error
	Count() (*int64, error)
	// CountByStatus 按处理状态统计事件数
	CountByStatus() (map[entity.MonitorTaskEventDealStatus]int64, error)
	// CountByLevel 按告警级别统计事件数
	CountByLevel() (map[int32]int64, error)
	// SelectRecent 查询最近 N 条事件（按创建时间倒序）
	SelectRecent(limit int) ([]entity.MonitorTaskEvent, error)
}

