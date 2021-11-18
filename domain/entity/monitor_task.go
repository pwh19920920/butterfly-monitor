package entity

import (
	"github.com/pwh19920920/butterfly-admin/common"
)

type MonitorTaskType int32
type MonitorTaskStatus int32
type MonitorAlertStatus int32
type MonitorSampledStatus int32

const (
	TaskTypeDatabase          MonitorTaskType      = 0
	TaskTypeURL               MonitorTaskType      = 1
	MonitorTaskStatusOpen     MonitorTaskStatus    = 1
	MonitorTaskStatusClose    MonitorTaskStatus    = 0
	MonitorAlertStatusOpen    MonitorAlertStatus   = 1
	MonitorAlertStatusClose   MonitorAlertStatus   = 0
	MonitorSampledStatusOpen  MonitorSampledStatus = 1
	MonitorSampledStatusClose MonitorSampledStatus = 0
)

type MonitorTask struct {
	common.BaseEntity

	PreExecuteTime *common.LocalTime    `json:"preExecuteTime" gorm:"column:pre_execute_time"` // 上一次执行时间
	PreSampleTime  *common.LocalTime    `json:"preSampleTime" gorm:"column:pre_sample_time"`   // 上一次样本执行时间
	SampleErrMsg   string               `json:"sampleErrMsg" gorm:"sample_err_msg"`            // 错误原因
	CollectErrMsg  string               `json:"collectErrMsg" gorm:"collect_err_msg"`          // 错误原因
	TaskKey        string               `json:"taskKey" gorm:"column:task_key"`                // 任务标识
	TaskName       string               `json:"taskName" gorm:"column:task_name"`              // 任务名称
	TimeSpan       int32                `json:"timeSpan" gorm:"column:time_span"`              // 时间间隔
	Command        string               `json:"command" gorm:"column:command"`                 // 执行指令, 可以是url, 也可以是sql
	TaskType       MonitorTaskType      `json:"taskType" gorm:"column:task_type"`              // 任务类型, db, url
	ExecParams     string               `json:"execParams" gorm:"exec_params"`                 // 任务执行参数
	TaskStatus     MonitorTaskStatus    `json:"taskStatus" gorm:"task_status"`                 // 任务开关
	AlertStatus    MonitorAlertStatus   `json:"alertStatus" gorm:"alert_status"`               // 报警开关
	Sampled        MonitorSampledStatus `json:"sampled" gorm:"sampled"`                        // 是否生成样本
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorTask) TableName() string {
	return "t_monitor_task"
}
