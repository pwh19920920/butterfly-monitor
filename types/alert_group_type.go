package types

import (
	"butterfly-monitor/domain/entity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pwh19920920/butterfly/response"
	"strconv"
)

type AlertGroupQueryRequest struct {
	response.RequestPaging
}

type AlertGroupRequestGroupUser int64

type AlertGroupCreateRequest struct {
	entity.AlertGroup
	GroupUsers      []string `json:"groupUsers"` // 用户列表
	GroupUserInt64s []int64  `json:"-"`
}

func (req *AlertGroupCreateRequest) GetGroupUserInt64s() []int64 {
	groupUserIds := make([]int64, 0)
	for _, groupUser := range req.GroupUsers {
		groupUserId, _ := strconv.ParseInt(groupUser, 10, 64)
		groupUserIds = append(groupUserIds, groupUserId)
	}
	return groupUserIds
}

type AlertGroupModifyRequest struct {
	entity.AlertGroup
	GroupUsers      []string `json:"groupUsers"` // 用户列表
	GroupUserInt64s []int64  `json:"-"`
}

func (req *AlertGroupModifyRequest) GetGroupUserInt64s() []int64 {
	groupUserIds := make([]int64, 0)
	for _, groupUser := range req.GroupUsers {
		groupUserId, _ := strconv.ParseInt(groupUser, 10, 64)
		groupUserIds = append(groupUserIds, groupUserId)
	}
	return groupUserIds
}

func (req AlertGroupCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.GroupUsers, validation.Required),
	)
}

func (req AlertGroupModifyRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Id, validation.Required),
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.GroupUsers, validation.Required),
	)
}
