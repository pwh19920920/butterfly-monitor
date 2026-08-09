package persistence

import (
	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type SysMenuOptionRepositoryImpl struct {
	db *gorm.DB
}

func NewSysMenuOptionRepositoryImpl(db *gorm.DB) *SysMenuOptionRepositoryImpl {
	return &SysMenuOptionRepositoryImpl{db: db}
}

// SelectByMenuId 通过菜单id搜索
func (repo *SysMenuOptionRepositoryImpl) SelectByMenuId(menuId int64) ([]entity.SysMenuOption, error) {
	var data []entity.SysMenuOption
	err := repo.db.Model(&entity.SysMenuOption{}).
		Where("menu_id = ?", menuId).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectAll 查询全部
func (repo *SysMenuOptionRepositoryImpl) SelectAll() ([]entity.SysMenuOption, error) {
	var data []entity.SysMenuOption
	err := repo.db.Model(&entity.SysMenuOption{}).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectByIds 批量查询
func (repo *SysMenuOptionRepositoryImpl) SelectByIds(ids []int64) ([]entity.SysMenuOption, error) {
	var data = make([]entity.SysMenuOption, 0)
	if ids == nil || len(ids) == 0 {
		return data, nil
	}
	err := repo.db.Model(&entity.SysMenuOption{}).
		Where("id in ?", ids).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// BatchUpdate 批量更新
func (repo *SysMenuOptionRepositoryImpl) BatchUpdate(options []entity.SysMenuOption) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for _, option := range options {
			if err := tx.Updates(&option).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
