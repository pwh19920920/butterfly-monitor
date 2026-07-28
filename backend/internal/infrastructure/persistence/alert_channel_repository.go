package persistence

import (
	"errors"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"gorm.io/gorm"
)

type AlertChannelRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertChannelRepositoryImpl(db *gorm.DB) *AlertChannelRepositoryImpl {
	return &AlertChannelRepositoryImpl{db: db}
}

// Select 分页查询
func (repo *AlertChannelRepositoryImpl) Select(req *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error) {
	var count int64 = 0
	notCase := &entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	whereSql := "1 = 1"
	whereArg := make([]interface{}, 0)
	if req.Name != "" {
		whereSql += " and name like ?"
		whereArg = append(whereArg, "%"+req.Name+"%")
	}
	if req.Type != nil {
		whereSql += " and type = ?"
		whereArg = append(whereArg, req.Type)
	}
	_ = repo.db.Model(&entity.AlertChannel{}).Where(whereSql, whereArg...).Not(notCase).Count(&count)

	var data []entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Order("id desc").
		Where(whereSql, whereArg...).
		Not(notCase).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// SelectAll 查询全部
func (repo *AlertChannelRepositoryImpl) SelectAll() ([]entity.AlertChannel, error) {
	var data []entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Order("id desc").
		Not(&entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// GetById 按主键查询
func (repo *AlertChannelRepositoryImpl) GetById(id int64) (*entity.AlertChannel, error) {
	var data entity.AlertChannel
	err := repo.db.Model(&entity.AlertChannel{}).
		Where("id = ?", id).
		Not(&entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// Save 保存
func (repo *AlertChannelRepositoryImpl) Save(channel *entity.AlertChannel) error {
	return repo.db.Model(&entity.AlertChannel{}).Create(channel).Error
}

// Modify 按主键更新；显式 Select template，允许清空回落到默认模板
func (repo *AlertChannelRepositoryImpl) Modify(id int64, channel *entity.AlertChannel) error {
	return repo.db.
		Where(&entity.AlertChannel{BaseEntity: common.BaseEntity{Id: id}}).
		Select("name", "type", "params", "handler", "fail_route", "template").
		Updates(&channel).Error
}

// Count 统计通道总数
func (repo *AlertChannelRepositoryImpl) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&entity.AlertChannel{}).
		Not(&entity.AlertChannel{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return count, err
}
