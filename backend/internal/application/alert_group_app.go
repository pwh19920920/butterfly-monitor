package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// AlertGroupApplication 告警分组应用服务
type AlertGroupApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewAlertGroupApplication 创建告警分组应用服务
func NewAlertGroupApplication(sequence *snowflake.Node, repository *persistence.Repository) AlertGroupApplication {
	return AlertGroupApplication{sequence: sequence, repository: repository}
}

// Query 分页查询
func (app *AlertGroupApplication) Query(ctx context.Context, req *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error) {
	total, data, err := app.repository.AlertGroupRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "AlertGroupRepository.Select() happen error for", err)
		return total, nil, err
	}
	return total, data, nil
}

// QueryAll 全量查询
func (app *AlertGroupApplication) QueryAll(ctx context.Context) ([]entity.AlertGroup, error) {
	data, err := app.repository.AlertGroupRepository.SelectAll()
	if err != nil {
		logger.Error(ctx, "AlertGroupRepository.SelectAll() happen error for", err)
	}
	return data, err
}

// QueryGroupUsers 查分组下用户 id 列表
func (app *AlertGroupApplication) QueryGroupUsers(ctx context.Context, groupId int64) ([]string, error) {
	users, err := app.repository.AlertGroupUserRepository.SelectByGroupId(groupId)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(users))
	for _, u := range users {
		result = append(result, strconv.FormatInt(u.UserId, 10))
	}
	return result, nil
}

// Create 创建分组
func (app *AlertGroupApplication) Create(ctx context.Context, req *types.AlertGroupSaveRequest) error {
	if req.Name == "" {
		return errors.New("分组名称不能为空")
	}
	if len(req.UserIds) == 0 {
		return errors.New("成员不能为空")
	}
	groupId := app.sequence.Generate().Int64()
	group := &entity.AlertGroup{
		BaseEntity: common.BaseEntity{Id: groupId},
		Name:       req.Name,
	}
	users, err := app.buildGroupUsers(groupId, req.UserIds)
	if err != nil {
		return err
	}
	return app.repository.AlertGroupRepository.Save(group, users)
}

// Modify 修改分组
func (app *AlertGroupApplication) Modify(ctx context.Context, req *types.AlertGroupSaveRequest) error {
	if req.Id == 0 {
		return errors.New("id 不能为空")
	}
	if len(req.UserIds) == 0 {
		return errors.New("成员不能为空")
	}
	group := &entity.AlertGroup{
		BaseEntity: common.BaseEntity{Id: req.Id},
		Name:       req.Name,
	}
	users, err := app.buildGroupUsers(req.Id, req.UserIds)
	if err != nil {
		return err
	}
	return app.repository.AlertGroupRepository.Modify(req.Id, group, users)
}

func (app *AlertGroupApplication) buildGroupUsers(groupId int64, userIds []string) ([]entity.AlertGroupUser, error) {
	users := make([]entity.AlertGroupUser, 0, len(userIds))
	for _, uidStr := range userIds {
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			return nil, errors.New("用户 id 格式错误")
		}
		users = append(users, entity.AlertGroupUser{
			BaseEntity: common.BaseEntity{Id: app.sequence.Generate().Int64()},
			UserId:     uid,
			GroupId:    groupId,
		})
	}
	return users, nil
}

// Count 统计分组总数
func (app *AlertGroupApplication) Count(ctx context.Context) (int64, error) {
	return app.repository.AlertGroupRepository.Count()
}
