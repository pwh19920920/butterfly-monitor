package handler

import (
	entity2 "butterfly-monitor/domain/entity"
)

type DatabaseHandler interface {

	// NewInstance 连接数据源
	NewInstance(database entity2.MonitorDatabase) (interface{}, error)

	// ExecuteQuery 执行查询
	ExecuteQuery(task entity2.MonitorTask) (interface{}, error)
}
