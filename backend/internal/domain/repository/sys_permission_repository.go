package repository

import (
	"butterfly-monitor/internal/domain/entity"
)

type SysPermissionRepository interface {

	// SelectByRoleId 根据角色id查询
	SelectByRoleId(roleId int64) ([]entity.SysPermission, error)

	// SelectByRoleIds 批量根据角色id查询
	SelectByRoleIds(roleIds []int64) ([]entity.SysPermission, error)

	// CountByMenuId 统计权限
	CountByMenuId(menuId int64) (int64, error)
}
