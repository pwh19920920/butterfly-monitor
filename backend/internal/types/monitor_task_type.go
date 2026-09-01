package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/pkg/response"
)

// taskKeyRe 限定 TaskKey 格式：字母或下划线开头，后跟字母/数字/下划线。
// 因为在 VictoriaMetrics / TDengine / Grafana 中作为 metric 名必须为合法标识符。
var taskKeyRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type MonitorTaskQueryRequest struct {
	response.RequestPaging
	TaskName    string                      `form:"taskName"`
	TaskKey     string                      `form:"taskKey"`
	TaskType    *entity.MonitorTaskType     `form:"taskType"`
	TaskStatus  *entity.MonitorTaskStatus   `form:"taskStatus"`
	AlertStatus *entity.MonitorAlertStatus  `form:"alertStatus"`
	DataType    *entity.MonitorTaskDataType `form:"dataType"`
}

type MonitorTaskExecParams struct {
	DatabaseId      *int64   `json:"databaseId,string"`
	ResultFieldPath string   `json:"resultFieldPath"`
	CollectName     string   `json:"collectName"`
	DefaultValue    *float64 `json:"defaultValue"`
	Database        string   `json:"database"`
	RetentionPolicy string   `json:"retentionPolicy"`
	Column          string   `json:"column"`

	// 聚合任务专用 (DataType=Aggregate)
	LabelColumns []string `json:"labelColumns"` // 分组维度列名，每个元素成为一个 tag
	ValueColumns []string `json:"valueColumns"` // 指标值列名，每个元素生成一个独立 metric

	// 系统下钻任务专用 (TaskType=Drilldown)
	SourceTaskId *int64       `json:"sourceTaskId,string"` // 依赖的聚合任务 ID
	Filters      []FilterRule `json:"filters"`             // 标签过滤条件，维度数量需与聚合分组层数一致
	QueryMetric  string       `json:"queryMetric"`         // 指定取聚合任务哪个 valueColumn 对应的 metric
}

// FilterRule 下钻任务的标签过滤条件
// FieldName 由源聚合任务的 LabelColumns 带过来（用户不可改），Value 由用户填写（单值精确匹配）
type FilterRule struct {
	FieldName string `json:"fieldName"` // 维度名，如 region / dept
	Operator  string `json:"operator"`  // 当前固定为等值匹配 eq
	Value     string `json:"value"`     // 过滤值，下钻需每个维度全填，恰好命中 1 个 series
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
	// needAlertGroups 仅决定"是否要求报警分组"，由此区分两类文案
	requireGroupWord := "（通道/分组/间隔/持续时间）" // #nosec G101 -- 中文提示文案，非凭据
	completeGroupHint := "报警通道、报警分组、检查间隔、持续时间"
	if !needAlertGroups {
		requireGroupWord = "（通道/间隔/持续时间）" // #nosec G101 -- 中文提示文案，非凭据
		completeGroupHint = "报警通道、检查间隔、持续时间（Webhook 无需分组）"
	}

	hasAnyCheckConfig := req.hasAnyCheckConfig()
	hasCheckParams := len(req.CheckParams) > 0

	switch {
	case !hasAnyCheckConfig && !hasCheckParams:
		// 两侧都空：允许（纯采集任务）
		return nil
	case hasAnyCheckConfig && !hasCheckParams:
		// 只填了报警检查配置
		return errors.New("已填写报警检查配置，必须补充异常检测规则")
	case !hasAnyCheckConfig:
		// 只填了异常检测规则
		return errors.New("已填写异常检测规则，必须补充报警检查配置" + requireGroupWord)
	case !req.hasCompleteCheckConfig(needAlertGroups):
		// 两侧都有内容，但报警检查配置不齐全
		return errors.New("报警检查配置需完整填写：" + completeGroupHint)
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
	TaskAlertStatus entity.MonitorTaskAlertStatus `json:"taskAlertStatus"`         // TaskAlertStatus 规则运行态：1正常 2异常(Pending) 3告警(Firing)；无配置视为 1
	FirstFlagTime   *common.LocalTime             `json:"firstFlagTime,omitempty"` // FirstFlagTime 首次出现异常的时间（first_flag_time）；从未异常为 nil
}

type MonitorTaskQueryResponse struct {
	entity.MonitorTask
	TaskExecParams  MonitorTaskExecParams         `json:"taskExecParams"`
	Dashboards      []string                      `json:"dashboards"`
	TaskAlert       MonitorTaskAlertCreateRequest `json:"taskAlert"`
	TaskAlertStatus entity.MonitorTaskAlertStatus `json:"taskAlertStatus"` // TaskAlertStatus 规则运行态（详情页同样返回，语义同列表）
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
	// Push / 系统下钻 任务由外部或时序库取数，无执行指令，跳过 command 校验
	cmdRule := []validation.Rule{validation.Required, validation.Length(1, 2000)}
	if req.TaskType != nil &&
		(*req.TaskType == entity.TaskTypePush || *req.TaskType == entity.TaskTypeDrilldown) {
		cmdRule = []validation.Rule{validation.Length(0, 2000)}
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.TaskKey, validation.Required,
			validation.Length(1, 255),
			validation.Match(taskKeyRe).Error("TaskKey 必须以字母或下划线开头，且仅允许字母、数字、下划线"),
		),
		validation.Field(&req.TaskName, validation.Required, validation.Length(1, 255)),
		validation.Field(&req.TimeSpan, validation.Required, validation.Min(1)),
		validation.Field(&req.StepSpan, validation.Required, validation.Min(1)),
		validation.Field(&req.Command, cmdRule...),
	); err != nil {
		return err
	}
	// 按任务形态做结构化字段校验（引用存在性在 application 层校验）
	return req.ValidateTypedParams()
}

// ValidateTypedParams 按 DataType / TaskType 校验 ExecParams 结构完整性。
// 不访问数据库；源任务是否存在、是否为聚合，由 application 层补齐。
func (req MonitorTaskCreateRequest) ValidateTypedParams() error {
	if req.DataType == entity.DataTypeAggregate {
		if len(req.TaskExecParams.LabelColumns) == 0 {
			return errors.New("聚合任务必须配置 labelColumns（分组维度）")
		}
		if len(req.TaskExecParams.ValueColumns) == 0 {
			return errors.New("聚合任务必须配置 valueColumns（指标列）")
		}
		// 聚合不支持下钻类型混用
		if req.TaskType != nil && *req.TaskType == entity.TaskTypeDrilldown {
			return errors.New("聚合任务不能同时为系统下钻类型")
		}
		if req.TaskType != nil && *req.TaskType == entity.TaskTypePush {
			return errors.New("聚合任务不支持 Push 类型")
		}
		return nil
	}

	// Push 任务由外部推送数据，无采集执行，不依赖默认值
	if req.TaskType != nil && *req.TaskType == entity.TaskTypePush {
		return nil
	}

	// 单值任务（Database/URL/mongo/下钻）：无结果默认值必填，供采集无数据时回落
	if req.TaskExecParams.DefaultValue == nil {
		return errors.New("无结果默认值（defaultValue）不能为空")
	}

	if req.TaskType != nil && *req.TaskType == entity.TaskTypeDrilldown {
		if req.TaskExecParams.SourceTaskId == nil || *req.TaskExecParams.SourceTaskId == 0 {
			return errors.New("下钻任务必须指定 sourceTaskId（依赖的聚合任务）")
		}
		// filters 允许为空（源只有单 series 时），但 fieldName 若填了不能为空串
		for i, f := range req.TaskExecParams.Filters {
			if f.FieldName == "" {
				return fmt.Errorf("下钻 filters[%d].fieldName 不能为空", i)
			}
			if f.Value == "" {
				return fmt.Errorf("下钻 filters[%d].value 不能为空（维度需全填以命中唯一 series）", i)
			}
		}
	}
	return nil
}

type MonitorTaskExecForRangeRequest struct {
	BeginDate *common.LocalTime `json:"beginDate"`
	EndDate   *common.LocalTime `json:"endDate"`
}

// MonitorTaskPreviewRequest 聚合预览请求：临时执行多行查询，返回结果列名，供前端勾选 label/value 维度
type MonitorTaskPreviewRequest struct {
	TaskType   *entity.MonitorTaskType `json:"taskType"`   // 数据源类型（Database/URL）
	DataSource *entity.DataSourceType  `json:"dataSource"` // Database 任务数据源类型（含 Mysql 协议族/Mongo/PG/ClickHouse/Prometheus/OpenSearch/ES）
	DatabaseId *int64                  `json:"databaseId,string"`
	Command    string                  `json:"command"`
	ExecParams MonitorTaskExecParams   `json:"execParams"`
}

// MonitorTaskPreviewResponse 聚合预览结果：列名列表
type MonitorTaskPreviewResponse struct {
	Columns []string `json:"columns"`
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
