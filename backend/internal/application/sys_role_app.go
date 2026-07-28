package application

import (
	"context"
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/config/sequence"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
)

type SysRoleApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewSysRoleApplication 创建角色应用服务
func NewSysRoleApplication(sequence *snowflake.Node, repository *persistence.Repository) SysRoleApplication {
	return SysRoleApplication{sequence: sequence, repository: repository}
}

// Query 分页查询
func (application *SysRoleApplication) Query(ctx context.Context, request *types.SysRoleQueryRequest) (int64, []entity.SysRole, error) {
	total, data, err := application.repository.SysRoleRepository.Select(request)
	// 错误记录
	if err != nil {
		logger.Error(ctx, "SysMenuRepository.Select() happen error for", err)
		return total, nil, err
	}
	return total, data, err
}

func (application *SysRoleApplication) SelectAll(ctx context.Context) ([]entity.SysRole, error) {
	data, err := application.repository.SysRoleRepository.SelectAll()
	// 错误记录
	if err != nil {
		logger.Error(ctx, "SysMenuRepository.Select() happen error for", err)
	}
	return data, err
}

// QueryPermissionByRoleId 查询
func (application *SysRoleApplication) QueryPermissionByRoleId(ctx context.Context, roleId int64) ([]types.SysRolePermissionQueryResponse, error) {
	data, err := application.repository.SysPermissionRepository.SelectByRoleId(roleId)
	// 错误记录
	if err != nil {
		logger.Error(ctx, "SysPermissionRepository.SelectByRoleId() happen error for", err)
		return nil, err
	}

	result := make([]types.SysRolePermissionQueryResponse, 0)
	for _, item := range data {
		options := make([]string, 0)
		if item.Option != "" {
			options = strings.Split(item.Option, ",")
		}
		result = append(result, types.SysRolePermissionQueryResponse{SysPermission: item, Options: options})
	}
	return result, nil
}

// Create 创建
func (application *SysRoleApplication) Create(ctx context.Context, request *types.SysRoleCreateRequest) error {
	role := common.Ptr(request.SysRole)
	role.Id = sequence.GetSequence().Generate().Int64()

	if request.Permissions != nil {
		for index, permission := range request.Permissions {
			permission.Id = sequence.GetSequence().Generate().Int64()
			permission.RoleId = role.Id
			request.Permissions[index] = permission
		}
	}
	return application.repository.SysRoleRepository.Save(&request.Permissions, role)
}

// Modify 创建
func (application *SysRoleApplication) Modify(ctx context.Context, request *types.SysRoleCreateRequest) error {
	role := common.Ptr(request.SysRole)
	if request.Permissions != nil {
		for index, permission := range request.Permissions {
			permission.Id = sequence.GetSequence().Generate().Int64()
			request.Permissions[index] = permission
		}
	}
	return application.repository.SysRoleRepository.UpdateById(request.Id, &request.Permissions, role)
}

// Delete 更新
func (application *SysRoleApplication) Delete(ctx context.Context, request int64) error {
	return application.repository.SysRoleRepository.Delete(request)
}
