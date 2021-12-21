package application

import (
	"butterfly-monitor/config/grafana"
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/infrastructure/support"
	"butterfly-monitor/types"
	"errors"
	"fmt"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
)

type MonitorDashboardApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaOptionHandler
	Grafana        *grafana.Config
}

// Query 分页查询
func (application *MonitorDashboardApplication) Query(request *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error) {
	total, data, err := application.repository.MonitorDashboardRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.Select() happen error for", err)
	}

	for index, item := range data {
		item.Url = application.Grafana.Addr + item.Url
		data[index] = item
	}
	return total, data, err
}

// Create 创建
func (application *MonitorDashboardApplication) Create(request *types.MonitorDashboardCreateRequest) error {
	monitorDashboard := request.MonitorDashboard
	monitorDashboard.Id = application.sequence.Generate().Int64()

	resp, err := application.grafanaHandler.CreateDashboard(monitorDashboard.Name)
	if err != nil || *resp.Status != "success" {
		logrus.Error("postDashBoard status is not success", err)
		return err
	}

	// 赋值
	monitorDashboard.Url = *resp.URL
	monitorDashboard.Slug = *resp.Slug
	monitorDashboard.Uid = *resp.UID
	monitorDashboard.BoardId = resp.ID

	err = application.repository.MonitorDashboardRepository.Save(&monitorDashboard)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改
func (application *MonitorDashboardApplication) Modify(request *types.MonitorDashboardCreateRequest) error {
	oldMonitorDashboard, err := application.repository.MonitorDashboardRepository.GetById(request.Id)
	if err != nil || oldMonitorDashboard == nil {
		logrus.Error("获取历史dashboard失败", oldMonitorDashboard.Id, err)
		return errors.New("获取历史dashboard失败")
	}

	resp, err := application.grafanaHandler.ModifyDashboardName(oldMonitorDashboard.Uid, request.Name)
	if err != nil || *resp.Status != "success" {
		logrus.Error("postDashBoard status is not success", err)
		return errors.New(fmt.Sprintf("postDashBoard status is not success, current status is %s", *resp.Status))
	}

	// 赋值
	monitorDashboard := request.MonitorDashboard
	monitorDashboard.Url = *resp.URL
	monitorDashboard.Slug = *resp.Slug
	monitorDashboard.Uid = *resp.UID
	monitorDashboard.BoardId = resp.ID

	err = application.repository.MonitorDashboardRepository.UpdateById(monitorDashboard.Id, &monitorDashboard)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorDashboardApplication) SelectAll() ([]entity.MonitorDashboard, error) {
	return application.repository.MonitorDashboardRepository.SelectSimpleAll()
}

func (application *MonitorDashboardApplication) SelectByDashboardId(dashboardId int64) ([]types.MonitorDashboardQueryTaskResponse, error) {
	monitorDashboardTasks, err := application.repository.MonitorDashboardTaskRepository.SelectByDashboardId(dashboardId)
	if err != nil {
		return nil, err
	}

	taskIds := make([]int64, 0)
	for _, task := range monitorDashboardTasks {
		taskIds = append(taskIds, task.TaskId)
	}

	taskMap, err := application.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)
	if err != nil {
		return nil, err
	}

	result := make([]types.MonitorDashboardQueryTaskResponse, 0)
	for _, task := range monitorDashboardTasks {
		result = append(result, types.MonitorDashboardQueryTaskResponse{
			MonitorDashboardTask: task,
			TaskKey:              taskMap[task.TaskId].TaskKey,
			TaskName:             taskMap[task.TaskId].TaskName},
		)
	}
	return result, nil
}

func (application *MonitorDashboardApplication) ModifyDashboardTaskSort(req *types.MonitorDashboardTaskModifyRequest) error {
	ids := make([]int64, 0)
	for _, item := range req.Data {
		ids = append(ids, item.Id)
	}

	monitorDashboardTasks, err := application.repository.MonitorDashboardTaskRepository.SelectByIds(ids)
	if err != nil {
		return err
	}

	if len(monitorDashboardTasks) != len(ids) {
		return errors.New("请求有误")
	}

	err = application.repository.MonitorDashboardTaskRepository.BatchModifySort(req.Data)
	if err != nil {
		return err
	}

	taskIds := make([]int64, 0)
	for _, dashboardTask := range req.Data {
		taskIds = append(taskIds, dashboardTask.TaskId)
	}

	monitorTaskMap, err := application.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)
	if err != nil {
		return err
	}

	dashboard, err := application.repository.MonitorDashboardRepository.GetById(req.Data[0].DashboardId)
	if err != nil || dashboard == nil {
		return errors.New("获取面板信息失败")
	}

	return application.grafanaHandler.ReSortDashboard(dashboard.Uid, taskIds, monitorTaskMap)
}
