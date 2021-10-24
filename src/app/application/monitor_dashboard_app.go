package application

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/types"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/src/app/config/sequence"
	"github.com/sirupsen/logrus"
)

type MonitorDashboardApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
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
	err := application.repository.MonitorDashboardRepository.Save(&monitorDashboard)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改
func (application *MonitorDashboardApplication) Modify(request *types.MonitorDashboardCreateRequest) error {
	monitorDashboard := request.MonitorDashboard
	err := application.repository.MonitorDashboardRepository.UpdateById(monitorDashboard.Id, &monitorDashboard)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDashboardRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorDashboardApplication) SelectAll() ([]entity.MonitorDashboard, error) {
	return application.repository.MonitorDashboardRepository.SelectSimpleAll()
}
