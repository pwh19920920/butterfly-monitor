package types

import (
	"dragonfly-monitor/internal/domain/entity"

	"github.com/pwh19920920/butterfly/pkg/response"
)

type MonitorTaskEventQueryRequest struct {
	response.RequestPaging
	TaskId     *int64                             `form:"taskId"`
	DealStatus *entity.MonitorTaskEventDealStatus `form:"dealStatus"`
	EventLevel *int32                             `form:"eventLevel"` // 事件等级
	StartTime  *string                            `form:"startTime"`  // 区间起 yyyy-MM-dd HH:mm:ss
	EndTime    *string                            `form:"endTime"`    // 区间止 yyyy-MM-dd HH:mm:ss
}

type MonitorTaskEventQueryResponse struct {
	entity.MonitorTaskEvent
	TaskName     string `json:"taskName"`
	DealUserName string `json:"dealUserName"`
}

type MonitorTaskEventProcessRequest struct {
	TaskId   int64  `json:"taskId,string"`
	AlertId  int64  `json:"alertId,string"`
	Content  string `json:"content"`
	DealUser *int64 `json:"dealUser,string"`
}
