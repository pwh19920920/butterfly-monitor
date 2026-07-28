package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type SysMenuRepository interface {
	// Save 保存
	Save(menu *entity.SysMenu, options *[]entity.SysMenuOption) error

	// UpdateById 更新
	UpdateById(id int64, menu *entity.SysMenu, options *[]entity.SysMenuOption) error

	// UpdateEntityAndChildRouteById 更新
	UpdateEntityAndChildRouteById(id int64, oldRoute string, menu *entity.SysMenu, options *[]entity.SysMenuOption) error

	// GetById 获取单条记录
	GetById(id int64) (*entity.SysMenu, error)

	// SelectByIds 批量获取
	SelectByIds(ids []int64) ([]entity.SysMenu, error)

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.SysMenuQueryRequest) (int64, []entity.SysMenu, error)

	// SelectAll 查询全部
	SelectAll() ([]entity.SysMenu, error)

	// CountByParent 统计下级菜单数量
	CountByParent(parentId int64) (int64, error)
}
