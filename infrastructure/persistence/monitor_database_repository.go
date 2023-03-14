package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
)

type MonitorDatabaseRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDatabaseRepositoryImpl(db *gorm.DB) *MonitorDatabaseRepositoryImpl {
	return &MonitorDatabaseRepositoryImpl{db: db}
}

func (repo *MonitorDatabaseRepositoryImpl) SelectAll(lastTime *common.LocalTime) ([]entity.MonitorDatabase, error) {
	var data []entity.MonitorDatabase
	tx := repo.db.Model(&entity.MonitorDatabase{})
	if lastTime != nil {
		tx.Where("updated_at >= ?", lastTime.Time)
	}
	err := tx.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *MonitorDatabaseRepositoryImpl) SelectSimpleAll() ([]entity.MonitorDatabase, error) {
	var data []entity.MonitorDatabase
	tx := repo.db.Model(&entity.MonitorDatabase{})
	err := tx.Select("id", "name", "database", "type").Order("id desc").Find(&data).Error
	return data, err
}

// Save 保存
func (repo *MonitorDatabaseRepositoryImpl) Save(jobDatabase *entity.MonitorDatabase) error {
	return repo.db.Model(&entity.MonitorDatabase{}).Create(&jobDatabase).Error
}

// UpdateById 更新
func (repo *MonitorDatabaseRepositoryImpl) UpdateById(id int64, jobDatabase *entity.MonitorDatabase) error {
	return repo.db.Model(&entity.MonitorDatabase{}).
		Where(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&jobDatabase).Error
}

// Delete 删除
func (repo *MonitorDatabaseRepositoryImpl) Delete(id int64) error {
	err := repo.db.Model(&entity.MonitorDatabase{}).
		Where(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: 1}}).Error
	return err
}

// Select 分页查询
func (repo *MonitorDatabaseRepositoryImpl) Select(req *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error) {
	var count int64 = 0
	whereCase := &entity.MonitorDatabase{
		Name: req.Name,
		Type: req.Type,
	}
	repo.db.Model(&entity.MonitorDatabase{}).Where(whereCase).Count(&count)

	var data []entity.MonitorDatabase
	err := repo.db.Model(&entity.MonitorDatabase{}).
		Where(whereCase).
		Not(&entity.MonitorDatabase{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
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
