package types

import (
	"butterfly-monitor/domain/entity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/response"
)

type AlertConfModifyRequest struct {
	entity.AlertConf
}

type AlertConfQueryRequest struct {
	response.RequestPaging
}

type AlertConfCreateRequest struct {
	entity.AlertConf
}

func (req AlertConfModifyRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ConfVal, validation.Required),
		validation.Field(&req.ConfKey, validation.Required),
		validation.Field(&req.ConfDesc, validation.Required),
		validation.Field(&req.ConfType, validation.Required),
	)
}

func (req AlertConfCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Id, validation.Required),
		validation.Field(&req.ConfVal, validation.Required),
		validation.Field(&req.ConfKey, validation.Required),
		validation.Field(&req.ConfDesc, validation.Required),
		validation.Field(&req.ConfType, validation.Required),
	)
}
