package repository

import sysEntity "github.com/pwh19920920/butterfly-admin/domain/entity"

type AlertGroupUserRepository interface {

	// SelectByGroupId 查询全部
	SelectByGroupId(groupId int64) ([]int64, error)

	// SelectUsersByUserIds 查询用户列表
	SelectUsersByUserIds(userIds []int64) ([]sysEntity.SysUser, error)
}
