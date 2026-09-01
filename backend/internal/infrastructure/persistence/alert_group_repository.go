package persistence

import (
	"errors"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"gorm.io/gorm"
)

type AlertGroupRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertGroupRepositoryImpl(db *gorm.DB) *AlertGroupRepositoryImpl {
	return &AlertGroupRepositoryImpl{db: db}
}

// SelectAll 查询全部
func (repo *AlertGroupRepositoryImpl) SelectAll() ([]entity.AlertGroup, error) {
	var data []entity.AlertGroup
	err := repo.db.Model(&entity.AlertGroup{}).
		Not(&entity.AlertGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Select 分页查询
func (repo *AlertGroupRepositoryImpl) Select(req *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error) {
	notCase := &entity.AlertGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	whereSql := "1 = 1"
	whereArg := make([]interface{}, 0)
	if req.Name != "" {
		whereSql += " and name like ?"
		whereArg = append(whereArg, "%"+req.Name+"%")
	}
	return paginate[entity.AlertGroup](repo.db, &entity.AlertGroup{}, whereSql, whereArg, notCase, req.PageSize, req.Offset(), "id desc")
}

// GetById 按主键查询
func (repo *AlertGroupRepositoryImpl) GetById(id int64) (*entity.AlertGroup, error) {
	var data entity.AlertGroup
	err := repo.db.Model(&entity.AlertGroup{}).
		Where(&entity.AlertGroup{BaseEntity: common.BaseEntity{Id: id}}).
		Not(&entity.AlertGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// Save 事务保存分组与用户关联
func (repo *AlertGroupRepositoryImpl) Save(group *entity.AlertGroup, users []entity.AlertGroupUser) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := repo.createAlertGroupUsers(tx, users); err != nil {
			return err
		}
		return tx.Model(&entity.AlertGroup{}).Create(&group).Error
	})
}

// Modify 事务更新分组并重建用户关联
func (repo *AlertGroupRepositoryImpl) Modify(id int64, group *entity.AlertGroup, users []entity.AlertGroupUser) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 软删除旧关联
		if err := tx.Where("group_id = ?", id).Updates(&entity.AlertGroupUser{
			BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
		}).Error; err != nil {
			return err
		}

		// 创建新关联
		if err := repo.createAlertGroupUsers(tx, users); err != nil {
			return err
		}

		return tx.Model(&entity.AlertGroup{}).
			Where(&entity.AlertGroup{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&group).Error
	})
}

// createAlertGroupUsers 批量创建分组用户关联，Save 和 Modify 共用。
func (repo *AlertGroupRepositoryImpl) createAlertGroupUsers(tx *gorm.DB, users []entity.AlertGroupUser) error {
	if len(users) == 0 {
		return nil
	}
	return tx.Model(&entity.AlertGroupUser{}).Create(users).Error
}

// Count 统计分组总数
func (repo *AlertGroupRepositoryImpl) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&entity.AlertGroup{}).
		Not(&entity.AlertGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return count, err
}
