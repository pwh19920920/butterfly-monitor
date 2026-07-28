package types

import (
	"errors"
	"strconv"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/pkg/response"
)

type MonitorTaskQueryRequest struct {
	response.RequestPaging
	TaskName    string                     `form:"taskName"`
	TaskKey     string                     `form:"taskKey"`
	TaskType    *entity.MonitorTaskType    `form:"taskType"`
	TaskStatus  *entity.MonitorTaskStatus  `form:"taskStatus"`
	AlertStatus *entity.MonitorAlertStatus `form:"alertStatus"`
}

type MonitorTaskExecParams struct {
	DatabaseId      *int64   `json:"databaseId,string"`
	ResultFieldPath string   `json:"resultFieldPath"`
	CollectName     string   `json:"collectName"`
	DefaultValue    *float64 `json:"defaultValue"`
	Database        string   `json:"database"`
	RetentionPolicy string   `json:"retentionPolicy"`
	Column          string   `json:"column"`
}

type MonitorTaskAlertCreateRequest struct {
	entity.MonitorTaskAlert
	EffectTimes   []string                         `json:"effectTimes"`
	AlertChannels []string                         `json:"alertChannels"`
	AlertGroups   []string                         `json:"alertGroups"`
	CheckParams   []entity.MonitorAlertCheckParams `json:"checkParams"`
}

// HasAlertConfig 判断是否填写了告警配置
// 只要存在检测规则、告警通道、告警分组，或配置了检测周期/持续时长，即视为填写了告警
func (req MonitorTaskAlertCreateRequest) HasAlertConfig() bool {
	if len(req.CheckParams) > 0 || len(req.AlertChannels) > 0 || len(req.AlertGroups) > 0 {
		return true
	}
	return req.TimeSpan > 0 || req.Duration > 0
}

// hasAnyCheckConfig 报警检查配置是否有任一字段
func (req MonitorTaskAlertCreateRequest) hasAnyCheckConfig() bool {
	return len(req.AlertChannels) > 0 || len(req.AlertGroups) > 0 || req.TimeSpan > 0 || req.Duration > 0
}

// hasCompleteCheckConfig 报警检查配置是否齐全
// needAlertGroups=false 时（所选通道全为 Webhook）不要求报警分组
func (req MonitorTaskAlertCreateRequest) hasCompleteCheckConfig(needAlertGroups bool) bool {
	if len(req.AlertChannels) == 0 || req.TimeSpan <= 0 || req.Duration <= 0 {
		return false
	}
	if needAlertGroups && len(req.AlertGroups) == 0 {
		return false
	}
	return true
}

// ValidateOptional 报警检查配置与异常检测规则均为非必填；
// 若填了任意一侧，则另一侧也必须完整填写（对称校验）
// needAlertGroups：所选通道是否需要报警分组（非 Webhook 需要）
func (req MonitorTaskAlertCreateRequest) ValidateOptional(needAlertGroups bool) error {
	hasAnyCheckConfig := req.hasAnyCheckConfig()
	hasCheckParams := len(req.CheckParams) > 0

	// 两侧都空：允许（纯采集任务）
	if !hasAnyCheckConfig && !hasCheckParams {
		return nil
	}
	// 只填了报警检查配置
	if hasAnyCheckConfig && !hasCheckParams {
		return errors.New("已填写报警检查配置，必须补充异常检测规则")
	}
	// 只填了异常检测规则
	if hasCheckParams && !hasAnyCheckConfig {
		if needAlertGroups {
			return errors.New("已填写异常检测规则，必须补充报警检查配置（通道/分组/间隔/持续时间）")
		}
		return errors.New("已填写异常检测规则，必须补充报警检查配置（通道/间隔/持续时间）")
	}
	// 两侧都有内容：报警检查配置必须齐全
	if !req.hasCompleteCheckConfig(needAlertGroups) {
		if needAlertGroups {
			return errors.New("报警检查配置需完整填写：报警通道、报警分组、检查间隔、持续时间")
		}
		return errors.New("报警检查配置需完整填写：报警通道、检查间隔、持续时间（Webhook 无需分组）")
	}
	// 每组规则至少一条
	for _, cp := range req.CheckParams {
		if len(cp.Rules) == 0 {
			return errors.New("每个规则条件组至少需要一条规则")
		}
	}
	return nil
}

type MonitorTaskCreateRequest struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams         `json:"taskExecParams"`
	Dashboards     []string                      `json:"dashboards"`
	TaskAlert      MonitorTaskAlertCreateRequest `json:"taskAlert"`
}

// MonitorTaskListItem 任务列表行：任务本体 + 规则运行态
// TaskAlertStatus：无告警配置时固定为 1(正常)；有配置时取 t_monitor_task_alert.alert_status
type MonitorTaskListItem struct {
	entity.MonitorTask
	// TaskAlertStatus 规则运行态：1正常 2异常(Pending) 3告警(Firing)；无配置视为 1
	TaskAlertStatus entity.MonitorTaskAlertStatus `json:"taskAlertStatus"`
	// FirstFlagTime 首次出现异常的时间（first_flag_time）；从未异常为 nil
	FirstFlagTime *common.LocalTime `json:"firstFlagTime,omitempty"`
}

type MonitorTaskQueryResponse struct {
	entity.MonitorTask
	TaskExecParams MonitorTaskExecParams         `json:"taskExecParams"`
	Dashboards     []string                      `json:"dashboards"`
	TaskAlert      MonitorTaskAlertCreateRequest `json:"taskAlert"`
	// TaskAlertStatus 规则运行态（详情页同样返回，语义同列表）
	TaskAlertStatus entity.MonitorTaskAlertStatus `json:"taskAlertStatus"`
}

func (req MonitorTaskCreateRequest) GetDashboardIds() ([]int64, error) {
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

func (req MonitorTaskCreateRequest) ValidateForCreate() error {
	// Push 任务由外部推送数据，无执行指令，跳过 command 校验
	cmdRule := []validation.Rule{validation.Required, validation.Length(1, 2000)}
	if req.TaskType != nil && *req.TaskType == entity.TaskTypePush {
		cmdRule = []validation.Rule{validation.Length(0, 2000)}
	}
	return validation.ValidateStruct(&req,
		validation.Field(&req.TaskKey, validation.Required, validation.Length(1, 255)),
		validation.Field(&req.TaskName, validation.Required, validation.Length(1, 255)),
		validation.Field(&req.TimeSpan, validation.Required, validation.Min(1)),
		validation.Field(&req.StepSpan, validation.Required, validation.Min(1)),
		validation.Field(&req.Command, cmdRule...),
	)
}

type MonitorTaskExecForRangeRequest struct {
	BeginDate *common.LocalTime `json:"beginDate"`
	EndDate   *common.LocalTime `json:"endDate"`
}

// MonitorTaskPushDataItem 单条推送数据
type MonitorTaskPushDataItem struct {
	Time      *common.LocalTime `json:"time"`      // 时间点(与 timestamp 二选一)
	Timestamp *int64            `json:"timestamp"` // 毫秒时间戳(与 time 二选一)
	Value     float64           `json:"value"`     // 指标值
}

// MonitorTaskPushDataRequest 外部数据推送请求
// TaskKey 由 body 绑定，handler 转换为 TaskId 后向下传递给 job
type MonitorTaskPushDataRequest struct {
	TaskKey string                    `json:"taskKey"` // 任务标识
	TaskId  int64                     `json:"-"`       // 由 TaskKey 转换得到
	Items   []MonitorTaskPushDataItem `json:"items"`
}
