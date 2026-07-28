package repository

import (
	"dragonfly-monitor/internal/domain/entity"
)

type SysMenuOptionRepository interface {

	// SelectByMenuId 通过菜单id搜索
	SelectByMenuId(menuId int64) ([]entity.SysMenuOption, error)

	// SelectAll 查询全部
	SelectAll() ([]entity.SysMenuOption, error)

	// SelectByIds 批量查询
	SelectByIds(ids []int64) ([]entity.SysMenuOption, error)

	// BatchUpdate 批量更新
	BatchUpdate(options []entity.SysMenuOption) error
}
