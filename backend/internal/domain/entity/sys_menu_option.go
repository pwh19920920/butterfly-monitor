package entity

import (
	"dragonfly-monitor/internal/common"
)

type SysMenuOption struct {
	common.BaseEntity

	Name   string `json:"name" gorm:"name"`             // 操作名称
	Value  string `json:"value" gorm:"value"`           // 操作权限
	Method string `json:"method" gorm:"method"`         // URL方法
	Path   string `json:"path" gorm:"path"`             // URL地址
	MenuId int64  `json:"menuId,string" gorm:"menu_id"` // 菜单id
	Code   string `json:"code" gorm:"code"`             // 唯一码
}

// TableName 会将 User 的表名重写为 `profiles`
func (SysMenuOption) TableName() string {
	return "t_sys_menu_option"
}
