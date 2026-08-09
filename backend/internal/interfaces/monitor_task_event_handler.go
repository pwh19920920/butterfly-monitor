package interfaces

import (
	"strconv"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorTaskEventHandler struct {
	app application.MonitorTaskEventApplication
}

func (h *monitorTaskEventHandler) query(context *gin.Context) {
	var req types.MonitorTaskEventQueryRequest
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

func (h *monitorTaskEventHandler) deal(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	var req types.MonitorTaskEventProcessRequest
	_ = context.ShouldBindJSON(&req)

	// 从 token 取当前用户
	ticket, err := GetUserTicket(context)
	if err == nil && ticket != nil {
		uid := new(int64)
		*uid = ticket.UserId
		req.DealUser = uid
	}

	if err := h.app.DealEvent(context.Request.Context(), id, &req); err != nil {
		response.BuildResponseBadRequest(context, "处理事件失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskEventHandler) complete(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	var req types.MonitorTaskEventProcessRequest
	_ = context.ShouldBindJSON(&req)

	ticket, err := GetUserTicket(context)
	if err == nil && ticket != nil {
		uid := new(int64)
		*uid = ticket.UserId
		req.DealUser = uid
	}

	if err := h.app.CompleteEvent(context.Request.Context(), id, &req); err != nil {
		response.BuildResponseBadRequest(context, "完成事件失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskEventHandler) ignore(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	var req types.MonitorTaskEventProcessRequest
	_ = context.ShouldBindJSON(&req)

	// 从 token 取当前用户
	ticket, err := GetUserTicket(context)
	if err == nil && ticket != nil {
		uid := new(int64)
		*uid = ticket.UserId
		req.DealUser = uid
	}

	if err := h.app.IgnoreEvent(context.Request.Context(), id, &req); err != nil {
		response.BuildResponseBadRequest(context, "忽略事件失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorTaskEventHandler 加载路由
func InitMonitorTaskEventHandler(app *application.Application) {
	handler := monitorTaskEventHandler{app.MonitorTaskEvent}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/deal/:id", HandlerFunc: handler.deal})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/complete/:id", HandlerFunc: handler.complete})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/ignore/:id", HandlerFunc: handler.ignore})
	server.RegisterRoute("/api/monitor/task/event", route)
}
