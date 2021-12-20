package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type alertChannelHandler struct {
	alertChannelApp application.AlertChannelApplication
}

func (handler *alertChannelHandler) handlers(context *gin.Context) {
	handlersResp := handler.alertChannelApp.Handlers()

	// 输出
	response.BuildResponseSuccess(context, handlersResp)
}

// 查询
func (handler *alertChannelHandler) query(context *gin.Context) {
	var alertChannelQueryRequest types.AlertChannelQueryRequest
	if context.ShouldBindQuery(&alertChannelQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	total, data, err := handler.alertChannelApp.Query(&alertChannelQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, alertChannelQueryRequest.RequestPaging, total, data)
}

// 查询
func (handler *alertChannelHandler) queryAll(context *gin.Context) {
	data, err := handler.alertChannelApp.QueryAll()
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误")
		return
	}

	// 输出
	response.BuildResponseSuccess(context, data)
}

// 创建
func (handler *alertChannelHandler) create(context *gin.Context) {
	var alertChannelCreateRequest types.AlertChannelCreateRequest
	err := context.ShouldBindJSON(&alertChannelCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	if err := alertChannelCreateRequest.Validate(); err != nil {
		response.BuildResponseSysErr(context, "请求数据有误:"+err.Error())
		return
	}

	// option
	err = handler.alertChannelApp.Create(&alertChannelCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "创建失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *alertChannelHandler) modify(context *gin.Context) {
	var alertChannelModifyRequest types.AlertChannelModifyRequest
	err := context.ShouldBindJSON(&alertChannelModifyRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	if err = alertChannelModifyRequest.Validate(); err != nil {
		response.BuildResponseSysErr(context, "请求数据有误:"+err.Error())
		return
	}

	// option
	err = handler.alertChannelApp.Modify(&alertChannelModifyRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "修改报警配置失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitAlertChannelHandler 加载路由
func InitAlertChannelHandler(app *application.Application) {
	// 组件初始化
	handler := alertChannelHandler{app.AlertChannelApp}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/handlers", HandlerFunc: handler.handlers})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	server.RegisterRoute("/api/alert/channel", route)
}
