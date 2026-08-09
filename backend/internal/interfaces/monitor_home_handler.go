package interfaces

import (
	"math/rand"

	"butterfly-monitor/internal/application"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorHomeHandler struct {
	app *application.Application
}

func (h *monitorHomeHandler) homeCount(c *gin.Context) {
	data, err := h.app.HomeCount(c.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(c, "统计失败")
		return
	}
	response.BuildResponseSuccess(c, data)
}

func InitMonitorHomeHandler(app *application.Application) {
	h := monitorHomeHandler{app: app}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: h.homeCount})
	server.RegisterRoute("/api/monitor/homeCount", route)
}

func InitMonitorHealthHandler(_ *application.Application) {
	health := func(c *gin.Context) {
		c.String(200, "OK")
	}
	test := func(c *gin.Context) {
		response.BuildResponseSuccess(c, rand.Float64())
	}
	var healthRoute []server.RouteInfo
	healthRoute = append(healthRoute, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: health})
	server.RegisterRoute("/api/health", healthRoute)

	var testRoute []server.RouteInfo
	testRoute = append(testRoute, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: test})
	server.RegisterRoute("/api/monitor/test", testRoute)
}
