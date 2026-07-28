package interfaces

import (
	"strconv"

	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type alertGroupHandler struct {
	app application.AlertGroupApplication
}

func (h *alertGroupHandler) query(context *gin.Context) {
	var req types.AlertGroupQueryRequest
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

func (h *alertGroupHandler) queryAll(context *gin.Context) {
	data, err := h.app.QueryAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *alertGroupHandler) queryGroupUser(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	data, err := h.app.QueryGroupUsers(context.Request.Context(), id)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *alertGroupHandler) create(context *gin.Context) {
	var req types.AlertGroupSaveRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "创建分组失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *alertGroupHandler) modify(context *gin.Context) {
	var req types.AlertGroupSaveRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "更新分组失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitAlertGroupHandler 加载路由
func InitAlertGroupHandler(app *application.Application) {
	handler := alertGroupHandler{app.AlertGroup}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/groupUser/:id", HandlerFunc: handler.queryGroupUser})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	server.RegisterRoute("/api/alert/group", route)
}
