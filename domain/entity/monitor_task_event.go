package entity

import "github.com/pwh19920920/butterfly-admin/common"

type MonitorTaskEventDealStatus int32

const (
	MonitorTaskEventDealStatusPending    MonitorTaskEventDealStatus = 1
	MonitorTaskEventDealStatusProcessing MonitorTaskEventDealStatus = 2
	MonitorTaskEventDealStatusComplete   MonitorTaskEventDealStatus = 3
	MonitorTaskEventDealStatusIgnore     MonitorTaskEventDealStatus = 4
)

type MonitorTaskEvent struct {
	common.BaseEntity

	AlertId       int64                      `json:"alertId,string" gorm:"column:alert_id"`       // 报警id
	TaskId        int64                      `json:"taskId,string" gorm:"column:task_id"`         // 任务id
	AlertMsg      string                     `json:"alertMsg" gorm:"column:alert_msg"`            // 报警信息
	DealTime      *common.LocalTime          `json:"dealTime" gorm:"column:deal_time"`            // 处理时间
	CompleteTime  *common.LocalTime          `json:"completeTime" gorm:"column:complete_time"`    // 完成时间
	Content       string                     `json:"content" gorm:"column:content"`               // 事件经过
	DealStatus    MonitorTaskEventDealStatus `json:"dealStatus" gorm:"column：deal_status"`        // 处理状态
	DealUser      *int64                     `json:"dealUser" gorm:"column:deal_user"`            // 处理用户
	PreAlertTime  *common.LocalTime          `json:"preAlertTime" gorm:"column:pre_alert_time"`   // 上次报警事件
	NextAlertTime *common.LocalTime          `json:"nextAlertTime" gorm:"column:next_alert_time"` // 下次报警事件，用于扫描报警, 首次加个1分钟
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorTaskEvent) TableName() string {
	return "t_monitor_task_event"
}
