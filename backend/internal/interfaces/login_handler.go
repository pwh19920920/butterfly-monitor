package interfaces

import (
	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type loginHandler struct {
	loginApp   application.LoginApplication
	sysUserApp application.SysUserApplication
}

// 登陆
func (hd *loginHandler) login(context *gin.Context) {
	var form types.LoginForm
	if err := context.ShouldBindJSON(&form); err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误:"+err.Error())
		return
	}

	// 数据校验
	err := form.ValidateForLogin()
	if err != nil {
		response.BuildResponseBadRequest(context, fmt.Sprintf("请求参数有误: %v", err.Error()))
		return
	}

	// option
	token, err := hd.loginApp.Login(context.Request.Context(), form.Username, form.Password)
	if err != nil {
		response.BuildResponseBadRequest(context, "用户名或者密码错误")
		return
	}

	// 输出
	response.BuildResponseSuccess(context, token)
}

// 退出
func (hd *loginHandler) logout(context *gin.Context) {
	// 尝试获取ticket
	ticket, err := GetUserTicket(context)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求数据有误")
		return
	}

	// 删除令牌
	_ = hd.loginApp.Logout(context.Request.Context(), ticket.Subject)

	// 输出
	response.BuildResponseSuccess(context, nil)
}

// 刷新令牌
func (hd *loginHandler) refresh(context *gin.Context) {
	// 尝试获取ticket
	ticket, err := GetUserTicket(context)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求数据有误")
		return
	}

	// 取令牌
	token := context.GetHeader(hd.loginApp.GetHeaderName(context.Request.Context()))
	newToken, err := hd.loginApp.RefreshToken(context.Request.Context(), ticket.UserId, ticket.Subject, token)
	if err != nil {
		response.BuildResponseBadRequest(context, "刷新令牌失败")
		return
	}

	// 删除令牌
	_ = hd.loginApp.Logout(context.Request.Context(), ticket.Subject)

	// 输出
	response.BuildResponseSuccess(context, newToken)
}

// 获取当前用户
func (hd *loginHandler) currentUser(context *gin.Context) {
	// 尝试获取ticket
	ticket, err := GetUserTicket(context)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求数据有误")
		return
	}

	user, err := hd.sysUserApp.GetUserById(context.Request.Context(), ticket.UserId)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求数据有误")
		return
	}

	// 用户序列化, 注意核心字段需要抹除
	user.Password = ""
	user.Salt = ""

	// 用户id, 获取全部角色
	// 通过角色获取全部菜单并集, 生成菜单树
	permission, err := hd.loginApp.GetUserMenuPermission(context.Request.Context(), ticket.UserId)
	if err != nil {
		response.BuildResponseSuccess(context, user)
		return
	}

	permission.SysUser = *user
	response.BuildResponseSuccess(context, permission)
}

// GetUserTicket 获取用户令牌
func GetUserTicket(context *gin.Context) (*entity.SysToken, error) {
	// 尝试获取ticket
	dataStr := context.Request.Header.Get(constant.ContextUser)
	ticket, err := entity.SysToken{}.UnMarshal(dataStr)
	if err != nil || ticket == nil {
		return nil, errors.New("用户获取令牌失败")
	}
	return ticket, err
}

// InitLoginHandler 加载路由
func InitLoginHandler(app *application.Application) {
	// 组件初始化
	handler := loginHandler{app.Login, app.SysUser}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/login", HandlerFunc: handler.login})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/logout", HandlerFunc: handler.logout})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/refresh", HandlerFunc: handler.refresh})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/currentUser", HandlerFunc: handler.currentUser})
	server.RegisterRoute("/api", route)
}
