package entity

import "github.com/pwh19920920/butterfly-admin/src/app/common"

type TaskType int32

const (
	TaskTypeDatabase TaskType = 0
	TaskTypeURL      TaskType = 1
)

type MonitorTask struct {
	common.BaseEntity

	PreExecuteTime common.LocalTime `json:"pre_execute_time" gorm:"column:pre_execute_time"` // 上一次执行时间
	TaskKey        string           `json:"task_key" gorm:"column:task_key"`                 // 任务标识
	TaskName       string           `json:"task_name" gorm:"column:task_name"`               // 任务名称
	TimeSpan       int32            `json:"time_space" gorm:"column:time_space"`             // 时间间隔
	Command        string           `json:"command" gorm:"column:command"`                   // 执行指令, 可以是url, 也可以是sql
	TaskType       TaskType         `json:"task_type" gorm:"column:task_type"`               // 任务类型, db, url
	DatabaseId     int64            `json:"database_id" gorm:"database_id"`                  // 如果是db, 执行db的id
	ExecParams     string           `json:"exec_params"`                                     // 任务执行参数
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorTask) TableName() string {
	return "t_monitor_task"
}
