package application

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/types"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/pwh19920920/butterfly-admin/src/app/common"
	"github.com/pwh19920920/butterfly-admin/src/app/config/sequence"
	"github.com/sirupsen/logrus"
	"time"
)

type MonitorTaskApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// Query 分页查询
func (application *MonitorTaskApplication) Query(request *types.MonitorTaskQueryRequest) (int64, []types.MonitorTaskQueryResponse, error) {
	total, data, err := application.repository.MonitorTaskRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.Select() happen error for", err)
		return total, nil, err
	}

	result := make([]types.MonitorTaskQueryResponse, 0)
	for _, item := range data {
		var execParam types.MonitorTaskExecParams
		_ = json.Unmarshal([]byte(item.ExecParams), &execParam)
		result = append(result, types.MonitorTaskQueryResponse{MonitorTask: item, TaskExecParams: execParam})
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
	err := application.repository.MonitorTaskRepository.Save(&monitorTask)

	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.Save() happen error", err)
	}
	return err
}

// Modify 修改数据源
func (application *MonitorTaskApplication) Modify(request *types.MonitorTaskCreateRequest) error {
	execParams, _ := json.Marshal(request.TaskExecParams)

	monitorTask := request.MonitorTask
	monitorTask.ExecParams = string(execParams)
	err := application.repository.MonitorTaskRepository.UpdateById(monitorTask.Id, &entity.MonitorTask{
		TaskKey:    monitorTask.TaskKey,
		TaskName:   monitorTask.TaskName,
		TimeSpan:   monitorTask.TimeSpan,
		ExecParams: monitorTask.ExecParams,
		TaskType:   monitorTask.TaskType,
		Command:    monitorTask.Command,
	})

	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateById() happen error", err)
	}
	return err
}

func (application *MonitorTaskApplication) ModifyTaskStatus(taskId int64, status entity.MonitorTaskStatus) error {
	err := application.repository.MonitorTaskRepository.UpdateTaskStatusById(taskId, status)
	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateTaskStatusById() happen error", err)
	}
	return err
}

func (application *MonitorTaskApplication) ModifyAlertStatus(taskId int64, status entity.MonitorAlertStatus) error {
	err := application.repository.MonitorTaskRepository.UpdateAlertStatusById(taskId, status)
	// 错误记录
	if err != nil {
		logrus.Error("MonitorTaskRepository.UpdateById() happen error", err)
	}
	return err
}
