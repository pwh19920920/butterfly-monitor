package persistence

import (
	"butterfly-monitor/src/app/domain/entity"
	"gorm.io/gorm"
)

type MonitorTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskRepositoryImpl(db *gorm.DB) *MonitorTaskRepositoryImpl {
	return &MonitorTaskRepositoryImpl{db: db}
}

func (repo *MonitorTaskRepositoryImpl) FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("id > ? and mod(id, ?) = ? and  date_add(now(), interval -time_space second) > pre_execute_time order by id limit 0, ?", lastId, shardTotal, shardIndex, pageSize).
		Find(&data).Error
	return data, err
}
