package entity

import (
	"dragonfly-monitor/internal/common"
)

// VolatilityDayType 波动日类型
type VolatilityDayType int32

const (
	VolatilityDayTypePeak   VolatilityDayType = 1 // 高峰
	VolatilityDayTypeTrough VolatilityDayType = 2 // 低谷
)

// String 类型中文名（日志/展示用）
func (t VolatilityDayType) String() string {
	if t == VolatilityDayTypeTrough {
		return "低谷"
	}
	return "高峰"
}

// MonitorVolatilityDay 波动日（大促/节假日日历区间）
type MonitorVolatilityDay struct {
	common.BaseEntity

	Name      string            `json:"name" gorm:"column:name"`            // 波动日名称
	StartTime *common.LocalTime `json:"startTime" gorm:"column:start_time"` // 开始时间
	EndTime   *common.LocalTime `json:"endTime" gorm:"column:end_time"`     // 结束时间（含）
	Type      VolatilityDayType `json:"type,string" gorm:"column:type"`     // 1=高峰 2=低谷
}

func (MonitorVolatilityDay) TableName() string {
	return "t_monitor_volatility_day"
}
