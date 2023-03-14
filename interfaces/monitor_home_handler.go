package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type monitorHomeHandler struct {
	monitorTaskApp      application.MonitorTaskApplication
	monitorEventApp     application.MonitorTaskEventApplication
	monitorDashboardApp application.MonitorDashboardApplication
	monitorDatabaseApp  application.MonitorDatabaseApplication
}

// 修改
func (handler *monitorHomeHandler) homeCount(context *gin.Context) {
	taskCount, err := handler.monitorTaskApp.Count()
	if err != nil {
		response.BuildResponseBadRequest(context, "读取失败:"+err.Error())
		return
	}

	eventCount, err := handler.monitorEventApp.Count()
	if err != nil {
		response.BuildResponseBadRequest(context, "读取失败:"+err.Error())
		return
	}

	dashboardCount, err := handler.monitorDashboardApp.Count()
	if err != nil {
		response.BuildResponseBadRequest(context, "读取失败:"+err.Error())
		return
	}

	databaseCount, err := handler.monitorDatabaseApp.Count()
	if err != nil {
		response.BuildResponseBadRequest(context, "读取失败:"+err.Error())
		return
	}

	// 输出成功数据
	response.BuildResponseSuccess(context, types.MonitorTaskHomeCountResponse{
		TaskCount:      taskCount,
		EventCount:     eventCount,
		DashboardCount: dashboardCount,
		DatabaseCount:  databaseCount,
	})
}

// InitMonitorHomeHandler 加载路由
func InitMonitorHomeHandler(app *application.Application) {
	// 组件初始化
	handler := monitorHomeHandler{
		monitorTaskApp:      app.MonitorTaskApp,
		monitorEventApp:     app.MonitorTaskEventApp,
		monitorDashboardApp: app.MonitorDashboardApp,
		monitorDatabaseApp:  app.MonitorDatabaseApp,
	}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.homeCount})
	server.RegisterRoute("/api/monitor/homeCount", route)
}
