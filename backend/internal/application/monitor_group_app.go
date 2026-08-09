package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// MonitorGroupApplication 监控分组应用服务
type MonitorGroupApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewMonitorGroupApplication 创建监控分组应用服务
func NewMonitorGroupApplication(sequence *snowflake.Node, repository *persistence.Repository) MonitorGroupApplication {
	return MonitorGroupApplication{sequence: sequence, repository: repository}
}

// Query 分页查询
func (app *MonitorGroupApplication) Query(ctx context.Context, req *types.MonitorGroupQueryRequest) (int64, []entity.MonitorGroup, error) {
	total, data, err := app.repository.MonitorGroupRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "MonitorGroupRepository.Select() happen error for", err)
		return total, nil, err
	}
	return total, data, nil
}

// QueryAll 全量查询
func (app *MonitorGroupApplication) QueryAll(ctx context.Context) ([]entity.MonitorGroup, error) {
	data, err := app.repository.MonitorGroupRepository.SelectAll()
	if err != nil {
		logger.Error(ctx, "MonitorGroupRepository.SelectAll() happen error for", err)
	}
	return data, err
}

// Create 创建分组
func (app *MonitorGroupApplication) Create(ctx context.Context, group *entity.MonitorGroup) error {
	if group.Name == "" {
		return errors.New("分组名称不能为空")
	}
	group.Id = app.sequence.Generate().Int64()
	if group.Parent == 0 {
		group.Route = fmt.Sprintf("/%d/", group.Id)
		return app.repository.MonitorGroupRepository.Save(group)
	}

	parent, err := app.repository.MonitorGroupRepository.GetById(group.Parent)
	if err != nil {
		return err
	}

	if parent == nil {
		return errors.New("上级分组不存在")
	}
	group.Route = fmt.Sprintf("%s%d/", parent.Route, group.Id)
	return app.repository.MonitorGroupRepository.Save(group)
}

// Modify 修改分组（重算 Route + 级联）
func (app *MonitorGroupApplication) Modify(ctx context.Context, group *entity.MonitorGroup) error {
	if group.Id == 0 {
		return errors.New("id 不能为空")
	}
	old, err := app.repository.MonitorGroupRepository.GetById(group.Id)
	if err != nil {
		return err
	}
	if old == nil {
		return errors.New("分组不存在")
	}
	oldRoute := old.Route

	if group.Parent == 0 {
		group.Route = fmt.Sprintf("/%d/", group.Id)
		return app.repository.MonitorGroupRepository.Modify(group.Id, oldRoute, group)
	}

	if group.Parent == group.Id {
		return errors.New("不能将自身设为上级")
	}
	parent, err := app.repository.MonitorGroupRepository.GetById(group.Parent)
	if err != nil {
		return err
	}
	if parent == nil {
		return errors.New("上级分组不存在")
	}
	if strings.HasPrefix(parent.Route, oldRoute) {
		return errors.New("不能将子节点设为上级")
	}
	group.Route = fmt.Sprintf("%s%d/", parent.Route, group.Id)
	return app.repository.MonitorGroupRepository.Modify(group.Id, oldRoute, group)
}
