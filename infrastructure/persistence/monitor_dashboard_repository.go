package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"errors"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
)

type MonitorDashboardRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDashboardRepositoryImpl(db *gorm.DB) *MonitorDashboardRepositoryImpl {
	return &MonitorDashboardRepositoryImpl{db: db}
}

func (repo *MonitorDashboardRepositoryImpl) SelectSimpleAll() ([]entity.MonitorDashboard, error) {
	var data []entity.MonitorDashboard
	tx := repo.db.Model(&entity.MonitorDashboard{})
	err := tx.Select("id", "name", "slug").Order("id desc").Find(&data).Error
	return data, err
}

// Save 保存
func (repo *MonitorDashboardRepositoryImpl) Save(monitorDashboard *entity.MonitorDashboard) error {
	return repo.db.Model(&entity.MonitorDashboard{}).Create(&monitorDashboard).Error
}

// UpdateById 更新
func (repo *MonitorDashboardRepositoryImpl) UpdateById(id int64, monitorDashboard *entity.MonitorDashboard) error {
	return repo.db.Model(&entity.MonitorDashboard{}).
		Where(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorDashboard).Error
}

func (repo *MonitorDashboardRepositoryImpl) GetById(id int64) (*entity.MonitorDashboard, error) {
	var data entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).First(&data, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

func (repo *MonitorDashboardRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorDashboard, error) {
	var data []entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Where("id in ?", ids).
		Find(&data).Error
	return data, err
}

// Select 分页查询
func (repo *MonitorDashboardRepositoryImpl) Select(req *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error) {
	var count int64 = 0
	repo.db.Model(&entity.MonitorDashboard{}).Count(&count)

	var data []entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}
