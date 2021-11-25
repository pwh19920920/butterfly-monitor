package handler

import (
	"butterfly-monitor/domain/entity"
)

type DatabaseHandler interface {

	// TestConnect 测试连接
	TestConnect(database entity.MonitorDatabase) error

	// NewInstance 连接数据源
	NewInstance(database entity.MonitorDatabase) (interface{}, error)

	// ExecuteQuery 执行查询
	ExecuteQuery(task entity.MonitorTask) (interface{}, error)
}
