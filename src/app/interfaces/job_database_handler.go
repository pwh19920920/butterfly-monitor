package interfaces

import (
	"butterfly-monitor/src/app/application"
	"butterfly-monitor/src/app/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
)

type jobDatabaseHandler struct {
	jobDatabaseApp application.JobDatabaseApplication
}

// 查询
func (handler *jobDatabaseHandler) query(context *gin.Context) {
	var jobDatabaseQueryRequest types.JobDatabaseQueryRequest
	if context.ShouldBindQuery(&jobDatabaseQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.jobDatabaseApp.Query(&jobDatabaseQueryRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, jobDatabaseQueryRequest.RequestPaging, total, data)
}

// 创建
func (handler *jobDatabaseHandler) create(context *gin.Context) {
	var jobDatabaseCreateRequest types.JobDatabaseCreateRequest
	err := context.ShouldBindJSON(&jobDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.jobDatabaseApp.Create(&jobDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "创建数据源失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *jobDatabaseHandler) modify(context *gin.Context) {
	var jobDatabaseCreateRequest types.JobDatabaseCreateRequest
	err := context.ShouldBindJSON(&jobDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.jobDatabaseApp.Modify(&jobDatabaseCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "修改数据源失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitJobDatabaseHandler 加载路由
func InitJobDatabaseHandler(app *application.Application) {
	// 组件初始化
	handler := jobDatabaseHandler{app.JobDatabase}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	server.RegisterRoute("/api/job/database", route)
}
