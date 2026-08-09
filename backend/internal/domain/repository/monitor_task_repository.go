package repository

import (
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
)

type MonitorTaskRepository interface {
	FindJobBySharding(shardIndex, shardTotal int64) ([]entity.MonitorTask, error)
	FindSamplingJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error)
	Save(monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, alert *entity.MonitorTaskAlert) error
	UpdateById(id int64, monitorTask *entity.MonitorTask) error
	UpdateTaskAndDashboardTaskAndAlertById(id int64, monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error
	Delete(id int64) error
	Select(req *types.MonitorTaskQueryRequest) (int64, []entity.MonitorTask, error)
	UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error
	UpdateTaskStatusById(id int64, status entity.MonitorTaskStatus) error
	UpdateSampledById(id int64, status entity.MonitorSampledStatus) error
	GetById(id int64) (*entity.MonitorTask, error)
	SelectByIdsWithMap(ids []int64) (map[int64]entity.MonitorTask, error)
	SelectByIds(ids []int64) ([]entity.MonitorTask, error)
	SelectByTaskKey(taskKey string) (*entity.MonitorTask, error)
	// CountDrilldownBySourceTaskId 统计仍依赖指定聚合任务的下钻任务数（onlyOpen=true 只统计开启的下钻）
	CountDrilldownBySourceTaskId(sourceTaskId int64, onlyOpen bool) (int64, error)
	Count() (*int64, error)
	SelectAll() ([]entity.MonitorTask, error)
}
