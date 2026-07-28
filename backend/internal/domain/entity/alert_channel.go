package entity

import "dragonfly-monitor/internal/common"

type AlertChannelType int32
type AlertChannelFailRoute int32

const (
	AlertChannelTypeEmail   AlertChannelType = 1
	AlertChannelTypeWebhook AlertChannelType = 2
	AlertChannelTypeSMS     AlertChannelType = 3

	AlertChannelFailRouteTrue  AlertChannelFailRoute = 1
	AlertChannelFailRouteFalse AlertChannelFailRoute = 2
)

// AlertChannel 告警通道
type AlertChannel struct {
	common.BaseEntity

	Name      string                `json:"name" gorm:"column:name"`
	Type      AlertChannelType      `json:"type" gorm:"column:type"`
	Params    string                `json:"params" gorm:"column:params"`
	Handler   string                `json:"handler" gorm:"column:handler"`
	FailRoute AlertChannelFailRoute `json:"failRoute" gorm:"column:fail_route"`
	// Template 通道级告警模板；为空时回落到 alertConf 中该 Handler 的默认模板
	Template string `json:"template" gorm:"column:template"`
}

func (AlertChannel) TableName() string {
	return "t_alert_channel"
}
