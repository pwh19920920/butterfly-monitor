package interfaces

import (
	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type sysUserHandler struct {
	sysUserApp application.SysUserApplication
}

// 查询
func (handler *sysUserHandler) query(context *gin.Context) {
	var sysUserQueryRequest types.SysUserQueryRequest
	if context.ShouldBindQuery(&sysUserQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.sysUserApp.Query(context.Request.Context(), &sysUserQueryRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, sysUserQueryRequest.RequestPaging, total, data)
}

// 查询全部
func (handler *sysUserHandler) queryAll(context *gin.Context) {
	// option
	data, err := handler.sysUserApp.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildResponseSuccess(context, data)
}

// 创建
func (handler *sysUserHandler) create(context *gin.Context) {
	var sysUser entity.SysUser
	if context.ShouldBindJSON(&sysUser) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err := handler.sysUserApp.Create(context.Request.Context(), &sysUser)
	if err != nil {
		response.BuildResponseBadRequest(context, "创建用户失败")
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// 创建
func (handler *sysUserHandler) modify(context *gin.Context) {
	var sysUser entity.SysUser
	if context.ShouldBindJSON(&sysUser) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err := handler.sysUserApp.Modify(context.Request.Context(), &sysUser)
	if err != nil {
		response.BuildResponseBadRequest(context, "更新用户失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitSysUserHandler 加载路由
func InitSysUserHandler(app *application.Application) {
	// 组件初始化
	handler := sysUserHandler{app.SysUser}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	server.RegisterRoute("/api/sys/user", route)
}
