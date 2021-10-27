package entity

import "github.com/pwh19920920/butterfly-admin/src/app/common"

type MonitorDashboardTask struct {
	common.BaseEntity

	DashboardId int64 `json:"dashboardId,string" gorm:"dashboard_id"` // 主板id
	TaskId      int64 `json:"taskId,string" gorm:"task_id"`           // 任务id
	Sort        int32 `json:"sort" gorm:"sort"`                       // 排序
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorDashboardTask) TableName() string {
	return "t_monitor_dashboard_task"
}
