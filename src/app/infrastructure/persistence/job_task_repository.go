package persistence

import (
	"butterfly-monitor/src/app/domain/entity"
	"gorm.io/gorm"
)

type JobTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewJobTaskRepositoryImpl(db *gorm.DB) *JobTaskRepositoryImpl {
	return &JobTaskRepositoryImpl{db: db}
}

func (repo *JobTaskRepositoryImpl) FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.JobTask, error) {
	var data []entity.JobTask
	err := repo.db.
		Model(&entity.JobTask{}).
		Where("id > ? and mod(id, ?) = ? and  date_add(now(), interval -time_space second) > pre_execute_time order by id limit 0, ?", lastId, shardTotal, shardIndex, pageSize).
		Find(&data).Error
	return data, err
}
