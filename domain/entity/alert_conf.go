package entity

import (
	"github.com/pwh19920920/butterfly-admin/common"
)

type AlertConfType int32

const (
	AlertConfTypeNumber AlertConfType = 1
	AlertConfTypeString AlertConfType = 2
)

type AlertConf struct {
	common.BaseEntity

	ConfKey  string        `json:"confKey" gorm:"column:conf_key"`   // 报警间隔
	ConfVal  string        `json:"confVal" gorm:"column:conf_val"`   // 报警模板
	ConfType AlertConfType `json:"confType" gorm:"column:conf_type"` // 配置类型
	ConfDesc string        `json:"confDesc" gorm:"column:conf_desc"` // 配置描述
}

// TableName 会将 User 的表名重写为 `profiles`
func (AlertConf) TableName() string {
	return "t_alert_conf"
}
