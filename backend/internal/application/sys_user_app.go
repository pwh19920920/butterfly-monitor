package application

import (
	"butterfly-monitor/internal/config/auth"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/domain/security"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"
	"context"
	"strings"

	"github.com/go-basic/uuid"
	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
)

type SysUserApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	encoderService security.EncodeService
	tokenService   security.TokenService
	authConfig     *auth.Config
}

// NewSysUserApplication 创建用户应用服务
func NewSysUserApplication(
	sequence *snowflake.Node,
	repository *persistence.Repository,
	encoderService security.EncodeService,
	tokenService security.TokenService,
	authConfig *auth.Config,
) SysUserApplication {
	return SysUserApplication{
		sequence:       sequence,
		repository:     repository,
		encoderService: encoderService,
		tokenService:   tokenService,
		authConfig:     authConfig,
	}
}

func (app *SysUserApplication) GetUserById(ctx context.Context, userId int64) (*entity.SysUser, error) {
	return app.repository.SysUserRepository.GetById(userId)
}

func (app *SysUserApplication) GetUserByUsername(ctx context.Context, username string) (*entity.SysUser, error) {
	return app.repository.SysUserRepository.GetByUsername(username)
}

func (app *SysUserApplication) coverQueryResult(data []entity.SysUser) []types.SysUserQueryResponse {
	// 重新赋值
	result := make([]types.SysUserQueryResponse, 0)
	for _, item := range data {
		roleList := make([]string, 0)
		if item.Roles != "" {
			roleList = strings.Split(item.Roles, ",")
		}
		result = append(result, types.SysUserQueryResponse{
			BaseEntity: item.BaseEntity,
			Name:       item.Name,
			Avatar:     item.Avatar,
			Roles:      item.Roles,
			Username:   item.Username,
			RoleList:   roleList,
			Email:      item.Email,
			Mobile:     item.Mobile,
		})
	}
	return result
}

func (app *SysUserApplication) Query(ctx context.Context, request *types.SysUserQueryRequest) (int64, []types.SysUserQueryResponse, error) {
	total, data, err := app.repository.SysUserRepository.Select(request)
	// 错误记录
	if err != nil {
		logger.Error(ctx, "SysUserRepository.Select() happen error for", err)
		return total, nil, err
	}

	// 重新赋值
	result := app.coverQueryResult(data)
	return total, result, err
}

// QueryAll 查询全部
func (app *SysUserApplication) QueryAll(ctx context.Context) ([]types.SysUserQueryResponse, error) {
	data, err := app.repository.SysUserRepository.SelectAll()
	// 重新赋值
	result := app.coverQueryResult(data)
	return result, err
}

// Create 创建用户
func (app *SysUserApplication) Create(ctx context.Context, user *entity.SysUser) error {
	user.Id = app.sequence.Generate().Int64()
	user.Salt = uuid.New()
	user.Password = app.encoderService.Encode(user.Password, user.Salt)
	return app.repository.SysUserRepository.Create(user)
}

// Modify 更新用户
func (app *SysUserApplication) Modify(ctx context.Context, user *entity.SysUser) error {
	if user.Password != "" {
		user.Salt = uuid.New()
		user.Password = app.encoderService.Encode(user.Password, user.Salt)
	}
	return app.repository.SysUserRepository.Modify(user)
}
