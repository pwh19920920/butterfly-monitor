package entity

import "butterfly-monitor/internal/common"

// MonitorDashboardTask 面板-任务关联
type MonitorDashboardTask struct {
	common.BaseEntity

	TaskId      int64 `json:"taskId,string" gorm:"column:task_id"`
	DashboardId int64 `json:"dashboardId,string" gorm:"column:dashboard_id"`
	Sort        int32 `json:"sort" gorm:"column:sort"`
}

func (MonitorDashboardTask) TableName() string {
	return "t_monitor_dashboard_task"
}
