package entity

import "dragonfly-monitor/internal/common"

// MonitorDashboard 监控面板（Grafana）
type MonitorDashboard struct {
	common.BaseEntity

	Name    string `json:"name" gorm:"column:name"`
	Slug    string `json:"slug" gorm:"column:slug"`
	Url     string `json:"url" gorm:"column:url"`
	Uid     string `json:"uid" gorm:"column:uid"`
	BoardId *int64 `json:"boardId,string" gorm:"column:board_id"`
}

func (MonitorDashboard) TableName() string {
	return "t_monitor_dashboard"
}
