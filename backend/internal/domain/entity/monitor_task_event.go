package entity

import "butterfly-monitor/internal/common"

type MonitorTaskEventDealStatus int32

const (
	MonitorTaskEventDealStatusPending    MonitorTaskEventDealStatus = 1
	MonitorTaskEventDealStatusProcessing MonitorTaskEventDealStatus = 2
	MonitorTaskEventDealStatusComplete   MonitorTaskEventDealStatus = 3
	MonitorTaskEventDealStatusIgnore     MonitorTaskEventDealStatus = 4
)

// MonitorTaskEvent 告警事件
type MonitorTaskEvent struct {
	common.BaseEntity

	AlertId       int64                      `json:"alertId,string" gorm:"column:alert_id"`
	TaskId        int64                      `json:"taskId,string" gorm:"column:task_id"`
	AlertMsg      string                     `json:"alertMsg" gorm:"column:alert_msg"`
	DealTime      *common.LocalTime          `json:"dealTime" gorm:"column:deal_time"`
	CompleteTime  *common.LocalTime          `json:"completeTime" gorm:"column:complete_time"`
	Content       string                     `json:"content" gorm:"column:content"`
	DealStatus    MonitorTaskEventDealStatus `json:"dealStatus" gorm:"column:deal_status"`
	DealUser      *int64                     `json:"dealUser,string" gorm:"column:deal_user"`
	PreAlertTime  *common.LocalTime          `json:"preAlertTime" gorm:"column:pre_alert_time"`
	NextAlertTime *common.LocalTime          `json:"nextAlertTime" gorm:"column:next_alert_time"`
	EventLevel    int32                      `json:"eventLevel" gorm:"column:event_level"`
	AlertCount    int32                      `json:"alertCount" gorm:"column:alert_count"` // 已成功告警次数，用于间隔升级
}

func (MonitorTaskEvent) TableName() string {
	return "t_monitor_task_event"
}
