package persistence

import (
	"errors"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorDatabaseRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDatabaseRepositoryImpl(db *gorm.DB) *MonitorDatabaseRepositoryImpl {
	return &MonitorDatabaseRepositoryImpl{db: db}
}

// SelectAll 查询全部数据源，可选按更新时间过滤
func (repo *MonitorDatabaseRepositoryImpl) SelectAll(lastTime *common.LocalTime) ([]entity.MonitorDatabase, error) {
	var data []entity.MonitorDatabase
	tx := repo.db.Model(&entity.MonitorDatabase{})
	if lastTime != nil {
		tx = tx.Where("updated_at >= ?", lastTime.Time)
	}
	err := tx.Order("id desc").Find(&data).Error
	return data, err
}

// SelectSimpleAll 简单查询（不含敏感字段）
func (repo *MonitorDatabaseRepositoryImpl) SelectSimpleAll() ([]entity.MonitorDatabase, error) {
	var data []entity.MonitorDatabase
	err := repo.db.Model(&entity.MonitorDatabase{}).
		Select("id", "name", "database", "type").
		Order("id desc").
		Find(&data).Error
	return data, err
}

// Save 保存
func (repo *MonitorDatabaseRepositoryImpl) Save(monitorDatabase *entity.MonitorDatabase) error {
	return repo.db.Model(&entity.MonitorDatabase{}).Create(&monitorDatabase).Error
}

// UpdateById 按主键更新
func (repo *MonitorDatabaseRepositoryImpl) UpdateById(id int64, jobDatabase *entity.MonitorDatabase) error {
	return repo.db.Model(&entity.MonitorDatabase{}).
		Where(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&jobDatabase).Error
}

// UpdateHealth 仅更新探活字段；用 map 写入，确保 last_error 清空、consecutive_fail=0 等也能落库
func (repo *MonitorDatabaseRepositoryImpl) UpdateHealth(id int64, healthStatus int32, lastCheck *common.LocalTime, lastError string, consecutiveFail int32) error {
	fields := map[string]interface{}{
		"health_status":    healthStatus,
		"last_error":       lastError,
		"consecutive_fail": consecutiveFail,
	}
	if lastCheck != nil {
		fields["last_check_time"] = lastCheck
	}
	return repo.db.Model(&entity.MonitorDatabase{}).
		Where("id = ?", id).
		Not(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Updates(fields).Error
}

// Delete 软删除
func (repo *MonitorDatabaseRepositoryImpl) Delete(id int64) error {
	return repo.db.Model(&entity.MonitorDatabase{}).
		Where(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
}

// Select 分页查询
func (repo *MonitorDatabaseRepositoryImpl) Select(req *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error) {
	var count int64 = 0
	whereCase := &entity.MonitorDatabase{
		Name: req.Name,
		Type: req.Type,
	}
	notCase := &entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	repo.db.Model(&entity.MonitorDatabase{}).Where(whereCase).Not(notCase).Count(&count)

	var data []entity.MonitorDatabase
	// 全量查；password/salt 由 application.Query 清空后再返回前端
	err := repo.db.Model(&entity.MonitorDatabase{}).
		Order("id desc").
		Where(whereCase).
		Not(notCase).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// GetById 按主键查询
func (repo *MonitorDatabaseRepositoryImpl) GetById(id int64) (*entity.MonitorDatabase, error) {
	var data entity.MonitorDatabase
	err := repo.db.Model(&entity.MonitorDatabase{}).
		Where("id = ?", id).
		Not(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// Count 统计总数
func (repo *MonitorDatabaseRepositoryImpl) Count() (*int64, error) {
	var count int64
	err := repo.db.
		Model(&entity.MonitorDatabase{}).
		Not(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return &count, err
}
