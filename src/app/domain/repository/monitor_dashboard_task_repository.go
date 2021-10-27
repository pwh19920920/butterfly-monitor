package repository

import (
	"butterfly-monitor/src/app/domain/entity"
)

type MonitorDashboardTaskRepository interface {
	SelectByTaskIds(taskIds []int64) ([]entity.MonitorDashboardTask, error)
	SelectByIds(ids []int64) ([]entity.MonitorDashboardTask, error)
	SelectByDashboardId(dashboardId int64) ([]entity.MonitorDashboardTask, error)
	BatchModifySort(data []entity.MonitorDashboardTask) error
}
