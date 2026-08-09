package entity

import (
	"butterfly-monitor/internal/common"
)

type SysPermission struct {
	common.BaseEntity

	RoleId      int64  `json:"roleId,string" gorm:"column:role_id"`   // 角色
	MenuId      int64  `json:"menuId,string" gorm:"column:menu_id"`   // 菜单
	Option      string `json:"option" gorm:"column:option"`           // 操作
	Independent bool   `json:"independent" gorm:"column:independent"` // 是否独立
	Half        bool   `json:"half" gorm:"half"`                      // 是否半选
	Root        bool   `json:"root" gorm:"root"`                      // 是否根
}

// TableName 会将 User 的表名重写为 `profiles`
func (SysPermission) TableName() string {
	return "t_sys_permission"
}
