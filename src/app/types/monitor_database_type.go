package types

import (
	"butterfly-monitor/src/app/domain/entity"
	"github.com/pwh19920920/butterfly/response"
)

type MonitorDatabaseQueryRequest struct {
	response.RequestPaging

	Name string                `form:"name"`
	Type entity.DataSourceType `form:"type"`
}

type MonitorDatabaseCreateRequest struct {
	entity.MonitorDatabase
}
