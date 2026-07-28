package types

import (
	"dragonfly-monitor/internal/domain/entity"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type MonitorDatabaseQueryRequest struct {
	response.RequestPaging
	Name string                `form:"name"`
	Type entity.DataSourceType `form:"type"`
}
