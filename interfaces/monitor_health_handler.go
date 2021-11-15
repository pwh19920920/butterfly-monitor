package interfaces

import (
	"butterfly-monitor/application"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type monitorHealthHandler struct {
}

// 修改
func (handler *monitorHealthHandler) health(context *gin.Context) {
	response.BuildResponseSuccess(context, "OK")
}

// InitMonitorHealthHandler 加载路由
func InitMonitorHealthHandler(app *application.Application) {
	// 组件初始化
	handler := monitorHealthHandler{}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.health})
	server.RegisterRoute("/api/health", route)
}
