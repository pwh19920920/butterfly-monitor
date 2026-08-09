package entity

import "butterfly-monitor/internal/common"

type MonitorTaskAlertStatus int32
type MonitorTaskAlertDealStatus int32

const (
	MonitorTaskAlertStatusNormal  MonitorTaskAlertStatus = 1
	MonitorTaskAlertStatusPending MonitorTaskAlertStatus = 2
	MonitorTaskAlertStatusFiring  MonitorTaskAlertStatus = 3

	MonitorTaskAlertDealStatusNormal     MonitorTaskAlertDealStatus = 1
	MonitorTaskAlertDealStatusProcessing MonitorTaskAlertDealStatus = 2
)

// MonitorTaskAlert 任务告警规则（一任务一条）
// DealStatus=处理中时不检测；FirstFlagTime 异常持续起点。
type MonitorTaskAlert struct {
	common.BaseEntity

	TaskId        int64                      `json:"taskId,string" gorm:"column:task_id"`
	AlertChannels string                     `json:"alertChannels" gorm:"column:alert_channels"`
	AlertGroups   string                     `json:"alertGroups" gorm:"column:alert_groups"`
	TimeSpan      int64                      `json:"timeSpan" gorm:"column:time_span"`
	Duration      int64                      `json:"duration" gorm:"column:duration"`
	Params        string                     `json:"params" gorm:"column:params"`
	AlertStatus   MonitorTaskAlertStatus     `json:"alertStatus" gorm:"column:alert_status"`
	DealStatus    MonitorTaskAlertDealStatus `json:"dealStatus" gorm:"column:deal_status"`
	PreCheckTime  *common.LocalTime          `json:"preCheckTime" gorm:"column:pre_check_time"`
	FirstFlagTime *common.LocalTime          `json:"firstFlagTime" gorm:"column:first_flag_time"`
}

type MonitorAlertCheckParamsRelation int32
type MonitorAlertCheckParamsCompareType int32
type MonitorAlertCheckParamsValueType int32
type MonitorAlertCheckParamsLevelType int32

const (
	// MonitorAlertCheckParamsValueTypePercent 比较值类型
	// 1 样本差阈值百分比：(实时-样本)/样本*100
	// 2 实时数值比较：直接用实时值（唯一不需要样本的类型）
	// 3 样本差阈值比较：实时-样本
	MonitorAlertCheckParamsValueTypePercent       MonitorAlertCheckParamsValueType = 1
	MonitorAlertCheckParamsValueTypeAbsoluteValue MonitorAlertCheckParamsValueType = 2
	MonitorAlertCheckParamsValueTypeValue         MonitorAlertCheckParamsValueType = 3

	// MonitorAlertCheckParamsRelationOr 组内条件关系（组与组之间固定 OR）
	MonitorAlertCheckParamsRelationOr  MonitorAlertCheckParamsRelation = 1
	MonitorAlertCheckParamsRelationAnd MonitorAlertCheckParamsRelation = 2

	// MonitorAlertCheckParamsCompareTypeGt 比较方式
	MonitorAlertCheckParamsCompareTypeGt  MonitorAlertCheckParamsCompareType = 1
	MonitorAlertCheckParamsCompareTypeLt  MonitorAlertCheckParamsCompareType = 2
	MonitorAlertCheckParamsCompareTypeEq  MonitorAlertCheckParamsCompareType = 3
	MonitorAlertCheckParamsCompareTypeEgt MonitorAlertCheckParamsCompareType = 4
	MonitorAlertCheckParamsCompareTypeElt MonitorAlertCheckParamsCompareType = 5

	// MonitorAlertCheckParamsLevelNormal 等级
	MonitorAlertCheckParamsLevelNormal   MonitorAlertCheckParamsLevelType = -1
	MonitorAlertCheckParamsLevelCritical MonitorAlertCheckParamsLevelType = 0
	MonitorAlertCheckParamsLevelHigh     MonitorAlertCheckParamsLevelType = 1
	MonitorAlertCheckParamsLevelMedium   MonitorAlertCheckParamsLevelType = 2
	MonitorAlertCheckParamsLevelLow      MonitorAlertCheckParamsLevelType = 3
)

type MonitorAlertCheckParamsItem struct {
	ValueType   MonitorAlertCheckParamsValueType   `json:"valueType"`
	Value       float64                            `json:"value"`
	CompareType MonitorAlertCheckParamsCompareType `json:"compareType"`
}

func (compareType MonitorAlertCheckParamsCompareType) GetTransferMsg() string {
	switch compareType {
	case MonitorAlertCheckParamsCompareTypeGt:
		return "高于"
	case MonitorAlertCheckParamsCompareTypeLt:
		return "低于"
	case MonitorAlertCheckParamsCompareTypeEq:
		return "等于"
	case MonitorAlertCheckParamsCompareTypeElt:
		return "小于等于"
	case MonitorAlertCheckParamsCompareTypeEgt:
		return "大于等于"
	}
	return ""
}

func (valueType MonitorAlertCheckParamsValueType) GetTransferMsg() string {
	switch valueType {
	case MonitorAlertCheckParamsValueTypePercent:
		return "样本差阈值百分比"
	case MonitorAlertCheckParamsValueTypeAbsoluteValue:
		return "实时数值比较"
	case MonitorAlertCheckParamsValueTypeValue:
		return "样本差阈值比较"
	}
	return ""
}

// MonitorAlertCheckParams 规则组
type MonitorAlertCheckParams struct {
	Relation    MonitorAlertCheckParamsRelation   `json:"relation"`
	EffectTimes []string                          `json:"effectTimes"`
	Rules       []MonitorAlertCheckParamsItem     `json:"rules"`
	Level       *MonitorAlertCheckParamsLevelType `json:"level"`
}

func (MonitorTaskAlert) TableName() string {
	return "t_monitor_task_alert"
}
