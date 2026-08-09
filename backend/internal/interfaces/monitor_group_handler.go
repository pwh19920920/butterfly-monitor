package interfaces

import (
	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorGroupHandler struct {
	app application.MonitorGroupApplication
}

func (h *monitorGroupHandler) query(context *gin.Context) {
	var req types.MonitorGroupQueryRequest
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

func (h *monitorGroupHandler) queryAll(context *gin.Context) {
	data, err := h.app.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorGroupHandler) create(context *gin.Context) {
	var group entity.MonitorGroup
	if context.ShouldBindJSON(&group) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &group); err != nil {
		response.BuildResponseBadRequest(context, "创建分组失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorGroupHandler) modify(context *gin.Context) {
	var group entity.MonitorGroup
	if context.ShouldBindJSON(&group) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &group); err != nil {
		response.BuildResponseBadRequest(context, "更新分组失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorGroupHandler 加载路由
func InitMonitorGroupHandler(app *application.Application) {
	handler := monitorGroupHandler{app.MonitorGroup}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	server.RegisterRoute("/api/monitor/group", route)
}
