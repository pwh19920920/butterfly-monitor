package interfaces

import (
	"strconv"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/job"
	"butterfly-monitor/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

type monitorTaskHandler struct {
	app application.MonitorTaskApplication
	job *job.MonitorDataCollectJob
}

func (h *monitorTaskHandler) query(context *gin.Context) {
	var req types.MonitorTaskQueryRequest
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

// getById 查询任务详情，返回带 taskAlert/checkParams 的完整结构，供编辑态回显
func (h *monitorTaskHandler) getById(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "id 参数有误")
		return
	}
	data, err := h.app.GetById(context.Request.Context(), id)
	if err != nil || data == nil {
		response.BuildResponseBadRequest(context, "任务不存在")
		return
	}
	response.BuildResponseSuccess(context, data)
}

func (h *monitorTaskHandler) create(context *gin.Context) {
	var req types.MonitorTaskCreateRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Create(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "创建任务失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskHandler) modify(context *gin.Context) {
	var req types.MonitorTaskCreateRequest
	err := context.ShouldBindJSON(&req)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.Modify(context.Request.Context(), &req); err != nil {
		response.BuildResponseBadRequest(context, "更新任务失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskHandler) modifyAlertStatus(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	statusInt, err := strconv.ParseInt(context.Param("status"), 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.ModifyAlertStatus(context.Request.Context(), id, entity.MonitorAlertStatus(statusInt)); err != nil {
		response.BuildResponseBadRequest(context, "更新告警开关失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskHandler) modifyTaskStatus(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	statusInt, err := strconv.ParseInt(context.Param("status"), 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.ModifyTaskStatus(context.Request.Context(), id, entity.MonitorTaskStatus(statusInt)); err != nil {
		response.BuildResponseBadRequest(context, "更新任务开关失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

func (h *monitorTaskHandler) modifySampled(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	statusInt, err := strconv.ParseInt(context.Param("status"), 10, 32)
	if err != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if err := h.app.ModifySampled(context.Request.Context(), id, entity.MonitorSampledStatus(statusInt)); err != nil {
		response.BuildResponseBadRequest(context, "更新样本开关失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// dataPush 外部数据推送（免登录）
// body 传 taskKey，handler 转换为 taskId 后交给 job
func (h *monitorTaskHandler) dataPush(context *gin.Context) {
	var req types.MonitorTaskPushDataRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	if req.TaskKey == "" {
		response.BuildResponseBadRequest(context, "taskKey不能为空")
		return
	}
	logger.InfoFormat(context, "收到dataPush请求, taskKey: %s", req.TaskKey)
	task, err := h.app.GetByTaskKey(context.Request.Context(), req.TaskKey)
	if err != nil {
		logger.ErrorFormat(context, "dataPush查询任务失败, taskKey: %s: %v", req.TaskKey, err)
		response.BuildResponseBadRequest(context, "请求发送错误")
		return
	}
	if task == nil {
		response.BuildResponseBadRequest(context, "任务不存在")
		return
	}
	req.TaskId = task.Id
	if err := h.job.DataPush(context.Request.Context(), &req); err != nil {
		logger.ErrorFormat(context, "dataPush失败, taskKey: %s: %v", req.TaskKey, err)
		response.BuildResponseBadRequest(context, err.Error())
		return
	}
	response.BuildResponseSuccess(context, "ok")
}

// PreviewAggregate 聚合预览：临时执行多行查询，返回结果列名，供前端勾选 label/value 维度（不落库）
func (h *monitorTaskHandler) previewAggregate(context *gin.Context) {
	var req types.MonitorTaskPreviewRequest
	if context.ShouldBindJSON(&req) != nil {
		response.BuildResponseBadRequest(context, "请求参数有误")
		return
	}
	data, err := h.app.PreviewAggregate(context.Request.Context(), &req)
	if err != nil {
		response.BuildResponseBadRequest(context, "预览查询失败: "+err.Error())
		return
	}
	response.BuildResponseSuccess(context, data)
}
func InitMonitorTaskHandler(app *application.Application, timerJob *job.Job) {
	handler := monitorTaskHandler{app: app.MonitorTask, job: &timerJob.DataCollect}
	var route []server.RouteInfo
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "", HandlerFunc: handler.query})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpGet, Path: "/:id", HandlerFunc: handler.getById})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "", HandlerFunc: handler.create})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "", HandlerFunc: handler.modify})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/alertStatus/:id/:status", HandlerFunc: handler.modifyAlertStatus})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/taskStatus/:id/:status", HandlerFunc: handler.modifyTaskStatus})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPut, Path: "/sampled/:id/:status", HandlerFunc: handler.modifySampled})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/dataPush", HandlerFunc: handler.dataPush})
	route = append(route, server.RouteInfo{HttpMethod: server.HttpPost, Path: "/previewAggregate", HandlerFunc: handler.previewAggregate})
	server.RegisterRoute("/api/monitor/task", route)
}
