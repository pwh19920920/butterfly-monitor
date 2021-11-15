package repository

import (
	entity2 "butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type MonitorTaskRepository interface {

	// FindJobBySharding 取余分页查询
	FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity2.MonitorTask, error)

	// Save 保存
	Save(monitorTask *entity2.MonitorTask, dashboardTasks []entity2.MonitorDashboardTask) error

	// UpdateById 更新
	UpdateById(id int64, monitorTask *entity2.MonitorTask, dashboardTasks []entity2.MonitorDashboardTask) error

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.MonitorTaskQueryRequest) (int64, []entity2.MonitorTask, error)

	// UpdateAlertStatusById 更新报警状态
	UpdateAlertStatusById(id int64, status entity2.MonitorAlertStatus) error

	// UpdateTaskStatusById 更新任务状态
	UpdateTaskStatusById(id int64, status entity2.MonitorTaskStatus) error

	// UpdateSampledById 更新收集状态
	UpdateSampledById(id int64, status entity2.MonitorSampledStatus) error

	// GetById 获取数据
	GetById(id int64) (*entity2.MonitorTask, error)

	SelectByIdsWithMap(ids []int64) (map[int64]entity2.MonitorTask, error)

	SelectByIds(ids []int64) ([]entity2.MonitorTask, error)
}
