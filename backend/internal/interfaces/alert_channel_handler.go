package interfaces

import (
	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type alertChannelHandler struct {
	app application.AlertChannelApplication
}

func (h *alertChannelHandler) query(context *gin.Context) {
	var req types.AlertChannelQueryRequest
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

func (h *alertChannelHandler) queryAll(context *gin.Context) {
	data, err := h.app.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *alertChannelHandler) handlers(context *gin.Context) {
	response.BuildResponseSuccess(context, h.app.Handlers(context.Request.Context()))
}

func (h *alertChannelHandler) create(context *gin.Context) {
	var req types.AlertChannelSaveRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &req.AlertChannel, req.TestParams); err != nil {
		response.BuildResponseBadRequest(context, "创建通道失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *alertChannelHandler) modify(context *gin.Context) {
	var req types.AlertChannelSaveRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &req.AlertChannel, req.TestParams); err != nil {
		response.BuildResponseBadRequest(context, "更新通道失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitAlertChannelHandler 加载路由
func InitAlertChannelHandler(app *application.Application) {
	handler := alertChannelHandler{app.AlertChannel}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/handlers", HandlerFunc: handler.handlers})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	server.RegisterRoute("/api/alert/channel", route)
}
