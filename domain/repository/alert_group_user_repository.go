package repository

type AlertGroupUserRepository interface {

	// SelectByGroupId 查询全部
	SelectByGroupId(groupId int64) ([]int64, error)
}
