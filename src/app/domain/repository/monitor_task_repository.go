package repository

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/types"
)

type MonitorTaskRepository interface {

	// FindJobBySharding 取余分页查询
	FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error)

	// Save 保存
	Save(monitorTask *entity.MonitorTask) error

	// UpdateById 更新
	UpdateById(id int64, monitorTask *entity.MonitorTask) error

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.MonitorTaskQueryRequest) (int64, []entity.MonitorTask, error)

	// UpdateAlertStatusById 更新报警状态
	UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error

	// UpdateTaskStatusById 更新任务状态
	UpdateTaskStatusById(id int64, status entity.MonitorTaskStatus) error
}
