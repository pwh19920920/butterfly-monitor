package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type SysPermissionRepositoryImpl struct {
	db *gorm.DB
}

func NewSysPermissionRepositoryImpl(db *gorm.DB) *SysPermissionRepositoryImpl {
	return &SysPermissionRepositoryImpl{db: db}
}

// SelectByRoleId 根据RoleId查询
func (s *SysPermissionRepositoryImpl) SelectByRoleId(roleId int64) ([]entity.SysPermission, error) {
	var data []entity.SysPermission
	err := s.db.Model(&entity.SysPermission{}).
		Where(&entity.SysPermission{RoleId: roleId}).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectByRoleIds 根据RoleIds查询
func (s *SysPermissionRepositoryImpl) SelectByRoleIds(roleIds []int64) ([]entity.SysPermission, error) {
	var data []entity.SysPermission
	err := s.db.Model(&entity.SysPermission{}).
		Where("role_id in ?", roleIds).
		Not(&entity.SysPermission{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

func (s *SysPermissionRepositoryImpl) CountByMenuId(menuId int64) (int64, error) {
	var count int64
	err := s.db.Model(&entity.SysPermission{}).
		Where("menu_id = ?", menuId).
		Not(&entity.SysPermission{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return count, err
}
