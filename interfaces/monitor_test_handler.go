package interfaces

import (
	"butterfly-monitor/application"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
	"math/rand"
)

// 修改
func (handler *monitorTaskHandler) test(context *gin.Context) {
	response.BuildResponseSuccess(context, rand.Float64())
}

// InitMonitorTestHandler 加载路由
func InitMonitorTestHandler(app *application.Application) {
	// 组件初始化
	handler := monitorTaskHandler{}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.test})
	server.RegisterRoute("/api/monitor/test", route)
}
