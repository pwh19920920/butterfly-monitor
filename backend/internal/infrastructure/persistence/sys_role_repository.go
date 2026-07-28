package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"gorm.io/gorm"
)

type SysRoleRepositoryImpl struct {
	db *gorm.DB
}

func NewSysRoleRepositoryImpl(db *gorm.DB) *SysRoleRepositoryImpl {
	return &SysRoleRepositoryImpl{db: db}
}

// Save 保存
func (repo *SysRoleRepositoryImpl) Save(permissions *[]entity.SysPermission, role *entity.SysRole) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if permissions != nil && len(*permissions) != 0 {
			err := tx.Model(&entity.SysPermission{}).Create(permissions).Error
			if err != nil {
				return err
			}
		}
		return tx.Model(&entity.SysRole{}).Create(&role).Error
	})
}

// UpdateById 更新
func (repo *SysRoleRepositoryImpl) UpdateById(id int64, permissions *[]entity.SysPermission, role *entity.SysRole) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的权限
		err := tx.Where(&entity.SysPermission{
			RoleId: role.Id,
		}).Updates(&entity.SysPermission{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
		if err != nil {
			return err
		}

		// 插入新得权限
		if permissions != nil && len(*permissions) != 0 {
			err := tx.Model(&entity.SysPermission{}).Create(permissions).Error
			if err != nil {
				return err
			}
		}
		return tx.Model(&entity.SysRole{}).
			Where(&entity.SysRole{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&role).Error
	})
}

// Delete 删除
func (repo *SysRoleRepositoryImpl) Delete(id int64) error {
	err := repo.db.Model(&entity.SysRole{}).
		Where(&entity.SysRole{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&entity.SysRole{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
	return err
}

// Select 查询全部
func (repo *SysRoleRepositoryImpl) Select(req *types.SysRoleQueryRequest) (int64, []entity.SysRole, error) {
	var count int64 = 0
	notCase := &entity.SysRole{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	repo.db.Model(&entity.SysRole{}).Not(notCase).Count(&count)

	var data []entity.SysRole
	err := repo.db.Model(&entity.SysRole{}).
		Not(notCase).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

func (repo *SysRoleRepositoryImpl) SelectAll() ([]entity.SysRole, error) {
	var data []entity.SysRole
	notCase := &entity.SysRole{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	err := repo.db.Model(&entity.SysRole{}).Not(notCase).Find(&data).Error
	return data, err
}
