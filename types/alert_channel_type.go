package types

import (
	"butterfly-monitor/domain/entity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/response"
)

type AlertChannelQueryRequest struct {
	response.RequestPaging
}

type AlertChannelTestParams struct {
	Template string `json:"template"`
	Email    string `json:"email"`
}

type AlertChannelCreateRequest struct {
	entity.AlertChannel
	TestParams AlertChannelTestParams `json:"testParams"`
}

type AlertChannelModifyRequest struct {
	entity.AlertChannel
	TestParams AlertChannelTestParams `json:"testParams"`
}

type AlertChannelHandlerResponse struct {
	ChannelType entity.AlertChannelType `json:"channelType"`
	Handlers    []string                `json:"handlers"`
}

func (req AlertChannelCreateRequest) Validate() error {
	valid := []*validation.FieldRules{
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Type, validation.Required),
		validation.Field(&req.Params, validation.Required),
		validation.Field(&req.Handler, validation.Required),
		validation.Field(&req.FailRoute, validation.Required),
		validation.Field(&req.TestParams, validation.Required),
	}
	return validation.ValidateStruct(&req, valid...)
}

func (req AlertChannelModifyRequest) Validate() error {
	valid := []*validation.FieldRules{
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Type, validation.Required),
		validation.Field(&req.Params, validation.Required),
		validation.Field(&req.Handler, validation.Required),
		validation.Field(&req.FailRoute, validation.Required),
		validation.Field(&req.TestParams, validation.Required),
	}
	return validation.ValidateStruct(&req, valid...)
}
