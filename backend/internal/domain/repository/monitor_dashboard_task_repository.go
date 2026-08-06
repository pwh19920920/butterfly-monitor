package repository

import "dragonfly-monitor/internal/domain/entity"

type MonitorDashboardTaskRepository interface {
	FindByDashboardId(dashboardId int64) ([]entity.MonitorDashboardTask, error)
	FindMonitorDashBoardsByTaskId(taskId int64) ([]entity.MonitorDashboard, error)
	FindByTaskId(taskId int64) ([]entity.MonitorDashboardTask, error)
	GetById(id int64) (*entity.MonitorDashboardTask, error)
	BatchModifySort(items []entity.MonitorDashboardTask) error
}
