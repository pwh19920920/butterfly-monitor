package repository

import (
	"butterfly-monitor/src/app/domain/entity"
)

type MonitorDashboardTaskRepository interface {
	SelectByTaskIds(taskIds []int64) ([]entity.MonitorDashboardTask, error)
}
