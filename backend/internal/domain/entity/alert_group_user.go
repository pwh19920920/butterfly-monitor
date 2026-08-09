package entity

import "butterfly-monitor/internal/common"

// AlertGroupUser 告警组-用户关联
type AlertGroupUser struct {
	common.BaseEntity

	UserId  int64 `json:"userId,string" gorm:"column:user_id"`
	GroupId int64 `json:"groupId,string" gorm:"column:group_id"`
}

func (AlertGroupUser) TableName() string {
	return "t_alert_group_user"
}
