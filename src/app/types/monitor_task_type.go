package types

import (
	"butterfly-monitor/src/app/domain/entity"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/response"
)

type MonitorTaskQueryRequest struct {
	response.RequestPaging

	TaskName string                 `form:"taskName"`
	TaskKey  string                 `form:"taskKey"`
	TaskType entity.MonitorTaskType `form:"taskType"`
}

type MonitorTaskExecParams struct {
	DatabaseId *int64 `json:"databaseId,string"`
}

type MonitorTaskQueryResponse struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams `json:"taskExecParams"`
}

type MonitorTaskCreateRequest struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams `json:"taskExecParams"`
}

func (req MonitorTaskCreateRequest) ValidateForCreate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.TaskKey, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.TaskName, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.TimeSpan, validation.Required, validation.Min(30)),
		validation.Field(&req.Command, validation.Required, validation.Length(10, 1000)),
		validation.Field(&req.TaskType, validation.Required),
	)
}
