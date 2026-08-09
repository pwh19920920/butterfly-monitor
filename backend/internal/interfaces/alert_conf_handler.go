package interfaces

import (
	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type alertConfHandler struct {
	app application.AlertConfApplication
}

func (h *alertConfHandler) query(context *gin.Context) {
	var req types.AlertConfQueryRequest
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

func (h *alertConfHandler) create(context *gin.Context) {
	var conf entity.AlertConf
	if context.ShouldBindJSON(&conf) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &conf); err != nil {
		response.BuildResponseBadRequest(context, "创建配置失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *alertConfHandler) modify(context *gin.Context) {
	var conf entity.AlertConf
	if context.ShouldBindJSON(&conf) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &conf); err != nil {
		response.BuildResponseBadRequest(context, "更新配置失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// InitAlertConfHandler 加载路由
func InitAlertConfHandler(app *application.Application) {
	handler := alertConfHandler{app.AlertConf}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	server.RegisterRoute("/api/alert/conf", route)
}
