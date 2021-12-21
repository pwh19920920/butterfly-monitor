package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	adminApp "github.com/pwh19920920/butterfly-admin/application"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
	"strconv"
)

type monitorTaskEventHandler struct {
	monitorTaskEventApp application.MonitorTaskEventApplication
	adminApp            *adminApp.Application
}

func (handler *monitorTaskEventHandler) query(context *gin.Context) {
	var monitorTaskEventQueryRequest types.MonitorTaskEventQueryRequest
	if context.ShouldBindQuery(&monitorTaskEventQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.monitorTaskEventApp.Query(&monitorTaskEventQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "查询失败:"+err.Error())
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, monitorTaskEventQueryRequest.RequestPaging, total, data)
}

func (handler *monitorTaskEventHandler) deal(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// 不被忽略则判断是否有令牌
	token := context.GetHeader(handler.adminApp.Login.GetHeaderName())
	ticket, err := handler.adminApp.Login.CheckAndGetTicket(token)
	if err != nil {
		response.BuildResponseBadRequest(context, "用户信息读取失败")
		return
	}

	var monitorTaskEventProcessRequest types.MonitorTaskEventProcessRequest
	if context.ShouldBindJSON(&monitorTaskEventProcessRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	monitorTaskEventProcessRequest.DealUser = &ticket.UserId
	if err = handler.monitorTaskEventApp.DealEvent(id, &monitorTaskEventProcessRequest); err != nil {
		response.BuildResponseSysErr(context, "查询失败:"+err.Error())
		return
	}

	// 输出
	response.BuildResponseSuccess(context, "OK")
}

func (handler *monitorTaskEventHandler) complete(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	var monitorTaskEventProcessRequest types.MonitorTaskEventProcessRequest
	if context.ShouldBindJSON(&monitorTaskEventProcessRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	if err = handler.monitorTaskEventApp.CompleteEvent(id, &monitorTaskEventProcessRequest); err != nil {
		response.BuildResponseSysErr(context, "查询失败:"+err.Error())
		return
	}

	// 输出
	response.BuildResponseSuccess(context, "OK")
}

// InitMonitorTaskEventHandler 加载路由
func InitMonitorTaskEventHandler(app *application.Application) {
	// 组件初始化
	handler := monitorTaskEventHandler{app.MonitorTaskEventApp, app.AdminApp}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/deal/:id", HandlerFunc: handler.deal})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/complete/:id", HandlerFunc: handler.complete})
	server.RegisterRoute("/api/monitor/task/event", route)
}
