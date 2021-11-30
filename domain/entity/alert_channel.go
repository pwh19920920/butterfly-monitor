package entity

import (
	"github.com/pwh19920920/butterfly-admin/common"
)

type AlertChannelType int32
type AlertChannelFailRoute int32

const (
	AlertChannelTypeEmail   AlertChannelType = 1
	AlertChannelTypeWebhook AlertChannelType = 2
	AlertChannelTypeSMS     AlertChannelType = 3

	AlertChannelFailRouteTrue  AlertChannelFailRoute = 1
	AlertChannelFailRouteFalse AlertChannelFailRoute = 2
)

type AlertChannel struct {
	common.BaseEntity

	Name      string                `json:"name" gorm:"column:name"`            // 通道名称
	Type      AlertChannelType      `json:"type" gorm:"column:type"`            // 通道类型
	Params    string                `json:"params" gorm:"column:params"`        // 通道参数
	Handler   string                `json:"handler" gorm:"column:handler"`      // 通道key
	FailRoute AlertChannelFailRoute `json:"failRoute" gorm:"column:fail_route"` // 失败路由
}

// TableName 会将 User 的表名重写为 `profiles`
func (AlertChannel) TableName() string {
	return "t_alert_channel"
}
