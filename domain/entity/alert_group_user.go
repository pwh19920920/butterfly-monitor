package entity

import (
	"github.com/pwh19920920/butterfly-admin/common"
)

type AlertGroupUser struct {
	common.BaseEntity

	UserId  int64 `json:"userId" gorm:"column:user_id"`   // 用户
	GroupId int64 `json:"groupId" gorm:"column:group_id"` // 分组
}

// TableName 会将 User 的表名重写为 `profiles`
func (AlertGroupUser) TableName() string {
	return "t_alert_group_user"
}
