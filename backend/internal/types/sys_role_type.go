package types

import (
	"dragonfly-monitor/internal/domain/entity"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/pkg/response"
)

type SysRoleQueryRequest struct {
	response.RequestPaging
}

type SysRolePermissionQueryResponse struct {
	entity.SysPermission
	Options []string `json:"options"`
}

type SysRoleCreateRequest struct {
	entity.SysRole
	Permissions []entity.SysPermission `json:"permissions"`
}

func (req SysRoleCreateRequest) ValidateForCreate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(0, 255)),
	)
}

func (req SysRoleCreateRequest) ValidateForModify() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(0, 255)),
		validation.Field(&req.Id, validation.Required),
	)
}
