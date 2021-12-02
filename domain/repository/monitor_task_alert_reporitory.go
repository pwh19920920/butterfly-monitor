package repository

import (
	"butterfly-monitor/domain/entity"
)

type MonitorTaskAlertRepository interface {
	// FindCheckJob 获取需要执行的job
	FindCheckJob(shardIndex, shardTotal int64) ([]entity.MonitorTaskAlert, error)

	// Modify 更新
	Modify(id int64, monitorAlert *entity.MonitorTaskAlert) error
}
