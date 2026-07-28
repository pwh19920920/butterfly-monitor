package application

import (
	"context"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
)

type SysPermissionApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewSysPermissionApplication 创建权限应用服务
func NewSysPermissionApplication(sequence *snowflake.Node, repository *persistence.Repository) SysPermissionApplication {
	return SysPermissionApplication{sequence: sequence, repository: repository}
}

// Query 分页查询
func (application *SysPermissionApplication) Query(ctx context.Context, roleId int64) ([]entity.SysPermission, error) {
	data, err := application.repository.SysPermissionRepository.SelectByRoleId(roleId)

	// 错误记录
	if err != nil {
		logger.Error(ctx, "SysMenuRepository.Select() happen error for", err)
	}
	return data, err
}
