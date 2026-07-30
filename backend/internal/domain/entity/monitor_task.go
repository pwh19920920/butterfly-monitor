package entity

import (
	"dragonfly-monitor/internal/common"
)

type MonitorTaskType int32
type MonitorTaskStatus int32
type MonitorAlertStatus int32
type MonitorSampledStatus int32
type MonitorTaskDataType int32

const (
	// TaskTypeDrilldown 系统下钻：数据源为时序库(VM)，从依赖的聚合任务结果中按标签过滤取数。
	// 值为 0 是特殊标记，与正常任务(1/2/3)互斥，必须在 executeCollect 中显式 ==0 判断。
	TaskTypeDrilldown MonitorTaskType = 0
	TaskTypeDatabase  MonitorTaskType = 1
	TaskTypeURL       MonitorTaskType = 2
	TaskTypePush      MonitorTaskType = 3

	// DataTypeNormal 单值采集（默认）
	DataTypeNormal MonitorTaskDataType = 1
	// DataTypeAggregate 分组聚合采集：多行分组结果写入时序库，仅收集不做采样/告警
	DataTypeAggregate MonitorTaskDataType = 2

	MonitorTaskStatusOpen  MonitorTaskStatus = 1
	MonitorTaskStatusClose MonitorTaskStatus = 0

	MonitorAlertStatusOpen  MonitorAlertStatus = 1
	MonitorAlertStatusClose MonitorAlertStatus = 0

	MonitorSampledStatusOpen  MonitorSampledStatus = 1
	MonitorSampledStatusClose MonitorSampledStatus = 0
)

// MonitorTask 监控任务
type MonitorTask struct {
	common.BaseEntity

	PreExecuteTime *common.LocalTime    `json:"preExecuteTime" gorm:"column:pre_execute_time"` // 上次采集时间
	PreSampleTime  *common.LocalTime    `json:"preSampleTime" gorm:"column:pre_sample_time"`   // 上次样本生成时间
	SampleErrMsg   string               `json:"sampleErrMsg" gorm:"column:sample_err_msg"`     // 样本错误
	CollectErrMsg  string               `json:"collectErrMsg" gorm:"column:collect_err_msg"`   // 采集错误
	TaskKey        string               `json:"taskKey" gorm:"column:task_key"`                // 任务标识 / VM 指标名
	TaskName       string               `json:"taskName" gorm:"column:task_name"`              // 任务名称
	TimeSpan       int32                `json:"timeSpan" gorm:"column:time_span"`              // 窗口前进间隔(秒)
	StepSpan       int32                `json:"stepSpan" gorm:"column:step_span"`              // 查询区间宽度(秒)
	Command        string               `json:"command" gorm:"column:command"`                 // 执行指令
	TaskType       *MonitorTaskType     `json:"taskType" gorm:"column:task_type"`              // 任务类型
	ExecParams     string               `json:"execParams" gorm:"column:exec_params"`          // 执行参数 JSON
	TaskStatus     MonitorTaskStatus    `json:"taskStatus" gorm:"column:task_status"`          // 任务开关
	AlertStatus    MonitorAlertStatus   `json:"alertStatus" gorm:"column:alert_status"`        // 告警开关
	Sampled        MonitorSampledStatus `json:"sampled" gorm:"column:sampled"`                 // 样本展示开关(仅 Grafana)
	MonitorGroup   string               `json:"monitorGroup" gorm:"column:monitor_group"`      // 依赖分组
	Labels         string               `json:"labels" gorm:"column:labels"`                   // 标签 JSON
	DataType       MonitorTaskDataType  `json:"dataType" gorm:"column:data_type"`              // 采集数据类型：1单值 2聚合
	RelatedTaskIds string               `json:"relatedTaskIds" gorm:"column:related_task_ids"` // 关联任务 ID（逗号分隔），叠加实时/样本曲线到本面板
}

func (MonitorTask) TableName() string {
	return "t_monitor_task"
}
