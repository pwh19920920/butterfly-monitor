package application

import (
	"butterfly-monitor/internal/config"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/domain/security"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/infrastructure/support"
	"butterfly-monitor/internal/infrastructure/system"
	"butterfly-monitor/internal/types"
	"context"

	"golang.org/x/sync/errgroup"
)

// Application 应用聚合：对外只暴露各业务 App，装配细节收敛在 NewApplication
type Application struct {
	Login                LoginApplication
	SysMenu              SysMenuApplication
	SysUser              SysUserApplication
	SysRole              SysRoleApplication
	SysPermission        SysPermissionApplication
	CommonMap            *CommonMapApplication
	MonitorDatabase      MonitorDatabaseApplication
	MonitorTask          MonitorTaskApplication
	MonitorDashboard     MonitorDashboardApplication
	MonitorGroup         MonitorGroupApplication
	MonitorTaskEvent     MonitorTaskEventApplication
	MonitorVolatilityDay MonitorVolatilityDayApplication
	AlertConf            AlertConfApplication
	AlertGroup           AlertGroupApplication
	AlertChannel         AlertChannelApplication
	System               SystemApplication
}

// NewApplication 装配全部应用服务
func NewApplication(
	cfg config.Config,
	repository *persistence.Repository,
	encoderService security.EncodeService,
	tokenService security.TokenService,
	systemCollector *system.Collector,
) *Application {
	// 共享依赖
	seq := cfg.Sequence
	commonMap := NewCommonMapApplication(repository)
	grafanaHandler := support.NewGrafanaHandler(cfg.Grafana, cfg.MetricQuery)
	alertConf := NewAlertConfApplication(seq, repository)

	return &Application{
		// 系统
		Login:         NewLoginApplication(seq, repository, encoderService, tokenService, cfg.AuthConfig),
		SysMenu:       NewSysMenuApplication(seq, repository),
		SysUser:       NewSysUserApplication(seq, repository, encoderService, tokenService, cfg.AuthConfig),
		SysRole:       NewSysRoleApplication(seq, repository),
		SysPermission: NewSysPermissionApplication(seq, repository),

		// 运行期映射
		CommonMap: commonMap,

		// 监控
		MonitorDatabase:      NewMonitorDatabaseApplication(seq, repository, commonMap),
		MonitorTask:          NewMonitorTaskApplication(seq, repository, grafanaHandler, commonMap),
		MonitorDashboard:     NewMonitorDashboardApplication(seq, repository, grafanaHandler, cfg.Grafana),
		MonitorGroup:         NewMonitorGroupApplication(seq, repository),
		MonitorTaskEvent:     NewMonitorTaskEventApplication(seq, repository, &alertConf),
		MonitorVolatilityDay: NewMonitorVolatilityDayApplication(seq, repository),

		// 告警
		AlertConf:    alertConf,
		AlertGroup:   NewAlertGroupApplication(seq, repository),
		AlertChannel: NewAlertChannelApplication(seq, repository, commonMap, alertConf),

		// 系统指标
		System: NewSystemApplication(systemCollector),
	}
}

// HomeCount 首页统计聚合：并发查询各业务计数，降低串行 RT
func (app *Application) HomeCount(ctx context.Context) (*types.MonitorHomeCountResponse, error) {
	var (
		taskCount       int64
		eventCount      int64
		dashboardCount  int64
		databaseCount   int64
		alertChannelCnt int64
		alertGroupCount int64
		statusCounts    map[entity.MonitorTaskEventDealStatus]int64
		alertLevelDist  map[int32]int64
		recentEvents    []entity.MonitorTaskEvent
	)

	eg, ctxt := errgroup.WithContext(ctx)
	eg.Go(func() (err error) { taskCount, err = app.MonitorTask.Count(ctxt); return })
	eg.Go(func() (err error) { eventCount, err = app.MonitorTaskEvent.Count(ctxt); return })
	eg.Go(func() (err error) { dashboardCount, err = app.MonitorDashboard.Count(ctxt); return })
	eg.Go(func() (err error) { databaseCount, err = app.MonitorDatabase.Count(ctxt); return })
	eg.Go(func() (err error) { alertChannelCnt, err = app.AlertChannel.Count(ctxt); return })
	eg.Go(func() (err error) { alertGroupCount, err = app.AlertGroup.Count(ctxt); return })
	eg.Go(func() (err error) { statusCounts, err = app.MonitorTaskEvent.CountByStatus(ctxt); return })
	eg.Go(func() (err error) { alertLevelDist, err = app.MonitorTaskEvent.CountByLevel(ctxt); return })
	eg.Go(func() (err error) { recentEvents, err = app.MonitorTaskEvent.SelectRecent(5); return })
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	taskIds := make([]int64, 0)
	taskIdSet := make(map[int64]bool)
	for _, ev := range recentEvents {
		if !taskIdSet[ev.TaskId] {
			taskIdSet[ev.TaskId] = true
			taskIds = append(taskIds, ev.TaskId)
		}
	}
	taskMap, _ := app.MonitorTask.repository.MonitorTaskRepository.SelectByIdsWithMap(taskIds)

	recentList := make([]types.MonitorTaskEventListItem, 0, len(recentEvents))
	for _, ev := range recentEvents {
		msg := ev.AlertMsg
		if len(msg) > 60 {
			msg = msg[:60]
		}
		ct := ""
		if ev.CreatedAt != nil {
			ct = ev.CreatedAt.Format("2006-01-02 15:04:05")
		}
		tName := ""
		if t, ok := taskMap[ev.TaskId]; ok {
			tName = t.TaskName
		}
		recentList = append(recentList, types.MonitorTaskEventListItem{
			Id:         ev.Id,
			TaskName:   tName,
			AlertMsg:   msg,
			DealStatus: ev.DealStatus,
			EventLevel: ev.EventLevel,
			CreateTime: ct,
		})
	}

	return &types.MonitorHomeCountResponse{
		TaskCount:              taskCount,
		EventCount:             eventCount,
		DashboardCount:         dashboardCount,
		DatabaseCount:          databaseCount,
		AlertChannelCount:      alertChannelCnt,
		AlertGroupCount:        alertGroupCount,
		PendingEvents:          statusCounts[entity.MonitorTaskEventDealStatusPending],
		ProcessingEvents:       statusCounts[entity.MonitorTaskEventDealStatusProcessing],
		CompleteEvents:         statusCounts[entity.MonitorTaskEventDealStatusComplete],
		IgnoreEvents:           statusCounts[entity.MonitorTaskEventDealStatusIgnore],
		AlertLevelDistribution: alertLevelDist,
		RecentEvents:           recentList,
	}, nil
}
