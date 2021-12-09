package repository

import (
	"butterfly-monitor/domain/entity"
	"time"
)

type MonitorTaskAlertRepository interface {
	// FindCheckJob 获取需要执行的job
	FindCheckJob(shardIndex, shardTotal int64) ([]entity.MonitorTaskAlert, error)

	// BatchGetByIds 批量获取
	BatchGetByIds(ids []int64) ([]entity.MonitorTaskAlert, error)

	// BatchGetByTaskIds 批量获取
	BatchGetByTaskIds(taskIds []int64) ([]entity.MonitorTaskAlert, error)

	// Modify 更新
	Modify(id int64, monitorAlert *entity.MonitorTaskAlert) error

	// ModifyByPending 等待报警
	ModifyByPending(id int64, currentTime time.Time) error

	// ModifyForNormal 恢复正常
	ModifyForNormal(id int64, currentTime time.Time) error

	// ModifyByFiring 开始报警
	ModifyByFiring(id int64, currentTime time.Time, monitorTaskEvent *entity.MonitorTaskEvent) error

	// ModifyByAlert 更新
	ModifyByAlert(whereCase *entity.MonitorTaskAlert, monitorTaskAlert *entity.MonitorTaskAlert) error
}
