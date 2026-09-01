package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/config/auth"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/domain/security"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"

	"github.com/google/uuid"
	"github.com/pwh19920920/butterfly/pkg/helper"
	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
)

// 鉴权路径白名单：由 sync.Once 保证只初始化一次，初始化完成后不再被写，
// 鉴权中间件并发读这些 map/slice 是安全的（Once.Do 会阻塞并发调用直到初始化完成）。
var ignorePathMap = make(map[string]bool, 0)
var ignorePrefixPaths = make([]string, 0)
var commonPathMap = make(map[string]bool, 0)
var pathInitOnce sync.Once

func init() {
	ignorePathMap["POST - /api/login"] = true
	commonPathMap["POST - /api/logout"] = true
	commonPathMap["POST - /api/refresh"] = true
	commonPathMap["GET - /api/currentUser"] = true
}

type LoginApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	encoderService security.EncodeService
	tokenService   security.TokenService
	authConfig     *auth.Config
}

// NewLoginApplication 创建登录应用服务
func NewLoginApplication(
	sequence *snowflake.Node,
	repository *persistence.Repository,
	encoderService security.EncodeService,
	tokenService security.TokenService,
	authConfig *auth.Config,
) LoginApplication {
	return LoginApplication{
		sequence:       sequence,
		repository:     repository,
		encoderService: encoderService,
		tokenService:   tokenService,
		authConfig:     authConfig,
	}
}

// Logout 退出
func (application *LoginApplication) Logout(ctx context.Context, subject string) error {
	return application.repository.SysTokenRepository.Delete(subject)
}

// Login 登陆
func (application *LoginApplication) Login(ctx context.Context, username, password string) (ticket string, err error) {
	user, err := application.repository.SysUserRepository.GetByUsername(username)
	if err != nil || user == nil {
		return "", errors.New("用户不存在或者获取失败")
	}

	// 检查密码
	encPassword := application.encoderService.Encode(password, user.Salt)
	if encPassword != user.Password {
		return "", errors.New("用户密码不正确")
	}

	// 生成令牌数据
	return application.genericToken(ctx, user.Id)
}

// GetHeaderName 获取配置名称
func (application *LoginApplication) GetHeaderName(ctx context.Context) string {
	return application.authConfig.HeaderName
}

// CheckAndGetTicket 检查并获取用户id
func (application *LoginApplication) CheckAndGetTicket(ctx context.Context, token string) (*entity.SysToken, error) {
	// 取出票据id，并剥离 HeaderType 前缀（如 "Bearer "）
	rawToken, err := application.parseToken(token)
	if err != nil {
		return nil, errors.New("token数据不正确")
	}

	subject, err := application.tokenService.GetSubjectFromToken(rawToken)
	if err != nil {
		return nil, err
	}

	// 取出票据对象
	ticket, err := application.repository.SysTokenRepository.GetBySubject(subject)
	if err != nil {
		return nil, err
	}

	// 判断票据是否为空， 并校验
	if ticket == nil {
		return nil, errors.New("token不存在")
	}

	if !application.tokenService.CheckToken(rawToken, ticket.Secret) {
		return nil, errors.New("令牌校验失败")
	}

	// 校验成功，返回用户id
	return ticket, nil
}

// RefreshToken 刷新令牌
func (application *LoginApplication) RefreshToken(ctx context.Context, userId int64, subject, token string) (string, error) {
	// 取出票据id
	token, err := application.parseToken(token)
	if err != nil {
		return "", errors.New("token数据不正确")
	}

	// 生成令牌数据
	return application.genericToken(ctx, userId)
}

// GetAuthConfigPaths 获取忽略auth的地址，获取普通过滤的地址
func (application *LoginApplication) GetAuthConfigPaths(ctx context.Context) (ignorePathResultMap map[string]bool,
	ignorePrefixResultPaths []string, commonPathResultMap map[string]bool) {
	// sync.Once 保证初始化只执行一次且与后续读操作 happens-before：
	// 初始化期间并发调用会阻塞在 Do 内，直到完成才返回引用，故并发读安全。
	pathInitOnce.Do(func() {
		for _, v := range application.authConfig.IgnorePath {
			ignorePathMap[v] = true
		}

		ignorePrefixPaths = append(ignorePrefixPaths, application.authConfig.IgnorePrefixPath...)

		for _, v := range application.authConfig.CommonPath {
			commonPathMap[v] = true
		}
	})
	return ignorePathMap, ignorePrefixPaths, commonPathMap
}

func (application *LoginApplication) SaveToken(ctx context.Context, token entity.SysToken) error {
	return application.repository.SysTokenRepository.Save(token)
}

func (application *LoginApplication) ModifyToken(ctx context.Context, token entity.SysToken) error {
	return application.repository.SysTokenRepository.ModifyById(token, token.Id)
}

// 生成令牌
func (application *LoginApplication) genericToken(ctx context.Context, userId int64) (string, error) {
	// 生成保存密钥
	secret := uuid.New().String()
	subject := uuid.New().String()

	// 保存用户信息与令牌之间的关系
	// subject -> userId
	// subject -> secret
	// userId -> subject
	expireTime := time.Now().Add(time.Duration(application.authConfig.ExpireTime) * time.Second)
	err := application.repository.SysTokenRepository.Save(entity.SysToken{
		Secret:   secret,
		Subject:  subject,
		UserId:   userId,
		ExpireAt: &common.LocalTime{Time: expireTime},
	})

	// 判定是否保存失败
	if err != nil {
		logger.Error(ctx, err)
		return "", errors.New("密钥保存失败")
	}

	// 生成令牌数据
	return application.tokenService.GenericToken(secret, subject, expireTime)
}

// GenericToken 暴露创建令牌
func (application *LoginApplication) GenericToken(ctx context.Context, secret, subject string, expireTime time.Time) (string, error) {
	return application.tokenService.GenericToken(secret, subject, expireTime)
}

// GetTokenBySubject 暴露获取令牌
func (application *LoginApplication) GetTokenBySubject(ctx context.Context, subject string) (*entity.SysToken, error) {
	return application.repository.SysTokenRepository.GetBySubject(subject)
}

// 从header中解析令牌
func (application *LoginApplication) parseToken(token string) (string, error) {
	// 检查数据
	typeKey := fmt.Sprintf("%s ", application.authConfig.HeaderType)
	typeIndex := strings.Index(token, typeKey)
	if typeIndex != 0 {
		return "", errors.New("token数据不正确")
	}

	// 取出票据id
	return helper.StringHelper.SubString(token, len(typeKey), len(token)), nil
}

// GetUserMenuPermission 用户菜单权限
func (application *LoginApplication) GetUserMenuPermission(ctx context.Context, userId int64) (*types.SysMenuPermissionForUser, error) {
	sysPermissions, err := application.GetUserSysPermission(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 计算菜单id列表, 计算操作id列表
	menuIdMap := make(map[int64]string, 0)
	menuIds := make([]int64, 0)
	opIdMap := make(map[int64]string, 0)
	opIds := make([]int64, 0)
	for _, permission := range sysPermissions {
		_, ok := menuIdMap[permission.MenuId]
		if !ok {
			menuIdMap[permission.MenuId] = ""
			menuIds = append(menuIds, permission.MenuId)
		}

		if permission.Option != "" {
			application.parsePermissionOptions(permission.Option, opIdMap, &opIds)
		}
	}

	// 获取数据组成树
	allMenus, err := application.repository.SysMenuRepository.SelectByIds(menuIds)
	if err != nil {
		return nil, err
	}

	menuCodes := make([]string, 0)
	rootMenus := make([]types.SysMenuPermissionForUserMenu, 0)
	var menuMap = make(map[int64][]types.SysMenuPermissionForUserMenu, 0)
	for _, item := range allMenus {
		// 放code
		menuCodes = append(menuCodes, item.Code)

		// 放数据到menuMap
		menu, ok := menuMap[*item.Parent]
		if !ok {
			menu = make([]types.SysMenuPermissionForUserMenu, 0)
		}

		menu = append(menu, types.SysMenuPermissionForUserMenu{
			Id:        item.Id,
			Icon:      item.Icon,
			Component: item.Component,
			Path:      item.Path,
			Name:      item.Code,
		})
		menuMap[*item.Parent] = menu

		// 得到rootMenus
		if *item.Parent == 0 {
			rootMenus = append(rootMenus, types.SysMenuPermissionForUserMenu{
				Id:        item.Id,
				Icon:      item.Icon,
				Component: item.Component,
				Path:      item.Path,
				Name:      item.Code,
			})
		}
	}

	application.recursionAssignmentForUserMenu(rootMenus, menuMap)

	// 获取操作组成树
	menuOptions, err := application.repository.SysMenuOptionRepository.SelectByIds(opIds)
	if err != nil {
		return nil, err
	}

	optionValues := make([]string, 0)
	for _, option := range menuOptions {
		optionValues = append(optionValues, option.Value)
	}
	return &types.SysMenuPermissionForUser{
		Permissions: optionValues,
		Menus:       rootMenus,
		Codes:       menuCodes,
	}, nil
}

// GetUserMenuUrl 获取用户拥有的权限路径
func (application *LoginApplication) GetUserMenuUrl(ctx context.Context, userId int64) (map[string]bool, error) {
	specMap := make(map[string]bool, 0)
	options, err := application.GetUserSysMenuOption(ctx, userId)
	if err != nil {
		return specMap, err
	}

	for _, option := range options {
		fullKey := fmt.Sprintf("%s - %s", option.Method, option.Path)
		specMap[fullKey] = true
	}
	return specMap, err
}

// GetUserSysMenuOption 获取用户拥有的路径
func (application *LoginApplication) GetUserSysMenuOption(ctx context.Context, userId int64) ([]entity.SysMenuOption, error) {
	sysPermissions, err := application.GetUserSysPermission(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 计算菜单id列表, 计算操作id列表
	opIdMap := make(map[int64]string, 0)
	opIds := make([]int64, 0)
	for _, permission := range sysPermissions {
		if permission.Option != "" {
			application.parsePermissionOptions(permission.Option, opIdMap, &opIds)
		}
	}
	// 获取操作组成树
	return application.repository.SysMenuOptionRepository.SelectByIds(opIds)
}

// GetUserSysPermission 获取用户所拥有的权限列表
func (application *LoginApplication) GetUserSysPermission(ctx context.Context, userId int64) ([]entity.SysPermission, error) {
	sysUser, err := application.repository.SysUserRepository.GetById(userId)
	if err != nil {
		return nil, err
	}

	if sysUser.Roles == "" {
		return nil, errors.New("角色信息不存在")
	}

	// 获取角色id列表
	roleIds := make([]int64, 0)
	roleStrIds := strings.Split(sysUser.Roles, ",")
	for _, roleStrId := range roleStrIds {
		roleId, err := strconv.ParseInt(roleStrId, 10, 64)
		if err != nil {
			return nil, errors.New("角色信息有误")
		}
		roleIds = append(roleIds, roleId)
	}

	// 角色id列表到permission表中查询
	return application.repository.SysPermissionRepository.SelectByRoleIds(roleIds)
}

// parsePermissionOptions 解析权限 CSV 选项字符串，去重追加到 opIds，GetUserMenuPermission 和 GetUserSysMenuOption 共用。
func (application *LoginApplication) parsePermissionOptions(option string, opIdMap map[int64]string, opIds *[]int64) {
	for _, opStrId := range strings.Split(option, ",") {
		opId, err := strconv.ParseInt(opStrId, 10, 64)
		if err != nil {
			continue
		}
		if _, ok := opIdMap[opId]; !ok {
			opIdMap[opId] = ""
			*opIds = append(*opIds, opId)
		}
	}
}

func (application *LoginApplication) recursionAssignmentForUserMenu(
	rootMenus []types.SysMenuPermissionForUserMenu,
	menuMap map[int64][]types.SysMenuPermissionForUserMenu) {

	for index, item := range rootMenus {
		menus, _ := menuMap[item.Id]
		item.Routes = menus
		rootMenus[index] = item

		// 判断退出条件
		if len(menus) != 0 {
			// 继续赋值
			application.recursionAssignmentForUserMenu(menus, menuMap)
		}
	}
}
