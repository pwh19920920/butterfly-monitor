package entity

import (
	"github.com/pwh19920920/butterfly-admin/common"
)

type AlertGroup struct {
	common.BaseEntity

	Name string `json:"name" gorm:"column:name"` // 通道名称
}

// TableName 会将 User 的表名重写为 `profiles`
func (AlertGroup) TableName() string {
	return "t_alert_group"
}
