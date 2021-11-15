package application

import (
	entity2 "butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/infrastructure/support"
	"butterfly-monitor/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/pwh19920920/butterfly-admin/config/sequence"
	"github.com/sirupsen/logrus"
	"time"
)

type MonitorTaskApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaOptionHandler
}

// Query 分页查询
func (application *MonitorTaskApplication) Query(request *types.MonitorTaskQueryRequest) (int64, []types.MonitorTaskQueryResponse, error) {
	total, data, err := application.repository.MonitorTaskRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.Select() happen error for", err)
		return total, nil, err
	}

	// 为空直接返回
	if len(data) == 0 {
		return total, []types.MonitorTaskQueryResponse{}, err
	}

	taskIds := make([]int64, 0)
	for _, item := range data {
		taskIds = append(taskIds, item.Id)
	}

	dashboardTasks, err := application.repository.MonitorDashboardTaskRepository.SelectByTaskIds(taskIds)
	if err != nil {
		return total, nil, errors.New("获取面板信息失败, 信息不匹配")
	}

	taskIdMap := make(map[int64][]string)
	for _, item := range dashboardTasks {
		list, ok := taskIdMap[item.TaskId]
		if !ok {
			list = make([]string, 0)
		}
		list = append(list, fmt.Sprintf("%v", item.DashboardId))
		taskIdMap[item.TaskId] = list
	}

	result := make([]types.MonitorTaskQueryResponse, 0)
	for _, item := range data {
		var execParam types.MonitorTaskExecParams
		_ = json.Unmarshal([]byte(item.ExecParams), &execParam)
		result = append(result, types.MonitorTaskQueryResponse{
			MonitorTask:    item,
			Dashboards:     taskIdMap[item.Id],
			TaskExecParams: execParam})
	}
	return total, result, err
}

// Create 创建数据源
func (application *MonitorTaskApplication) Create(request *types.MonitorTaskCreateRequest) error {
	monitorTask := request.MonitorTask
	execParams, _ := json.Marshal(request.TaskExecParams)
	monitorTask.ExecParams = string(execParams)
	monitorTask.Id = sequence.GetSequence().Generate().Int64()
	monitorTask.PreExecuteTime = &common.LocalTime{Time: time.Now()}

	// 转换
	dashboardIds, err := request.GetDashboardIds()
	if err != nil {
		return err
	}

	dashboards, err := application.repository.MonitorDashboardRepository.SelectByIds(dashboardIds)
	if err != nil || len(dashboards) != len(request.Dashboards) {
		return errors.New("获取面板信息失败, 信息不匹配")
	}

	// 往dashboard中push
	for _, dashboard := range dashboards {
		resp, err := application.grafanaHandler.AddPanel(dashboard.Uid, monitorTask)
		if err != nil || *resp.Status != "success" {
			return errors.New("创建grafana图表失败")
		}
	}

	monitorDashboardTasks := make([]entity2.MonitorDashboardTask, 0)
	for _, id := range dashboardIds {
		monitorDashboardTasks = append(monitorDashboardTasks, entity2.MonitorDashboardTask{
			BaseEntity:  common.BaseEntity{Id: application.sequence.Generate().Int64()},
			TaskId:      monitorTask.Id,
			DashboardId: id,
		})
	}

	// 错误记录
	err = application.repository.MonitorTaskRepository.Save(&monitorTask, monitorDashboardTasks)
	if err != nil {
		logrus.Error("MonitorTaskRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源， 原则上taskKey不允许修改
func (application *MonitorTaskApplication) Modify(request *types.MonitorTaskCreateRequest) error {
	// 转换
	dashboardIds, err := request.GetDashboardIds()
	if err != nil {
		return err
	}

	// 先校验一边新传入的有没有问题
	dashboards, err := application.repository.MonitorDashboardRepository.SelectByIds(dashboardIds)
	if err != nil || len(dashboards) != len(request.Dashboards) {
		return errors.New("获取面板信息失败, 信息不匹配")
	}

	dashboardIdMap := make(map[int64]bool, 0)
	for _, dashboardId := range dashboardIds {
		dashboardIdMap[dashboardId] = true
	}

	oldTask, err := application.repository.MonitorTaskRepository.GetById(request.Id)
	if err != nil || oldTask == nil {
		return errors.New("获取任务失败")
	}

	dashboardTasks, err := application.repository.MonitorDashboardTaskRepository.SelectByTaskIds([]int64{request.Id})
	if err != nil {
		return errors.New("获取面板信息失败, 信息不匹配")
	}

	oldDashboardIdMap := make(map[int64]bool, 0)
	for _, dashboardTask := range dashboardTasks {
		oldDashboardIdMap[dashboardTask.DashboardId] = true
	}

	// 遍历旧的, 不存在说明被移除, 存在说明是交集
	removeBoardIds := make([]int64, 0)
	intersectionBoardIds := make([]int64, 0)
	for _, dashboardTask := range dashboardTasks {
		_, ok := dashboardIdMap[dashboardTask.DashboardId]
		if !ok {
			removeBoardIds = append(removeBoardIds, dashboardTask.DashboardId)
		} else {
			intersectionBoardIds = append(intersectionBoardIds, dashboardTask.DashboardId)
		}
	}

	// 遍历新的, 不存在说明新增
	addBoardIds := make([]int64, 0)
	for _, dashboardId := range dashboardIds {
		_, ok := oldDashboardIdMap[dashboardId]
		if !ok {
			addBoardIds = append(addBoardIds, dashboardId)
		}
	}

	removeDashboardUIDs, err := application.getDashboardUIDs(removeBoardIds)
	if err != nil {
		return err
	}

	addDashboardUIDs, err := application.getDashboardUIDs(addBoardIds)
	if err != nil {
		return err
	}

	intersectionBoardUIDs, err := application.getDashboardUIDs(intersectionBoardIds)
	if err != nil {
		return err
	}

	err = application.grafanaHandler.ModifyDashBoardPanel(intersectionBoardUIDs, removeDashboardUIDs, addDashboardUIDs, request.MonitorTask)
	if err != nil {
		return err
	}

	monitorTask := request.MonitorTask
	execParams, _ := json.Marshal(request.TaskExecParams)
	monitorTask.ExecParams = string(execParams)

	monitorDashboardTasks := make([]entity2.MonitorDashboardTask, 0)
	for _, id := range dashboardIds {
		monitorDashboardTasks = append(monitorDashboardTasks, entity2.MonitorDashboardTask{
			BaseEntity:  common.BaseEntity{Id: application.sequence.Generate().Int64()},
			TaskId:      monitorTask.Id,
			DashboardId: id,
		})
	}

	err = application.repository.MonitorTaskRepository.UpdateById(monitorTask.Id, &entity2.MonitorTask{
		TaskName:   monitorTask.TaskName,
		TimeSpan:   monitorTask.TimeSpan,
		ExecParams: monitorTask.ExecParams,
		TaskType:   monitorTask.TaskType,
		Command:    monitorTask.Command,
	}, monitorDashboardTasks)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorTaskApplication) getDashboardUIDs(ids []int64) ([]string, error) {
	dashboards, err := application.repository.MonitorDashboardRepository.SelectByIds(ids)
	if err != nil {
		return nil, err
	}

	dashboardUIDs := make([]string, 0)
	for _, dashboard := range dashboards {
		dashboardUIDs = append(dashboardUIDs, dashboard.Uid)
	}
	return dashboardUIDs, nil
}

func (application *MonitorTaskApplication) ModifyTaskStatus(taskId int64, status entity2.MonitorTaskStatus) error {
	err := application.repository.MonitorTaskRepository.UpdateTaskStatusById(taskId, status)
	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateTaskStatusById() happen error", err)
	}
	return err
}

func (application *MonitorTaskApplication) ModifySampled(taskId int64, status entity2.MonitorSampledStatus) error {
	oldTask, err := application.repository.MonitorTaskRepository.GetById(taskId)
	if err != nil || oldTask == nil {
		return errors.New("获取任务失败")
	}

	dashboardTasks, err := application.repository.MonitorDashboardTaskRepository.SelectByTaskIds([]int64{taskId})
	if err != nil {
		return errors.New("获取面板信息失败, 信息不匹配")
	}

	dashboardIds := make([]int64, 0)
	for _, dashboardTask := range dashboardTasks {
		dashboardIds = append(dashboardIds, dashboardTask.DashboardId)
	}

	dashboardUIDs, err := application.getDashboardUIDs(dashboardIds)
	if err != nil {
		return err
	}

	oldTask.Sampled = status
	err = application.grafanaHandler.ModifySampleTarget(dashboardUIDs, oldTask)
	if err != nil {
		return err
	}

	err = application.repository.MonitorTaskRepository.UpdateSampledById(taskId, status)
	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateSampledById() happen error", err)
	}
	return err
}

func (application *MonitorTaskApplication) ModifyAlertStatus(taskId int64, status entity2.MonitorAlertStatus) error {
	err := application.repository.MonitorTaskRepository.UpdateAlertStatusById(taskId, status)
	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateById() happen error", err)
	}
	return err
}