package types

import (
	"butterfly-monitor/src/app/domain/entity"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly-admin/src/app/common"
	"github.com/pwh19920920/butterfly/response"
	"strconv"
)

type MonitorTaskQueryRequest struct {
	response.RequestPaging

	TaskName    string                     `form:"taskName"`
	TaskKey     string                     `form:"taskKey"`
	TaskType    *entity.MonitorTaskType    `form:"taskType"`
	TaskStatus  *entity.MonitorTaskStatus  `form:"taskStatus" `  // 任务开关
	AlertStatus *entity.MonitorAlertStatus `form:"alertStatus" ` // 报警开关
}

type MonitorTaskExecParams struct {
	DatabaseId      *int64 `json:"databaseId,string"`
	ResultFieldPath string `json:"resultFieldPath"`
}

type MonitorTaskQueryResponse struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams `json:"taskExecParams"`
	Dashboards     []string              `json:"dashboards"`
}

type MonitorTaskCreateRequest struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams `json:"taskExecParams"`
	Dashboards     []string              `json:"dashboards"`
}

func (req MonitorTaskCreateRequest) GetDashboardIds() ([]int64, error) {
	// 转换
	dashboardIds := make([]int64, 0)
	for _, dashboardIdStr := range req.Dashboards {
		id, err := strconv.ParseInt(dashboardIdStr, 10, 64)
		if err != nil {
			return dashboardIds, err
		}
		dashboardIds = append(dashboardIds, id)
	}
	return dashboardIds, nil
}

type MonitorTaskExecForRangeRequest struct {
	BeginDate *common.LocalTime `json:"beginDate"` // 开始日期
	EndDate   *common.LocalTime `json:"endDate"`   // 结束日期
}

func (req MonitorTaskCreateRequest) ValidateForCreate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.TaskKey, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.TaskName, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.TimeSpan, validation.Required, validation.Min(30)),
		validation.Field(&req.Command, validation.Required, validation.Length(10, 1000)),
		validation.Field(&req.TaskType, validation.Required),
		validation.Field(&req.Dashboards, validation.Required),
	)
}
