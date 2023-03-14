package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly-admin/interfaces"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
	"time"
)

type sysLoginHandler struct {
	app *application.Application
}

// authorize
func (handler *sysLoginHandler) authorize(context *gin.Context) {
	// 尝试获取ticket
	ticket, err := interfaces.GetUserTicket(context)
	if err != nil || ticket == nil {
		response.BuildResponseBadRequest(context, "请求数据有误")
		return
	}

	// 取数据
	var sysLoginAuthorizeRequest types.SysLoginAuthorizeRequest
	if err := context.ShouldBind(&sysLoginAuthorizeRequest); err != nil || sysLoginAuthorizeRequest.ClientId == "" {
		response.BuildResponseBadRequest(context, "请求数据有误: 客户端id为空")
		return
	}

	// 校验客户端id
	if handler.app.AllConfig.Grafana.ClientId != sysLoginAuthorizeRequest.ClientId {
		response.BuildResponseBadRequest(context, "客户端id不正确, 请确认")
		return
	}

	// 响应给客户端
	response.BuildResponseSuccess(context, types.SysLoginAuthorizeResponse{
		Code: ticket.Subject, RedirectUrl: handler.app.MonitorDashboardApp.Grafana.Addr + "/login/generic_oauth",
	})
}

// token
func (handler *sysLoginHandler) token(context *gin.Context) {
	var sysLoginTokenRequest types.SysLoginTokenRequest
	if err := context.ShouldBind(&sysLoginTokenRequest); err != nil || sysLoginTokenRequest.Code == "" {
		response.BuildResponseBadRequest(context, "请求数据有误: code为空")
		return
	}

	authorization := context.GetHeader(handler.app.AllConfig.AdminConfig.AuthConfig.HeaderName)
	baseAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", handler.app.AllConfig.Grafana.ClientId, handler.app.AllConfig.Grafana.Secret)))
	if fmt.Sprintf("Basic %s", baseAuth) != authorization {
		response.BuildResponseBadRequest(context, "非法访问, 密钥对不匹配")
		return
	}

	ticket, err := handler.app.AdminApp.Login.GetTokenBySubject(sysLoginTokenRequest.Code)
	if err != nil || ticket == nil {
		response.BuildResponseBadRequest(context, "用户获取令牌失败")
		return
	}

	token, err := handler.app.AdminApp.Login.GenericToken(ticket.Secret, ticket.Subject, ticket.ExpireAt.Time)
	if err != nil {
		response.BuildResponseBadRequest(context, "创建访问令牌失败")
		return
	}

	context.JSON(200, types.SysLoginTokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiryIn:    ticket.ExpireAt.Sub(time.Now()).Seconds(),
	})
}

// userInfo
func (handler *sysLoginHandler) userInfo(context *gin.Context) {
	token := context.GetHeader(handler.app.AdminApp.Login.GetHeaderName())
	ticket, err := handler.app.AdminApp.Login.CheckAndGetTicket(token)
	if err != nil {
		response.BuildResponseBadRequest(context, "创建访问令牌失败")
		return
	}

	user, err := handler.app.AdminApp.SysUser.GetUserById(ticket.UserId)
	if err != nil || user == nil {
		response.BuildResponseBadRequest(context, "获取用户信息失败")
		return
	}

	context.JSON(200, types.SysLoginUserInfoResponse{
		Name:  user.Email,
		Email: user.Email,
	})
}

// InitSysLoginHandler 加载路由
func InitSysLoginHandler(app *application.Application) {
	// 组件初始化
	handler := sysLoginHandler{app}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/oauth/authorize", HandlerFunc: handler.authorize})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/oauth/token", HandlerFunc: handler.token})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/oauth/userInfo", HandlerFunc: handler.userInfo})
	server.RegisterRoute("/api", route)
}
