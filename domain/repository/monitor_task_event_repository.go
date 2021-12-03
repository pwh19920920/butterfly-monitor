package repository

import "butterfly-monitor/domain/entity"

type MonitorTaskEventRepository interface {
	// FindEventJob 获取需要执行的job
	FindEventJob() ([]entity.MonitorTaskEvent, error)

	// Create 创建
	Create(monitorTaskEvent *entity.MonitorTaskEvent) error

	// Modify 更新
	Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error

	// ModifyByEvent 批量更新
	ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error
}
