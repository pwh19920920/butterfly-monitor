package types

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly/response"
)

type MonitorTaskEventQueryRequest struct {
	response.RequestPaging
	DealStatus *entity.MonitorTaskEventDealStatus `form:"dealStatus"`
	CreatedAts []string                           `form:"createdAts"`
}

type MonitorTaskEventProcessRequest struct {
	AlertId  int64  `json:"alertId,string"` // 报警id
	TaskId   int64  `json:"taskId,string"`  // 任务id
	Content  string `json:"content"`        // 事件经过
	AlertMsg string `json:"alertMsg"`       // 报警信息
	DealUser *int64 `json:"dealUser"`       // 处理人
}

type MonitorTaskEventQueryResponse struct {
	entity.MonitorTaskEvent
	TaskName     string `json:"taskName"`     // 任务名称
	DealUserName string `json:"dealUserName"` // 处理人名称
}
