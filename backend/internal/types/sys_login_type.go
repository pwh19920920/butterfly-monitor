package types

import (
	"butterfly-monitor/internal/domain/entity"
)
import "github.com/go-ozzo/ozzo-validation/v4"

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (req LoginForm) ValidateForLogin() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Username, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.Password, validation.Required, validation.Length(0, 255)),
	)
}

type SysMenuPermissionForUser struct {
	entity.SysUser
	Menus       []SysMenuPermissionForUserMenu `json:"menus"`
	Permissions []string                       `json:"permissions"`
	Codes       []string                       `json:"codes"`
}

type SysMenuPermissionForUserMenu struct {
	Id        int64                          `json:"id"`        // id
	Name      string                         `json:"name"`      // 菜单名称
	Path      string                         `json:"path"`      // 菜单路径
	Icon      string                         `json:"icon"`      // 菜单图标
	Component string                         `json:"component"` // 菜单组件
	Routes    []SysMenuPermissionForUserMenu `json:"routes"`
}
