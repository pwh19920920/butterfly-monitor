package application

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/types"
	"errors"
	"github.com/pwh19920920/butterfly-admin/common"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
)

type AlertGroupApplication struct {
	repository *persistence.Repository
	sequence   *snowflake.Node
}

// Query 分页查询
func (application *AlertGroupApplication) Query(request *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error) {
	total, data, err := application.repository.AlertGroupRepository.Select(request)

	// 错误记录
	if err != nil {
		logrus.Errorf("AlertGroupRepository.Select() happen error for %s", err.Error())
	}
	return total, data, err
}

// QueryGroupUserIdsByGroupId 查询分组下的用户
func (application *AlertGroupApplication) QueryGroupUserIdsByGroupId(groupId int64) ([]int64, error) {
	return application.repository.AlertGroupUserRepository.SelectByGroupId(groupId)
}

// QueryAll 查询全部
func (application *AlertGroupApplication) QueryAll() ([]entity.AlertGroup, error) {
	return application.repository.AlertGroupRepository.SelectAll()
}

// Save 保存
func (application *AlertGroupApplication) Save(req *types.AlertGroupCreateRequest) error {
	if req.GroupUsers == nil || len(req.GroupUsers) == 0 {
		return errors.New("分组用户不能为空")
	}
	group := req.AlertGroup
	group.Id = application.sequence.Generate().Int64()
	groupUsers := application.coverCreateModifyData(group, req.GetGroupUserInt64s())
	return application.repository.AlertGroupRepository.Save(&group, groupUsers)
}

// Modify 保存
func (application *AlertGroupApplication) Modify(req *types.AlertGroupModifyRequest) error {
	if req.GroupUsers == nil || len(req.GroupUsers) == 0 {
		return errors.New("分组用户不能为空")
	}
	group := req.AlertGroup
	groupUsers := application.coverCreateModifyData(group, req.GetGroupUserInt64s())
	return application.repository.AlertGroupRepository.Modify(req.Id, &group, groupUsers)
}

// coverCreateModifyData
func (application *AlertGroupApplication) coverCreateModifyData(group entity.AlertGroup, userIds []int64) []entity.AlertGroupUser {
	// 组装数据
	groupUsers := make([]entity.AlertGroupUser, 0)
	for _, user := range userIds {
		groupUsers = append(groupUsers, entity.AlertGroupUser{
			BaseEntity: common.BaseEntity{Id: application.sequence.Generate().Int64()},
			GroupId:    group.Id,
			UserId:     user,
		})
	}
	return groupUsers
}
