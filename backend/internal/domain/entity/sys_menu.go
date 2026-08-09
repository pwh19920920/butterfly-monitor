package entity

import (
	"butterfly-monitor/internal/common"
)

type SysMenu struct {
	common.BaseEntity

	Code      string `json:"code" gorm:"column:code"`            // 菜单代码
	Name      string `json:"name" gorm:"column:name"`            // 菜单名称
	Path      string `json:"path" gorm:"column:path"`            // 菜单路径
	Icon      string `json:"icon" gorm:"column:icon"`            // 菜单图标
	Component string `json:"component" gorm:"column:component"`  // 菜单组件
	Sort      int32  `json:"sort" gorm:"column:sort"`            // 菜单排序
	Parent    *int64 `json:"parent,string" gorm:"column:parent"` // 上级目录
	Route     string `json:"route" gorm:"column:route"`          // 菜单路由
}

// TableName 会将 User 的表名重写为 `profiles`
func (SysMenu) TableName() string {
	return "t_sys_menu"
}
