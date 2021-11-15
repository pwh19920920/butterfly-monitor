package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/config/sequence"
	"github.com/sirupsen/logrus"
)

type MonitorDatabaseApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// Query 分页查询
func (application *MonitorDatabaseApplication) Query(request *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error) {
	total, data, err := application.repository.MonitorDatabaseRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.Select() happen error for", err)
	}

	return total, data, err
}

// Create 创建数据源
func (application *MonitorDatabaseApplication) Create(request *types.MonitorDatabaseCreateRequest) error {
	monitorDatabase := request.MonitorDatabase
	monitorDatabase.Id = sequence.GetSequence().Generate().Int64()
	err := application.repository.MonitorDatabaseRepository.Save(&monitorDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *MonitorDatabaseApplication) Modify(request *types.MonitorDatabaseCreateRequest) error {
	monitorDatabase := request.MonitorDatabase
	err := application.repository.MonitorDatabaseRepository.UpdateById(monitorDatabase.Id, &monitorDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorDatabaseApplication) SelectAll() ([]entity.MonitorDatabase, error) {
	return application.repository.MonitorDatabaseRepository.SelectSimpleAll()
}
