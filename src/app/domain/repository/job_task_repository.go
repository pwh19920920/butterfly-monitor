package repository

import "butterfly-monitor/src/app/domain/entity"

type JobTaskRepository interface {

	// FindJobBySharding 取余分页查询
	FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.JobTask, error)
}
