package interfaces

import (
	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorDatabaseHandler struct {
	app application.MonitorDatabaseApplication
}

func (h *monitorDatabaseHandler) query(context *gin.Context) {
	var req types.MonitorDatabaseQueryRequest
	if context.ShouldBindQuery(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	total, data, err := h.app.Query(context.Request.Context(), &req)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildPageResponseSuccess(context, req.RequestPaging, total, data)
}

func (h *monitorDatabaseHandler) queryAll(context *gin.Context) {
	data, err := h.app.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorDatabaseHandler) create(context *gin.Context) {
	var db entity.MonitorDatabase
	if context.ShouldBindJSON(&db) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &db); err != nil {
		response.BuildResponseBadRequest(context, "创建数据源失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorDatabaseHandler) modify(context *gin.Context) {
	var db entity.MonitorDatabase
	if context.ShouldBindJSON(&db) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &db); err != nil {
		response.BuildResponseBadRequest(context, "更新数据源失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorDatabaseHandler 加载路由
func InitMonitorDatabaseHandler(app *application.Application) {
	handler := monitorDatabaseHandler{app.MonitorDatabase}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	server.RegisterRoute("/api/monitor/database", route)
}
