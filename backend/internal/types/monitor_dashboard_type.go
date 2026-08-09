package types

import (
	"butterfly-monitor/internal/domain/entity"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type MonitorDashboardQueryRequest struct {
	response.RequestPaging
	Name string `form:"name"`
}

type MonitorDashboardTaskSortRequest struct {
	Items []MonitorDashboardTaskSortItem `json:"items"`
}

type MonitorDashboardTaskSortItem struct {
	Id   int64 `json:"id,string"`
	Sort int32 `json:"sort"`
}

type MonitorDashboardQueryResponse struct {
	entity.MonitorDashboard
}

// MonitorDashboardTaskResponse 面板下任务关联（补 TaskName，供排序展示）
type MonitorDashboardTaskResponse struct {
	entity.MonitorDashboardTask
	TaskName string `json:"taskName"`
}
