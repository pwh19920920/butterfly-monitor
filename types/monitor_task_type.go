package types

import (
	"butterfly-monitor/domain/entity"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly-admin/common"
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
	DatabaseId      *int64   `json:"databaseId,string"`
	ResultFieldPath string   `json:"resultFieldPath"`
	CollectName     string   `json:"collectName"`  // 集合名称
	DefaultValue    *float64 `json:"defaultValue"` // 默认值
	Database        string   `json:"database"`
	RetentionPolicy string   `json:"retentionPolicy"`
	Column          string   `json:"column"`
}

type MonitorTaskQueryResponse struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams         `json:"taskExecParams"`
	Dashboards     []string                      `json:"dashboards"`
	TaskAlert      MonitorTaskAlertCreateRequest `json:"taskAlert"`
}

type MonitorTaskAlertCreateRequest struct {
	entity.MonitorTaskAlert
	EffectTimes   []string                         `json:"effectTimes"`
	AlertChannels []string                         `json:"alertChannels"`
	AlertGroups   []string                         `json:"alertGroups"`
	CheckParams   []entity.MonitorAlertCheckParams `json:"checkParams"`
}

type MonitorTaskCreateRequest struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams         `json:"taskExecParams"`
	Dashboards     []string                      `json:"dashboards"`
	TaskAlert      MonitorTaskAlertCreateRequest `json:"taskAlert"`
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
		validation.Field(&req.StepSpan, validation.Required, validation.Min(30)),
		validation.Field(&req.Command, validation.Required, validation.Length(10, 1000)),
		validation.Field(&req.Dashboards, validation.Required),
	)
}

func (req MonitorTaskExecForRangeRequest) ValidateForExec() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.BeginDate, validation.Required),
		validation.Field(&req.EndDate, validation.Required),
	)
}
