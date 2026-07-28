package entity

import "dragonfly-monitor/internal/common"

type AlertConfType int32

const (
	AlertConfTypeNumber AlertConfType = 1
	AlertConfTypeString AlertConfType = 2
)

// AlertConf 告警全局 KV 配置
type AlertConf struct {
	common.BaseEntity

	ConfKey  string        `json:"confKey" gorm:"column:conf_key"`
	ConfVal  string        `json:"confVal" gorm:"column:conf_val"`
	ConfDesc string        `json:"confDesc" gorm:"column:conf_desc"`
	ConfType AlertConfType `json:"confType" gorm:"column:conf_type"`
}

func (AlertConf) TableName() string {
	return "t_alert_conf"
}
