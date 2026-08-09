package repository

import "butterfly-monitor/internal/domain/entity"

type AlertGroupUserRepository interface {
	SelectByGroupId(groupId int64) ([]entity.AlertGroupUser, error)
	SelectUsersByUserIds(userIds []int64) ([]entity.SysUser, error)
}
