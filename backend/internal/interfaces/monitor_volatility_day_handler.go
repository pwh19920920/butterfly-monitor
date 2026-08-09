package interfaces

import (
	"strconv"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorVolatilityDayHandler struct {
	app application.MonitorVolatilityDayApplication
}

func (h *monitorVolatilityDayHandler) queryAll(context *gin.Context) {
	data, err := h.app.SelectAll(context.Request.Context())
	if err != nil {
		response.BuildResponseBadRequest(context, "查询失败")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorVolatilityDayHandler) batchCreate(context *gin.Context) {
	var req types.MonitorVolatilityDayBatchCreateRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误:"+err.Error())
		return
	}
	if err := h.app.BatchCreate(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "批量添加失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorVolatilityDayHandler) modify(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	var day entity.MonitorVolatilityDay
	if context.ShouldBindJSON(&day) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	// 以 URL 路径 id 为准，不信任 body 里的 id
	day.Id = id
	if err := h.app.Modify(context.Request.Context(), &day); err != nil {
		response.BuildResponseBadRequest(context, "更新失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorVolatilityDayHandler) delete(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Delete(context.Request.Context(), id); err != nil {
		response.BuildResponseBadRequest(context, "删除失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorVolatilityDayHandler 加载路由
func InitMonitorVolatilityDayHandler(app *application.Application) {
	handler := monitorVolatilityDayHandler{app.MonitorVolatilityDay}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/batch", HandlerFunc: handler.batchCreate})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/:id", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpDelete, Path: "/:id", HandlerFunc: handler.delete})
	server.RegisterRoute("/api/monitor/volatilityDay", route)
}
