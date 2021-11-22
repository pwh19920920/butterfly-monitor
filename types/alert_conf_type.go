package types

import (
	"butterfly-monitor/domain/entity"
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
