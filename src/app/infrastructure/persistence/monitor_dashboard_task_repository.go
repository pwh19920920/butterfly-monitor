package persistence

import (
	"butterfly-monitor/src/app/domain/entity"
	"gorm.io/gorm"
)

type MonitorDashboardTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDashboardTaskRepositoryImpl(db *gorm.DB) *MonitorDashboardTaskRepositoryImpl {
	return &MonitorDashboardTaskRepositoryImpl{db: db}
}

func (repo *MonitorDashboardTaskRepositoryImpl) SelectByTaskIds(taskIds []int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	tx := repo.db.Model(&entity.MonitorDashboardTask{})
	err := tx.Where("task_id in ?", taskIds).Find(&data).Error
	return data, err
}
