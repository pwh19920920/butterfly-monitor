package interfaces

import (
	"butterfly-monitor/application"
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/job"
	"butterfly-monitor/types"
	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/response"
	"github.com/pwh19920920/butterfly/server"
	"strconv"
)

type monitorTaskHandler struct {
	monitorTaskApp application.MonitorTaskApplication
	monitorExecApp job.MonitorDataCollectJob
}

// 查询
func (handler *monitorTaskHandler) query(context *gin.Context) {
	var monitorTaskQueryRequest types.MonitorTaskQueryRequest
	if context.ShouldBindQuery(&monitorTaskQueryRequest) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	total, data, err := handler.monitorTaskApp.Query(&monitorTaskQueryRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "请求发送错误")
		return
	}

	// 输出
	response.BuildPageResponseSuccess(context, monitorTaskQueryRequest.RequestPaging, total, data)
}

// 创建
func (handler *monitorTaskHandler) create(context *gin.Context) {
	var monitorTaskCreateRequest types.MonitorTaskCreateRequest
	err := context.ShouldBindJSON(&monitorTaskCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	err = monitorTaskCreateRequest.ValidateForCreate()
	if err != nil {
		response.BuildResponseBadRequest(context, err.Error())
		return
	}

	// option
	err = handler.monitorTaskApp.Create(&monitorTaskCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "创建任务失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

func (handler *monitorTaskHandler) execForTimeRange(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	var monitorTaskExecForRangeRequest types.MonitorTaskExecForRangeRequest
	err = context.ShouldBindJSON(&monitorTaskExecForRangeRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	err = monitorTaskExecForRangeRequest.ValidateForExec()
	if err != nil {
		response.BuildResponseBadRequest(context, err.Error())
		return
	}

	// option
	err = handler.monitorExecApp.ExecDataCollectForTimeRange(id, &monitorTaskExecForRangeRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "执行操作失败")
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorTaskHandler) modify(context *gin.Context) {
	var monitorTaskCreateRequest types.MonitorTaskCreateRequest
	err := context.ShouldBindJSON(&monitorTaskCreateRequest)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorTaskApp.Modify(&monitorTaskCreateRequest)
	if err != nil {
		response.BuildResponseSysErr(context, "修改任务失败: "+err.Error())
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorTaskHandler) modifyTaskStatus(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	taskStatusStr := context.Param("status")
	taskStatus, err := strconv.ParseInt(taskStatusStr, 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorTaskApp.ModifyTaskStatus(id, entity.MonitorTaskStatus(taskStatus))
	if err != nil {
		response.BuildResponseBadRequest(context, "修改任务状态失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorTaskHandler) modifySampled(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	taskStatusStr := context.Param("status")
	taskStatus, err := strconv.ParseInt(taskStatusStr, 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorTaskApp.ModifySampled(id, entity.MonitorSampledStatus(taskStatus))
	if err != nil {
		response.BuildResponseBadRequest(context, "修改收集状态失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// 修改
func (handler *monitorTaskHandler) modifyAlertStatus(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	alertStatusStr := context.Param("status")
	alertStatus, err := strconv.ParseInt(alertStatusStr, 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}

	// option
	err = handler.monitorTaskApp.ModifyAlertStatus(id, entity.MonitorAlertStatus(alertStatus))
	if err != nil {
		response.BuildResponseSysErr(context, "修改任务状态失败")
		return
	}

	response.BuildResponseSuccess(context, "ok")
}

// InitMonitorTaskHandler 加载路由
func InitMonitorTaskHandler(app *application.Application, timer *job.Job) {
	// 组件初始化
	handler := monitorTaskHandler{monitorTaskApp: app.MonitorTaskApp, monitorExecApp: timer.MonitorDataCollectJob}

	// 路由初始化
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/alertStatus/:id/:status", HandlerFunc: handler.modifyAlertStatus})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/taskStatus/:id/:status", HandlerFunc: handler.modifyTaskStatus})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/sampled/:id/:status", HandlerFunc: handler.modifySampled})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/execForTimeRange/:id", HandlerFunc: handler.execForTimeRange})
	server.RegisterRoute("/api/monitor/task", route)
}
