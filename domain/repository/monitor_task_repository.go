package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type MonitorTaskRepository interface {

	// FindJobBySharding 取余分页查询
	FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error)

	// FindJobByShardingNoPaging 取余分页查询
	FindJobByShardingNoPaging(shardIndex, shardTotal int64) ([]entity.MonitorTask, error)

	// FindSamplingJobBySharding 取余分页查询
	FindSamplingJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error)

	// Save 保存
	Save(monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, alert entity.MonitorTaskAlert) error

	// UpdateById 更新
	UpdateById(id int64, monitorTask *entity.MonitorTask) error

	// UpdateTaskAndDashboardTaskAndAlertById 更新任务以及报警信息
	UpdateTaskAndDashboardTaskAndAlertById(id int64, monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.MonitorTaskQueryRequest) (int64, []entity.MonitorTask, error)

	// UpdateAlertStatusById 更新报警状态
	UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error

	// UpdateTaskStatusById 更新任务状态
	UpdateTaskStatusById(id int64, status entity.MonitorTaskStatus) error

	// UpdateSampledById 更新收集状态
	UpdateSampledById(id int64, status entity.MonitorSampledStatus) error

	// GetById 获取数据
	GetById(id int64) (*entity.MonitorTask, error)

	SelectByIdsWithMap(ids []int64) (map[int64]entity.MonitorTask, error)

	SelectByIds(ids []int64) ([]entity.MonitorTask, error)

	// SelectByTaskKey 通过任务key查询
	SelectByTaskKey(taskKey string) (*entity.MonitorTask, error)

	// Count 统计总数
	Count() (*int64, error)

	// SelectAll 查询全部
	SelectAll() ([]entity.MonitorTask, error)
}
