package persistence

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
)

type MonitorTaskEventRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskEventRepositoryImpl(db *gorm.DB) *MonitorTaskEventRepositoryImpl {
	return &MonitorTaskEventRepositoryImpl{db: db}
}

// FindEventJob 获取需要执行的job
func (repo *MonitorTaskEventRepositoryImpl) FindEventJob() ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("deal_status = ? and now() >= next_alert_time", entity.MonitorTaskEventDealStatusPending).
		Find(&data).Error
	return data, err
}

// Create 创建
func (repo *MonitorTaskEventRepositoryImpl) Create(monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).Create(&monitorTaskEvent).Error
}

// Modify 更新
func (repo *MonitorTaskEventRepositoryImpl) Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTaskEvent).Error
}

// ModifyByEvent 更新
func (repo *MonitorTaskEventRepositoryImpl) ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(whereCase).
		Updates(&monitorTaskEvent).Error
}

func (repo *MonitorTaskEventRepositoryImpl) BatchModifyByEvents(eventIds []int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where("id in (?)", eventIds).
		Updates(&monitorTaskEvent).Error
}
