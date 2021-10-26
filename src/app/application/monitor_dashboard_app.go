package application

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/infrastructure/support"
	"butterfly-monitor/src/app/types"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/src/app/config/sequence"
	"github.com/sirupsen/logrus"
)

type MonitorDashboardApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaOptionHandler
}

// Query 分页查询
func (application *MonitorDashboardApplication) Query(request *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error) {
	total, data, err := application.repository.MonitorDashboardRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.Select() happen error for", err)
	}
	return total, data, err
}

// Create 创建
func (application *MonitorDashboardApplication) Create(request *types.MonitorDashboardCreateRequest) error {
	monitorDashboard := request.MonitorDashboard
	monitorDashboard.Id = sequence.GetSequence().Generate().Int64()

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
