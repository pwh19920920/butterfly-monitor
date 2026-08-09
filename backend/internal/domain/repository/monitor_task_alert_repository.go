package repository

import (
	"time"

	"butterfly-monitor/internal/domain/entity"
)

type MonitorTaskAlertRepository interface {
	FindCheckJob(shardIndex, shardTotal int64) ([]entity.MonitorTaskAlert, error)
	BatchGetByIds(ids []int64) ([]entity.MonitorTaskAlert, error)
	BatchGetByTaskIds(taskIds []int64) ([]entity.MonitorTaskAlert, error)
	Modify(id int64, monitorAlert *entity.MonitorTaskAlert) error
	ModifyByPending(id int64, currentTime time.Time) error
	ModifyForNormal(id int64, currentTime time.Time) error
	ModifyByFiring(id int64, currentTime time.Time, monitorTaskEvent *entity.MonitorTaskEvent) error
	ModifyByAlert(whereCase *entity.MonitorTaskAlert, monitorTaskAlert *entity.MonitorTaskAlert) error
	GetByTaskId(taskId int64) (*entity.MonitorTaskAlert, error)
	// SoftDeleteAlert 软删除告警规则并忽略关联 Pending 事件，用于任务被删除后清理孤儿规则。
	SoftDeleteAlert(id int64) error
}
