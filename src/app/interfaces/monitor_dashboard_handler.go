package interfaces

import (
	"butterfly-monitor/src/app/application"
	"butterfly-monitor/src/app/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type monitorDashboardHandler struct {
	monitorDashboardApp application.MonitorDashboardApplication
}

// 查询
func (handler *monitorDashboardHandler) query(context *gin.Context) {
	var monitorDashboardQueryRequest types.MonitorDashboardQueryRequest
	if context.ShouldBindQuery(&monitorDashboardQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.monitorDashboardApp.Query(&monitorDashboardQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, monitorDashboardQueryRequest.RequestPaging, total, data)
}

// 创建
func (handler *monitorDashboardHandler) create(context *gin.Context) {
	var monitorDashboardCreateRequest types.MonitorDashboardCreateRequest
	err := context.ShouldBindJSON(&monitorDashboardCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorDashboardApp.Create(&monitorDashboardCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "创建失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorDashboardHandler) modify(context *gin.Context) {
	var monitorDashboardCreateRequest types.MonitorDashboardCreateRequest
	err := context.ShouldBindJSON(&monitorDashboardCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorDashboardApp.Modify(&monitorDashboardCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "修改失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 查询
func (handler *monitorDashboardHandler) selectAll(context *gin.Context) {
	// option
	data, err := handler.monitorDashboardApp.SelectAll()
	if err != nil {
		response.BuildResponseSysErr(context, "查询失败")
		return
	}

	response.BuildResponseSuccess(context, data)
}

// InitMonitorDashboardHandler 加载路由
func InitMonitorDashboardHandler(app *application.Application) {
	// 组件初始化
	handler := monitorDashboardHandler{app.MonitorDashboard}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.selectAll})
	server.RegisterRoute("/api/monitor/dashboard", route)
}
