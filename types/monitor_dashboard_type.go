package types

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly/response"
)

type MonitorDashboardQueryRequest struct {
	response.RequestPaging
}

type MonitorDashboardCreateRequest struct {
	entity.MonitorDashboard
}

type MonitorDashboardQueryTaskResponse struct {
	entity.MonitorDashboardTask
	TaskName string `json:"taskName"`
	TaskKey  string `json:"taskKey"`
}

type MonitorDashboardTaskModifyRequest struct {
	Data []entity.MonitorDashboardTask `json:"data"`
}
