package persistence

import (
	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type MonitorVolatilityDayRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorVolatilityDayRepositoryImpl(db *gorm.DB) *MonitorVolatilityDayRepositoryImpl {
	return &MonitorVolatilityDayRepositoryImpl{db: db}
}

// SelectAll 查询全部（未删除）
func (repo *MonitorVolatilityDayRepositoryImpl) SelectAll() ([]entity.MonitorVolatilityDay, error) {
	var data []entity.MonitorVolatilityDay
	err := repo.db.Model(&entity.MonitorVolatilityDay{}).
		Not(&entity.MonitorVolatilityDay{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("start_time").
		Find(&data).Error
	return data, err
}

// Save 保存单条
func (repo *MonitorVolatilityDayRepositoryImpl) Save(day *entity.MonitorVolatilityDay) error {
	return repo.db.Model(&entity.MonitorVolatilityDay{}).Create(day).Error
}

// BatchSave 批量保存
func (repo *MonitorVolatilityDayRepositoryImpl) BatchSave(days []entity.MonitorVolatilityDay) error {
	return repo.db.Model(&entity.MonitorVolatilityDay{}).Create(&days).Error
}

// Modify 按主键更新
func (repo *MonitorVolatilityDayRepositoryImpl) Modify(id int64, day *entity.MonitorVolatilityDay) error {
	return repo.db.Model(&entity.MonitorVolatilityDay{}).
		Where(&entity.MonitorVolatilityDay{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(day).Error
}

// Delete 软删除
func (repo *MonitorVolatilityDayRepositoryImpl) Delete(id int64) error {
	return repo.db.Model(&entity.MonitorVolatilityDay{}).
		Where("id = ?", id).
		Updates(&entity.MonitorVolatilityDay{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
}
