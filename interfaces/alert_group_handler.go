package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
	"strconv"
)

type alertGroupHandler struct {
	alertGroupApp application.AlertGroupApplication
}

// 查询
func (handler *alertGroupHandler) query(context *gin.Context) {
	var alertGroupQueryRequest types.AlertGroupQueryRequest
	if context.ShouldBindQuery(&alertGroupQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	total, data, err := handler.alertGroupApp.Query(&alertGroupQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误："+err.Error())
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, alertGroupQueryRequest.RequestPaging, total, data)
}

// 查询全部
func (handler *alertGroupHandler) queryAll(context *gin.Context) {
	data, err := handler.alertGroupApp.QueryAll()
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误："+err.Error())
		return
	}

	// 输出
	response.BuildResponseSuccess(context, data)
}

// 查询底下的用户
func (handler *alertGroupHandler) queryGroupUserIdsByGroupId(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	data, err := handler.alertGroupApp.QueryGroupUserIdsByGroupId(id)
	if err != nil {
		response.BuildResponseSysErr(context, "请求查询错误："+err.Error())
		return
	}

	// 数字转为字符串，避免产生前端处理问题
	result := make([]string, 0)
	for _, item := range data {
		result = append(result, strconv.FormatInt(item, 10))
	}

	// 输出
	response.BuildResponseSuccess(context, result)
}

// 创建
func (handler *alertGroupHandler) create(context *gin.Context) {
	var alertGroupCreateRequest types.AlertGroupCreateRequest
	err := context.ShouldBindJSON(&alertGroupCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.alertGroupApp.Save(&alertGroupCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "创建分组失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *alertGroupHandler) modify(context *gin.Context) {
	var alertGroupModifyRequest types.AlertGroupModifyRequest
	err := context.ShouldBindJSON(&alertGroupModifyRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.alertGroupApp.Modify(&alertGroupModifyRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "修改分组失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitAlertGroupHandler 加载路由
func InitAlertGroupHandler(app *application.Application) {
	// 组件初始化
	handler := alertGroupHandler{app.AlertGroupApp}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/all", HandlerFunc: handler.queryAll})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/groupUser/:id", HandlerFunc: handler.queryGroupUserIdsByGroupId})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	server.RegisterRoute("/api/alert/group", route)
}
