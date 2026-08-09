package interfaces

import (
	"strconv"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorDashboardHandler struct {
	app application.MonitorDashboardApplication
}

func (h *monitorDashboardHandler) query(context *gin.Context) {
	var req types.MonitorDashboardQueryRequest
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

func (h *monitorDashboardHandler) queryAll(context *gin.Context) {
	data, err := h.app.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorDashboardHandler) create(context *gin.Context) {
	var dashboard entity.MonitorDashboard
	if context.ShouldBindJSON(&dashboard) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &dashboard); err != nil {
		response.BuildResponseBadRequest(context, "创建面板失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorDashboardHandler) modify(context *gin.Context) {
	var dashboard entity.MonitorDashboard
	if context.ShouldBindJSON(&dashboard) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &dashboard); err != nil {
		response.BuildResponseBadRequest(context, "更新面板失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorDashboardHandler) queryTasks(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	data, err := h.app.QueryTasks(context.Request.Context(), id)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorDashboardHandler) taskSort(context *gin.Context) {
	var req types.MonitorDashboardTaskSortRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.TaskSort(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "排序失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorDashboardHandler 加载路由
func InitMonitorDashboardHandler(app *application.Application) {
	handler := monitorDashboardHandler{app.MonitorDashboard}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/task/:id", HandlerFunc: handler.queryTasks})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/taskSort", HandlerFunc: handler.taskSort})
	server.RegisterRoute("/api/monitor/dashboard", route)
}
