package interfaces

import (
	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/types"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type sysRoleHandler struct {
	sysRoleApp application.SysRoleApplication
}

// 查询
func (handler *sysRoleHandler) query(context *gin.Context) {
	var sysRoleQueryRequest types.SysRoleQueryRequest
	if context.ShouldBindQuery(&sysRoleQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.sysRoleApp.Query(context.Request.Context(), &sysRoleQueryRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, sysRoleQueryRequest.RequestPaging, total, data)
}

// 查询
func (handler *sysRoleHandler) queryAll(context *gin.Context) {
	// option
	data, err := handler.sysRoleApp.SelectAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildResponseSuccess(context, data)
}

// 查询
func (handler *sysRoleHandler) queryByRoleId(context *gin.Context) {
	idStr := context.Param("roleId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	data, err := handler.sysRoleApp.QueryPermissionByRoleId(context.Request.Context(), id)
	if err != nil {
		response.BuildResponseBadRequest(context, "查询权限出错")
		return
	}

	// 输出
	response.BuildResponseSuccess(context, data)
}

// 创建
func (handler *sysRoleHandler) create(context *gin.Context) {
	var sysRoleCreateRequest types.SysRoleCreateRequest
	if context.ShouldBindJSON(&sysRoleCreateRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	err := sysRoleCreateRequest.ValidateForCreate()
	if err != nil {
		response.BuildResponseBadRequest(context, fmt.Sprintf("请求参数有误: %v", err.Error()))
		return
	}

	// option
	if err = handler.sysRoleApp.Create(context.Request.Context(), &sysRoleCreateRequest); err != nil {
		response.BuildResponseBadRequest(context, "创建角色失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 创建
func (handler *sysRoleHandler) modify(context *gin.Context) {
	var sysRoleCreateRequest types.SysRoleCreateRequest
	if context.ShouldBindJSON(&sysRoleCreateRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	err := sysRoleCreateRequest.ValidateForModify()
	if err != nil {
		response.BuildResponseBadRequest(context, fmt.Sprintf("请求参数有误: %v", err.Error()))
		return
	}

	// option
	if err = handler.sysRoleApp.Modify(context.Request.Context(), &sysRoleCreateRequest); err != nil {
		response.BuildResponseBadRequest(context, "创建角色失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 删除
func (handler *sysRoleHandler) delete(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	if err = handler.sysRoleApp.Delete(context.Request.Context(), id); err != nil {
		response.BuildResponseBadRequest(context, "删除角色失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitSysRoleHandler 加载路由
func InitSysRoleHandler(app *application.Application) {
	// 组件初始化
	handler := sysRoleHandler{app.SysRole}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/permission/:roleId", HandlerFunc: handler.queryByRoleId})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpDelete, Path: "/:id", HandlerFunc: handler.delete})
	server.RegisterRoute("/api/sys/role", route)
}
