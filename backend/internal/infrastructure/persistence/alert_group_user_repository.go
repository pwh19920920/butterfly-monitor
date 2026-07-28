package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type AlertGroupUserRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertGroupUserRepositoryImpl(db *gorm.DB) *AlertGroupUserRepositoryImpl {
	return &AlertGroupUserRepositoryImpl{db: db}
}

// SelectByGroupId 按分组 id 查询用户关联
func (repo *AlertGroupUserRepositoryImpl) SelectByGroupId(groupId int64) ([]entity.AlertGroupUser, error) {
	var data []entity.AlertGroupUser
	err := repo.db.Model(&entity.AlertGroupUser{}).
		Where("group_id = ?", groupId).
		Not(&entity.AlertGroupUser{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectUsersByUserIds 按用户 id 列表查询 t_sys_user
func (repo *AlertGroupUserRepositoryImpl) SelectUsersByUserIds(userIds []int64) ([]entity.SysUser, error) {
	if userIds == nil || len(userIds) == 0 {
		return make([]entity.SysUser, 0), nil
	}
	var data []entity.SysUser
	err := repo.db.Model(&entity.SysUser{}).
		Where("id in (?)", userIds).
		Not(&entity.SysUser{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}
