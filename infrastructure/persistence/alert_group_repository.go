package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
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

// Select 获取分组
func (repo *AlertGroupRepositoryImpl) Select(req *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error) {
	var count int64 = 0
	_ = repo.db.Model(&entity.AlertGroup{}).Count(&count)

	var data []entity.AlertGroup
	err := repo.db.Model(&entity.AlertGroup{}).
		Order("id desc").
		Not(&entity.AlertGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// Save 保存
func (repo *AlertGroupRepositoryImpl) Save(alertGroup *entity.AlertGroup, alertGroupUsers []entity.AlertGroupUser) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if alertGroupUsers != nil && len(alertGroupUsers) > 0 {
			err := tx.Model(&entity.AlertGroupUser{}).Create(alertGroupUsers).Error
			if err != nil {
				return err
			}
		}
		return tx.Model(&entity.AlertGroup{}).Create(&alertGroup).Error
	})
}

// Modify 更新
func (repo *AlertGroupRepositoryImpl) Modify(id int64, alertGroup *entity.AlertGroup, alertGroupUsers []entity.AlertGroupUser) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的alertGroupUser
		err := tx.Where("group_id = ?", alertGroup.Id).Updates(&entity.AlertGroupUser{
			BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
		}).Error
		if err != nil {
			return err
		}

		// 创建新的alertGroupUsers
		if alertGroupUsers != nil && len(alertGroupUsers) > 0 {
			err := tx.Model(&entity.AlertGroupUser{}).Create(alertGroupUsers).Error
			if err != nil {
				return err
			}
		}

		// 更新数据库
		return tx.Model(&entity.AlertGroup{}).Where(&entity.AlertGroup{
			BaseEntity: common.BaseEntity{Id: id},
		}).Updates(&alertGroup).Error
	})
}
