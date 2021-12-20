package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type alertConfHandler struct {
	alertConfApp application.AlertConfApplication
}

// 查询
func (handler *alertConfHandler) query(context *gin.Context) {
	var alertConfQueryRequest types.AlertConfQueryRequest
	if context.ShouldBindQuery(&alertConfQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	total, data, err := handler.alertConfApp.Query(&alertConfQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, alertConfQueryRequest.RequestPaging, total, data)
}

// 创建
func (handler *alertConfHandler) create(context *gin.Context) {
	var alertConfCreateRequest types.AlertConfCreateRequest
	err := context.ShouldBindJSON(&alertConfCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.alertConfApp.Create(&alertConfCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "创建报警配置失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *alertConfHandler) modify(context *gin.Context) {
	var alertConfModifyRequest types.AlertConfModifyRequest
	err := context.ShouldBindJSON(&alertConfModifyRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.alertConfApp.Modify(&alertConfModifyRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "修改报警配置失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitAlertConfHandler 加载路由
func InitAlertConfHandler(app *application.Application) {
	// 组件初始化
	handler := alertConfHandler{app.AlertConfApp}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	server.RegisterRoute("/api/alert/conf", route)
}
