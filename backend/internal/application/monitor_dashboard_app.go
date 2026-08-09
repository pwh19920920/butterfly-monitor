package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/config/grafana"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/infrastructure/support"
	"butterfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// MonitorDashboardApplication 监控面板应用服务
type MonitorDashboardApplication struct {
	sequence       *snowflake.Node
	repository     *persistence.Repository
	grafanaHandler *support.GrafanaHandler
	grafanaConf    *grafana.Config
}

// NewMonitorDashboardApplication 创建监控面板应用服务
func NewMonitorDashboardApplication(sequence *snowflake.Node, repository *persistence.Repository, grafanaHandler *support.GrafanaHandler, grafanaConf *grafana.Config) MonitorDashboardApplication {
	return MonitorDashboardApplication{sequence: sequence, repository: repository, grafanaHandler: grafanaHandler, grafanaConf: grafanaConf}
}

// Query 分页查询，Url 前拼 Grafana.Addr
func (app *MonitorDashboardApplication) Query(ctx context.Context, req *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error) {
	total, data, err := app.repository.MonitorDashboardRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "MonitorDashboardRepository.Select() happen error for", err)
		return total, nil, err
	}
	for i := range data {
		data[i].Url = app.prefixUrl(data[i].Url)
	}
	return total, data, nil
}

// QueryAll 全量查询（下拉）
func (app *MonitorDashboardApplication) QueryAll(ctx context.Context) ([]entity.MonitorDashboard, error) {
	data, err := app.repository.MonitorDashboardRepository.SelectSimpleAll()
	if err != nil {
		logger.Error(ctx, "MonitorDashboardRepository.SelectSimpleAll() happen error for", err)
	}
	return data, err
}

// Create 创建面板：调 Grafana 回填 Url/Slug/Uid/BoardId
func (app *MonitorDashboardApplication) Create(ctx context.Context, dashboard *entity.MonitorDashboard) error {
	if dashboard.Name == "" {
		return errors.New("面板名称不能为空")
	}
	dashboard.Id = app.sequence.Generate().Int64()

	if app.grafanaHandler != nil {
		url, slug, uid, boardId, err := app.grafanaHandler.CreateDashboard(ctx, dashboard.Name)
		if err != nil {
			return err
		}
		dashboard.Url = url
		dashboard.Slug = slug
		dashboard.Uid = uid
		if boardId != 0 {
			dashboard.BoardId = &boardId
		}
	}
	return app.repository.MonitorDashboardRepository.Save(dashboard)
}

// Modify 修改面板名称
func (app *MonitorDashboardApplication) Modify(ctx context.Context, dashboard *entity.MonitorDashboard) error {
	if dashboard.Id == 0 {
		return errors.New("id 不能为空")
	}
	old, err := app.repository.MonitorDashboardRepository.GetById(dashboard.Id)
	if err != nil {
		return err
	}
	if old == nil {
		return errors.New("面板不存在")
	}
	if app.grafanaHandler != nil && old.Uid != "" && dashboard.Name != "" {
		if err := app.grafanaHandler.ModifyDashboardName(old.Uid, dashboard.Name); err != nil {
			return err
		}
	}
	return app.repository.MonitorDashboardRepository.UpdateById(dashboard.Id, dashboard)
}

// QueryTasks 查面板下任务关联，补 TaskName 供排序展示
func (app *MonitorDashboardApplication) QueryTasks(ctx context.Context, dashboardId int64) ([]types.MonitorDashboardTaskResponse, error) {
	relations, err := app.repository.MonitorDashboardTaskRepository.FindByDashboardId(dashboardId)
	if err != nil {
		logger.Error(ctx, "MonitorDashboardTaskRepository.FindByDashboardId() happen error for", err)
		return nil, err
	}
	// 批量取任务名
	taskIds := make([]int64, 0, len(relations))
	for _, r := range relations {
		taskIds = append(taskIds, r.TaskId)
	}
	taskMap, err := app.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)
	if err != nil {
		return nil, err
	}
	result := make([]types.MonitorDashboardTaskResponse, 0, len(relations))
	for _, r := range relations {
		resp := types.MonitorDashboardTaskResponse{MonitorDashboardTask: r}
		if t, ok := taskMap[r.TaskId]; ok {
			resp.TaskName = t.TaskName
		}
		result = append(result, resp)
	}
	return result, nil
}

// TaskSort 批量改排序并同步 Grafana 重排
func (app *MonitorDashboardApplication) TaskSort(ctx context.Context, req *types.MonitorDashboardTaskSortRequest) error {
	if req == nil || len(req.Items) == 0 {
		return errors.New("排序项不能为空")
	}
	// 用第一条关联反查面板（排序请求为单面板，前端保证 items 同属一面板）
	first, err := app.repository.MonitorDashboardTaskRepository.GetById(req.Items[0].Id)
	if err != nil {
		return err
	}
	if first == nil {
		return errors.New("排序项不存在")
	}

	sortItems := make([]entity.MonitorDashboardTask, 0, len(req.Items))
	for _, item := range req.Items {
		sortItems = append(sortItems, entity.MonitorDashboardTask{
			BaseEntity: common.BaseEntity{Id: item.Id},
			Sort:       item.Sort,
		})
	}
	if err := app.repository.MonitorDashboardTaskRepository.BatchModifySort(sortItems); err != nil {
		return err
	}

	// 同步 Grafana 重排：库已更新，重排失败返回 error 便于前端提示重试（幂等可重复提交）
	if app.grafanaHandler != nil {
		if err := app.syncDashboardSort(ctx, first.DashboardId); err != nil {
			return fmt.Errorf("同步 Grafana 面板排序失败: %w", err)
		}
	}
	return nil
}

// syncDashboardSort 按面板当前 sort 顺序取任务 key，让 Grafana 按序重排 panel。
// FindByDashboardId 已按 sort desc 返回，顺序即期望展示顺序。
func (app *MonitorDashboardApplication) syncDashboardSort(ctx context.Context, dashboardId int64) error {
	related, err := app.repository.MonitorDashboardTaskRepository.FindByDashboardId(dashboardId)
	if err != nil {
		return err
	}
	if len(related) == 0 {
		return nil
	}
	taskIds := make([]int64, 0, len(related))
	for _, r := range related {
		taskIds = append(taskIds, r.TaskId)
	}
	tasks, err := app.repository.MonitorTaskRepository.SelectByIds(taskIds)
	if err != nil {
		return err
	}
	taskKeyMap := make(map[int64]string, len(tasks))
	for _, t := range tasks {
		taskKeyMap[t.Id] = t.TaskKey
	}
	taskKeys := make([]string, 0, len(related))
	for _, r := range related {
		if k, ok := taskKeyMap[r.TaskId]; ok {
			taskKeys = append(taskKeys, k)
		}
	}
	d, err := app.repository.MonitorDashboardRepository.GetById(dashboardId)
	if err != nil {
		return err
	}
	if d == nil || d.Uid == "" {
		return nil
	}
	return app.grafanaHandler.ReSortDashboard(d.Uid, taskKeys)
}

// Count 统计
func (app *MonitorDashboardApplication) Count(ctx context.Context) (int64, error) {
	c, err := app.repository.MonitorDashboardRepository.Count()
	if err != nil || c == nil {
		return 0, err
	}
	return *c, nil
}

func (app *MonitorDashboardApplication) prefixUrl(url string) string {
	if url == "" || app.grafanaConf == nil || app.grafanaConf.Addr == "" {
		return url
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	addr := strings.TrimRight(app.grafanaConf.Addr, "/")
	if !strings.HasPrefix(url, "/") {
		return fmt.Sprintf("%s/%s", addr, url)
	}
	return addr + url
}
