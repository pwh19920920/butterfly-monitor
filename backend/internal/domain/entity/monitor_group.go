package entity

import "dragonfly-monitor/internal/common"

// MonitorGroup 树形依赖分组
type MonitorGroup struct {
	common.BaseEntity

	Name   string `json:"name" gorm:"column:name"`
	Route  string `json:"route" gorm:"column:route"` // 形如 /1/2/3/
	Parent int64  `json:"parent,string" gorm:"column:parent"`
}

func (MonitorGroup) TableName() string {
	return "t_monitor_group"
}
