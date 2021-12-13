package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
)

type AlertChannelRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertChannelRepositoryImpl(db *gorm.DB) *AlertChannelRepositoryImpl {
	return &AlertChannelRepositoryImpl{db: db}
}

// Select 查询
func (repo *AlertChannelRepositoryImpl) Select(req *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error) {
	var count int64 = 0
	_ = repo.db.Model(&entity.AlertChannel{}).Count(&count)

	var data []entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Order("id desc").
		Not(&entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// SelectAll 查询
func (repo *AlertChannelRepositoryImpl) SelectAll() ([]entity.AlertChannel, error) {
	var data []entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Order("id desc").
		Not(&entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// GetById 查询
func (repo *AlertChannelRepositoryImpl) GetById(id int64) (entity.AlertChannel, error) {
	var data entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Where("id = ?", id).Find(&data).Error
	return data, err
}

// Delete 删除
func (repo *AlertChannelRepositoryImpl) Delete(id int64) error {
	return repo.db.Where("id = ?", id).Updates(&entity.AlertChannel{
		BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
	}).Error
}

// Save 保存
func (repo *AlertChannelRepositoryImpl) Save(alertChannel *entity.AlertChannel) error {
	return repo.db.Model(&entity.AlertChannel{}).Create(alertChannel).Error
}

// Modify 更新
func (repo *AlertChannelRepositoryImpl) Modify(id int64, alertChannel *entity.AlertChannel) error {
	return repo.db.
		Where(&entity.AlertChannel{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&alertChannel).Error
}
