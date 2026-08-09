package types

import (
	"butterfly-monitor/internal/domain/entity"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// MonitorVolatilityDayBatchCreateRequest 批量添加请求
// Name + Type 共用，Items 为多行日期区间
type MonitorVolatilityDayBatchCreateRequest struct {
	Name  string                        `json:"name"`
	Type  entity.VolatilityDayType      `json:"type,string"`
	Items []entity.MonitorVolatilityDay `json:"items"`
}

// ValidateForCreate 校验
func (req MonitorVolatilityDayBatchCreateRequest) ValidateForCreate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Type, validation.Required, validation.In(
			entity.VolatilityDayTypePeak, entity.VolatilityDayTypeTrough,
		)),
		validation.Field(&req.Items, validation.Required, validation.Length(1, 0)),
	)
}
