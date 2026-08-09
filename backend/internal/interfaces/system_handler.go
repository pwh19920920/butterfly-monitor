package interfaces

import (
	"butterfly-monitor/internal/application"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorSystemHandler struct {
	app *application.Application
}

// systemMetrics 返回本平台服务器自身的硬件/软件性能指标快照
func (h *monitorSystemHandler) systemMetrics(c *gin.Context) {
	data := h.app.System.Metrics()
	response.BuildResponseSuccess(c, data)
}

func InitMonitorSystemHandler(app *application.Application) {
	h := monitorSystemHandler{app: app}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: h.systemMetrics})
	server.RegisterRoute("/api/monitor/systemMetrics", route)
}
