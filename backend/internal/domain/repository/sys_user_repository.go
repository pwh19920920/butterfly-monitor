package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type SysUserRepository interface {
	// GetByUsername 通过用户名获取用户
	GetByUsername(username string) (*entity.SysUser, error)

	// GetById 通过id获取用户
	GetById(id int64) (*entity.SysUser, error)

	// Select 分页查询用户
	Select(request *types.SysUserQueryRequest) (int64, []entity.SysUser, error)

	// SelectAll 查询全部
	SelectAll() ([]entity.SysUser, error)

	// Create 创建
	Create(user *entity.SysUser) error

	// Modify 更新
	Modify(user *entity.SysUser) error
}
