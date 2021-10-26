package entity

import "github.com/pwh19920920/butterfly-admin/src/app/common"

type MonitorDashboard struct {
	common.BaseEntity

	Name    string `json:"name" gorm:"column:name"` // 中文名
	Slug    string `json:"slug" column:"slug"`      // 英文名
	Url     string `json:"url" gorm:"column:url"`   // 地址
	Uid     string `json:"uid" gorm:"column:uid"`   // uid
	BoardId *uint  `json:"BoardId" gorm:"board_id"` // 主板id
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorDashboard) TableName() string {
	return "t_monitor_dashboard"
}
