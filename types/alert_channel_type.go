package types

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly/response"
)

type AlertChannelQueryRequest struct {
	response.RequestPaging
}

type AlertChannelCreateRequest struct {
	entity.AlertChannel
}

type AlertModifyModifyRequest struct {
	entity.AlertChannel
}
