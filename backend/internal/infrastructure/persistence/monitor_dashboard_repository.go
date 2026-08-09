package persistence

import (
	"errors"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorDashboardRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDashboardRepositoryImpl(db *gorm.DB) *MonitorDashboardRepositoryImpl {
	return &MonitorDashboardRepositoryImpl{db: db}
}

// SelectSimpleAll 简单查询
func (repo *MonitorDashboardRepositoryImpl) SelectSimpleAll() ([]entity.MonitorDashboard, error) {
	var data []entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Select("id", "name", "slug").
		Order("id desc").
		Find(&data).Error
	return data, err
}

// Save 保存
func (repo *MonitorDashboardRepositoryImpl) Save(monitorDashboard *entity.MonitorDashboard) error {
	return repo.db.Model(&entity.MonitorDashboard{}).Create(&monitorDashboard).Error
}

// UpdateById 按主键更新
func (repo *MonitorDashboardRepositoryImpl) UpdateById(id int64, monitorDashboard *entity.MonitorDashboard) error {
	return repo.db.Model(&entity.MonitorDashboard{}).
		Where(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorDashboard).Error
}

// GetById 按主键查询
func (repo *MonitorDashboardRepositoryImpl) GetById(id int64) (*entity.MonitorDashboard, error) {
	var data entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Where(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Id: id}}).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// SelectByIds 批量查询
func (repo *MonitorDashboardRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorDashboard, error) {
	var data []entity.MonitorDashboard
	if ids == nil || len(ids) == 0 {
		return data, nil
	}
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Where("id in ?", ids).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Select 分页查询
func (repo *MonitorDashboardRepositoryImpl) Select(req *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error) {
	var count int64 = 0
	notCase := &entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	whereSql := "1 = 1"
	whereArg := make([]interface{}, 0)
	if req.Name != "" {
		whereSql += " and name like ?"
		whereArg = append(whereArg, "%"+req.Name+"%")
	}
	repo.db.Model(&entity.MonitorDashboard{}).Where(whereSql, whereArg...).Not(notCase).Count(&count)

	var data []entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Where(whereSql, whereArg...).
		Not(notCase).
		Order("id desc").
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// SelectAll 查询全部
func (repo *MonitorDashboardRepositoryImpl) SelectAll() ([]entity.MonitorDashboard, error) {
	var data []entity.MonitorDashboard
	err := repo.db.Model(&entity.MonitorDashboard{}).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Count 统计总数
func (repo *MonitorDashboardRepositoryImpl) Count() (*int64, error) {
	var count int64
	err := repo.db.
		Model(&entity.MonitorDashboard{}).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return &count, err
}
