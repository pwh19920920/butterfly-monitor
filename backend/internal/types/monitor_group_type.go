package types

import (
	"github.com/pwh19920920/butterfly/pkg/response"
)

type MonitorGroupQueryRequest struct {
	response.RequestPaging
	Name string `form:"name"`
}
