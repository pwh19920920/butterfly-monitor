package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/config/grafana"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/infrastructure/support"
	"dragonfly-monitor/internal/types"

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

	// 同步 Grafana：取第一个关联的 dashboard 重排
	if app.grafanaHandler != nil {
		// 尝试用 id 反查 dashboard 关联
		firstId := req.Items[0].Id
		// 从全部关联里找
		all, err := app.repository.MonitorDashboardTaskRepository.SelectAll()
		if err == nil {
			var dashboardId int64
			taskIds := make([]int64, 0)
			for _, item := range all {
				if item.Id == firstId {
					dashboardId = item.DashboardId
					break
				}
			}
			if dashboardId != 0 {
				related, _ := app.repository.MonitorDashboardTaskRepository.FindByDashboardId(dashboardId)
				for _, r := range related {
					taskIds = append(taskIds, r.TaskId)
				}
				// 按 sort 已更新后的顺序
				tasks, _ := app.repository.MonitorTaskRepository.SelectByIds(taskIds)
				taskKeyMap := make(map[int64]string)
				for _, t := range tasks {
					taskKeyMap[t.Id] = t.TaskKey
				}
				taskKeys := make([]string, 0)
				for _, r := range related {
					if k, ok := taskKeyMap[r.TaskId]; ok {
						taskKeys = append(taskKeys, k)
					}
				}
				d, _ := app.repository.MonitorDashboardRepository.GetById(dashboardId)
				if d != nil {
					_ = app.grafanaHandler.ReSortDashboard(d.Uid, taskKeys)
				}
			}
		}
	}
	return nil
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
