package persistence

import (
	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SysMenuRepositoryImpl struct {
	db *gorm.DB
}

func NewSysMenuRepositoryImpl(db *gorm.DB) *SysMenuRepositoryImpl {
	return &SysMenuRepositoryImpl{db: db}
}

// Save 保存
func (s *SysMenuRepositoryImpl) Save(menu *entity.SysMenu, options *[]entity.SysMenuOption) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if options != nil && len(*options) > 0 {
			if err := tx.Model(&entity.SysMenuOption{}).Create(options).Error; err != nil {
				return err
			}
		}
		return tx.Model(&entity.SysMenu{}).Create(&menu).Error
	})
}

// GetById 获取单条记录
func (s *SysMenuRepositoryImpl) GetById(id int64) (*entity.SysMenu, error) {
	var data entity.SysMenu
	err := s.db.Model(&entity.SysMenu{}).
		Where(&entity.SysMenu{BaseEntity: common.BaseEntity{Id: id}}).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// SelectByIds 批量获取
func (s *SysMenuRepositoryImpl) SelectByIds(ids []int64) ([]entity.SysMenu, error) {
	var data = make([]entity.SysMenu, 0)
	if ids == nil || len(ids) == 0 {
		return data, nil
	}
	err := s.db.Model(&entity.SysMenu{}).
		Where("id in ?", ids).
		Not(&entity.SysMenuOption{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

func (s *SysMenuRepositoryImpl) UpdateEntityAndChildRouteById(id int64, oldRoute string, menu *entity.SysMenu, options *[]entity.SysMenuOption) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的 update menu_option set deleted = 1 where menu_id = xx
		err := tx.Model(&entity.SysMenuOption{}).
			Where(&entity.SysMenuOption{MenuId: id}).
			Updates(&entity.SysMenuOption{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
		if err != nil {
			return err
		}

		// insert into on duplicate key update
		if options != nil && len(*options) > 0 {
			// insert into on duplicate key update
			if err = tx.Model(&entity.SysMenuOption{}).Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"deleted"}),
			}).Create(options).Error; err != nil {
				return err
			}
		}

		// UPDATE `config` SET `value`=REPLACE(`value`,'8080','8989') WHERE `value` LIKE '%8080%'
		if err = tx.Model(&entity.SysMenu{}).
			Where("route like ?", oldRoute+"%").
			UpdateColumn("route", gorm.Expr("REPLACE(route, ?, ?)", oldRoute, menu.Route)).Error; err != nil {
			return err
		}

		return tx.Model(&entity.SysMenu{}).
			Where(&entity.SysMenu{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&menu).Error
	})
}

// UpdateById 更新
func (s *SysMenuRepositoryImpl) UpdateById(id int64, menu *entity.SysMenu, options *[]entity.SysMenuOption) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的 update menu_option set deleted = 1 where menu_id = xx
		err := tx.Model(&entity.SysMenuOption{}).
			Where(&entity.SysMenuOption{MenuId: id}).
			Updates(&entity.SysMenuOption{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
		if err != nil {
			return err
		}

		if options != nil && len(*options) > 0 {
			// insert into on duplicate key update
			if err = tx.Model(&entity.SysMenuOption{}).Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"deleted"}),
			}).Create(options).Error; err != nil {
				return err
			}
		}

		// update menu base info
		return tx.Model(&entity.SysMenu{}).
			Where(&entity.SysMenu{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&menu).Error
	})
}

// Delete 删除
func (s *SysMenuRepositoryImpl) Delete(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.SysMenu{}).
			Where(&entity.SysMenu{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.SysMenuOption{}).
			Where(&entity.SysMenuOption{MenuId: id}).
			Updates(&entity.SysMenuOption{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
	})

}

// Select 查询分页
func (s *SysMenuRepositoryImpl) Select(req *types.SysMenuQueryRequest) (int64, []entity.SysMenu, error) {
	notCase := &entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	return paginate[entity.SysMenu](s.db, &entity.SysMenu{}, "1 = 1", nil, notCase, req.PageSize, req.Offset(), "id desc")
}

// SelectAll 查询全部
func (s *SysMenuRepositoryImpl) SelectAll() ([]entity.SysMenu, error) {
	var data []entity.SysMenu
	err := s.db.
		Model(&entity.SysMenu{}).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").Find(&data).Error
	return data, err
}

// CountByParent 统计子菜单数量
func (s *SysMenuRepositoryImpl) CountByParent(parentId int64) (int64, error) {
	var count int64 = 0
	err := s.db.
		Model(&entity.SysMenu{}).
		Where(&entity.SysMenu{Parent: &parentId}).
		Not(&entity.SysMenu{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return count, err
}
