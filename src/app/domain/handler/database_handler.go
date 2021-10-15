package handler

import "butterfly-monitor/src/app/domain/entity"

type DatabaseHandler interface {

	// NewInstance 连接数据源
	NewInstance(database entity.JobDatabase) (interface{}, error)

	// ExecuteQuery 执行查询
	ExecuteQuery(task entity.JobTask) (int64, error)
}
