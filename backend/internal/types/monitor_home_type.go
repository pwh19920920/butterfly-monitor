package types

import (
	"butterfly-monitor/internal/domain/entity"
)

// MonitorHomeCountResponse 首页统计
type MonitorHomeCountResponse struct {
	TaskCount              int64                      `json:"taskCount"`
	EventCount             int64                      `json:"eventCount"`
	DashboardCount         int64                      `json:"dashboardCount"`
	DatabaseCount          int64                      `json:"databaseCount"`
	AlertChannelCount      int64                      `json:"alertChannelCount"`
	AlertGroupCount        int64                      `json:"alertGroupCount"`
	PendingEvents          int64                      `json:"pendingEvents"`
	ProcessingEvents       int64                      `json:"processingEvents"`
	CompleteEvents         int64                      `json:"completeEvents"`
	IgnoreEvents           int64                      `json:"ignoreEvents"`
	AlertLevelDistribution map[int32]int64            `json:"alertLevelDistribution"`
	RecentEvents           []MonitorTaskEventListItem `json:"recentEvents"`
}

// MonitorTaskEventListItem 最近事件简项
type MonitorTaskEventListItem struct {
	Id         int64                             `json:"id"`
	TaskName   string                            `json:"taskName"`
	AlertMsg   string                            `json:"alertMsg"`
	DealStatus entity.MonitorTaskEventDealStatus `json:"dealStatus"`
	EventLevel int32                             `json:"eventLevel"`
	CreateTime string                            `json:"createTime"`
}
