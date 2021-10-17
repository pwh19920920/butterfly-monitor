package interfaces

import (
	"butterfly-monitor/src/app/application"
	"butterfly-monitor/src/app/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type monitorDatabaseHandler struct {
	monitorDatabaseApp application.MonitorDatabaseApplication
}

// 查询
func (handler *monitorDatabaseHandler) query(context *gin.Context) {
	var monitorDatabaseQueryRequest types.MonitorDatabaseQueryRequest
	if context.ShouldBindQuery(&monitorDatabaseQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.monitorDatabaseApp.Query(&monitorDatabaseQueryRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, monitorDatabaseQueryRequest.RequestPaging, total, data)
}

// 创建
func (handler *monitorDatabaseHandler) create(context *gin.Context) {
	var monitorDatabaseCreateRequest types.MonitorDatabaseCreateRequest
	err := context.ShouldBindJSON(&monitorDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorDatabaseApp.Create(&monitorDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "创建数据源失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorDatabaseHandler) modify(context *gin.Context) {
	var monitorDatabaseCreateRequest types.MonitorDatabaseCreateRequest
	err := context.ShouldBindJSON(&monitorDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorDatabaseApp.Modify(&monitorDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "修改数据源失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorDatabaseHandler 加载路由
func InitMonitorDatabaseHandler(app *application.Application) {
	// 组件初始化
	handler := monitorDatabaseHandler{app.MonitorDatabase}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	server.RegisterRoute("/api/monitor/database", route)
}
