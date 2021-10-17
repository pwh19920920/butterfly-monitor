package application

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/types"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/src/app/config/sequence"
	"github.com/sirupsen/logrus"
)

type MonitorTaskApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// Query 分页查询
func (application *MonitorTaskApplication) Query(request *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error) {
	total, data, err := application.repository.MonitorDatabaseRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.Select() happen error for", err)
	}
	return total, data, err
}

// Create 创建数据源
func (application *MonitorTaskApplication) Create(request *types.MonitorDatabaseCreateRequest) error {
	jobDatabase := request.MonitorDatabase
	jobDatabase.Id = sequence.GetSequence().Generate().Int64()
	err := application.repository.MonitorDatabaseRepository.Save(&jobDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *MonitorTaskApplication) Modify(request *types.MonitorDatabaseCreateRequest) error {
	jobDatabase := request.MonitorDatabase
	err := application.repository.MonitorDatabaseRepository.UpdateById(jobDatabase.Id, &jobDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.UpdateById() happen error", err)
	}
	return err
}
