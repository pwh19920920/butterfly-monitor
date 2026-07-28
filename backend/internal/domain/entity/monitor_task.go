package entity

import (
	"dragonfly-monitor/internal/common"
)

type MonitorTaskType int32
type MonitorTaskStatus int32
type MonitorAlertStatus int32
type MonitorSampledStatus int32

const (
	TaskTypeDatabase MonitorTaskType = 1
	TaskTypeURL      MonitorTaskType = 2
	TaskTypePush     MonitorTaskType = 3

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
}

func (MonitorTask) TableName() string {
	return "t_monitor_task"
}
