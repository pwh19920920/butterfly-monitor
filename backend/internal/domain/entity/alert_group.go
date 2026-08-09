package entity

import "butterfly-monitor/internal/common"

// AlertGroup 告警接收人分组
type AlertGroup struct {
	common.BaseEntity

	Name string `json:"name" gorm:"column:name"`
}

func (AlertGroup) TableName() string {
	return "t_alert_group"
}
