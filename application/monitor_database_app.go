package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"errors"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
)

type MonitorDatabaseApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	commonMap  CommonMapApplication
}

func (application *MonitorDatabaseApplication) Count() (*int64, error) {
	return application.repository.MonitorDatabaseRepository.Count()
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

// 数据源检测
func (application *MonitorDatabaseApplication) checkDatabase(monitorDatabase entity.MonitorDatabase) error {
	// 检测database是否能连接得上
	databaseHandler, ok := application.commonMap.GetDatabaseHandlerMap()[monitorDatabase.Type]
	if !ok {
		return errors.New("不存在此数据源类型")
	}

	// 检测连接
	return databaseHandler.TestConnect(monitorDatabase)
}

// Create 创建数据源
func (application *MonitorDatabaseApplication) Create(request *types.MonitorDatabaseCreateRequest) error {
	monitorDatabase := request.MonitorDatabase

	// 检测database是否能连接得上
	err := application.checkDatabase(monitorDatabase)
	if err != nil {
		return err
	}

	monitorDatabase.Id = application.sequence.Generate().Int64()
	err = application.repository.MonitorDatabaseRepository.Save(&monitorDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *MonitorDatabaseApplication) Modify(request *types.MonitorDatabaseCreateRequest) error {
	monitorDatabase := request.MonitorDatabase

	// 检测database是否能连接得上
	err := application.checkDatabase(monitorDatabase)
	if err != nil {
		return err
	}

	err = application.repository.MonitorDatabaseRepository.UpdateById(monitorDatabase.Id, &monitorDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorDatabaseRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorDatabaseApplication) SelectAll() ([]entity.MonitorDatabase, error) {
	return application.repository.MonitorDatabaseRepository.SelectSimpleAll()
}
