package entity

import "github.com/pwh19920920/butterfly-admin/common"

type MonitorTaskAlertStatus int32
type MonitorTaskAlertDealStatus int32

const (
	MonitorTaskAlertStatusNormal  MonitorTaskAlertStatus = 1
	MonitorTaskAlertStatusPending MonitorTaskAlertStatus = 2
	MonitorTaskAlertStatusFiring  MonitorTaskAlertStatus = 3

	MonitorTaskAlertDealStatusNormal MonitorTaskAlertDealStatus = 1
	MonitorTaskAlertDealStatusHandle MonitorTaskAlertDealStatus = 2
)

// MonitorTaskAlert 逻辑注意点：
// 当DealStatus为处理中, 不进行检测, 表示人工处理中, 处理完成, 需要更新 AlertStatus, DealStatus, PreCheckTime, FirstFlagTime
// 当date_add(now(), interval -time_span second) < pre_check_time 表示还没到达下一次检测时间
// FirstFlagTime首次标记时间, 如果未出现异常, 则此值持续更新, 如果出现异常, 则这个值不再更新
type MonitorTaskAlert struct {
	common.BaseEntity

	TaskId        int64                      `json:"taskId" gorm:"column:task_id"`                // 任务id
	AlertChannels string                     `json:"alertChannels" gorm:"column:alert_channels"`  // 报警渠道列表
	AlertGroups   string                     `json:"alertGroups" gorm:"column:alert_groups"`      // 报警组列表
	EffectTime    *string                    `json:"effectTime" gorm:"column:effect_time"`        // 生效时间
	TimeSpan      int64                      `json:"timeSpan" gorm:"column:time_span"`            // 检查隔间
	Duration      int64                      `json:"duration" gorm:"column:duration"`             // 持续时间, s为单位
	Params        string                     `json:"params" gorm:"column:params"`                 // 规则参数：[{比较方式，值，关系，比较值类型}]
	AlertStatus   MonitorTaskAlertStatus     `json:"alertStatus" gorm:"column:alert_status"`      // 报警状态：1正常，2出现异常，3达到报警条件
	DealStatus    MonitorTaskAlertDealStatus `json:"dealStatus" gorm:"column:deal_status"`        // 处理状态：1正常，2处理中
	PreCheckTime  *common.LocalTime          `json:"preCheckTime" gorm:"column:pre_check_time"`   // 上一次检查时间
	FirstFlagTime *common.LocalTime          `json:"firstFlagTime" gorm:"column:first_flag_time"` // 首次标记时间, 如果未出现异常, 则此值持续更新, 如果出现异常, 则这个值不再更新
}

// TableName 会将 User 的表名重写为 `profiles`
func (MonitorTaskAlert) TableName() string {
	return "t_monitor_task_alert"
}
