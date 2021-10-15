package application

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/types"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/src/app/config/sequence"
	"github.com/sirupsen/logrus"
)

type JobDatabaseApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// Query 分页查询
func (application *JobDatabaseApplication) Query(request *types.JobDatabaseQueryRequest) (int64, []entity.JobDatabase, error) {
	total, data, err := application.repository.JobDatabaseRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("JobDatabaseRepository.Select() happen error for", err)
	}
	return total, data, err
}

// Create 创建数据源
func (application *JobDatabaseApplication) Create(request *types.JobDatabaseCreateRequest) error {
	jobDatabase := request.JobDatabase
	jobDatabase.Id = sequence.GetSequence().Generate().Int64()
	err := application.repository.JobDatabaseRepository.Save(&jobDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("JobDatabaseRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *JobDatabaseApplication) Modify(request *types.JobDatabaseCreateRequest) error {
	jobDatabase := request.JobDatabase
	err := application.repository.JobDatabaseRepository.UpdateById(jobDatabase.Id, &jobDatabase)

	// 错误记录
	if err != nil {
		logrus.Error("JobDatabaseRepository.UpdateById() happen error", err)
	}
	return err
}
