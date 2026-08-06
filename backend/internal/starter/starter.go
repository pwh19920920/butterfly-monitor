package starter

import (
	"context"
	"dragonfly-monitor/internal/common"
	"time"

	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/config"
	"dragonfly-monitor/internal/config/xxljob"
	"dragonfly-monitor/internal/domain/entity"
	infraHandler "dragonfly-monitor/internal/infrastructure/handler"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/infrastructure/security"
	"dragonfly-monitor/internal/infrastructure/system"
	"dragonfly-monitor/internal/interfaces"
	"dragonfly-monitor/internal/interfaces/middleware"
	"dragonfly-monitor/internal/job"

	"github.com/gin-gonic/gin"
	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/butterfly/pkg/response"
	"github.com/pwh19920920/butterfly/pkg/server"
)

func route401(context *gin.Context) {
	response.Response(context, 401, "请登录后在进行此操作", nil)
}

func route403(context *gin.Context) {
	response.Response(context, 403, "您没有权限进行此操作", nil)
}

func InitButterflyAdmin() (config.Config, *application.Application) {
	// 1. 现有 repository / app / routes
	allConfig := config.InitAll()
	repository := persistence.NewRepository(allConfig)
	encodeService := security.NewEncodeServiceImpl()
	tokenService := security.NewJwtServiceImpl()
	// 系统指标采集器：后台 goroutine 低频(30s)采样，供首页性能卡片
	sysCollector := system.NewCollector()
	app := application.NewApplication(
		allConfig,
		repository,
		encodeService,
		tokenService,
		sysCollector,
	)

	// XXL-JOB 执行器（不在 InitAll 里初始化）：需在 NewJob 前注入到 cfg
	xxlExec := xxljob.GetXxlJobExec()
	allConfig.XxlJobExec = xxlExec

	// XXL-JOB 聚合：需在路由注册前创建，供 dataPush handler 引用 DataCollect
	timerJob := job.NewJob(allConfig, repository, app)

	// 系统路由
	interfaces.InitLoginHandler(app)
	interfaces.InitSysMenuHandler(app)
	interfaces.InitSysRoleHandler(app)
	interfaces.InitSysUserHandler(app)

	// 监控管理路由
	interfaces.InitMonitorDatabaseHandler(app)
	interfaces.InitMonitorTaskHandler(app, timerJob)
	interfaces.InitMonitorDashboardHandler(app)
	interfaces.InitMonitorGroupHandler(app)
	interfaces.InitAlertConfHandler(app)
	interfaces.InitAlertGroupHandler(app)
	interfaces.InitAlertChannelHandler(app)
	interfaces.InitMonitorTaskEventHandler(app)
	interfaces.InitMonitorVolatilityDayHandler(app)
	interfaces.InitMonitorHomeHandler(app)
	interfaces.InitMonitorHealthHandler(app)
	interfaces.InitMonitorSystemHandler(app)

	// 2. 注册 CommonMap handlers
	registerHandlers(app, repository, allConfig)

	// 4. 可选：每分钟扫数据源建连
	go refreshDatabaseConnections(app, repository)

	// 5. 可选：每 30s 采系统指标，供首页性能卡片（后台 goroutine，不影响请求）
	go sysCollector.Run()

	// 5. 注册四个 XXL Job
	timerJob.RegisterJobExec()

	// 注册中间件
	server.RegisterMiddleware(middleware.JwtAuth(
		app,
		route401,
		route403,
	))
	return allConfig, app
}

// registerHandlers 装配 Database / Command / Channel handler 到 CommonMap
func registerHandlers(app *application.Application, repository *persistence.Repository, allConfig config.Config) {
	if app.CommonMap == nil {
		return
	}
	ctx := context.Background()
	// DatabaseMysqlHandler：type 2
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeMysql), &infraHandler.DatabaseMysqlHandler{})

	// DatabaseMongoHandler：type 1
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeMongo), &infraHandler.DatabaseMongoHandler{})

	// DatabasePostgresHandler：type 3
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypePostgres), &infraHandler.DatabasePostgresHandler{})

	// DatabaseClickHouseHandler：type 9（独立方言）
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeClickHouse), &infraHandler.DatabaseClickHouseHandler{})

	// DatabasePrometheusHandler：type 10（只读 PromQL，无写）；type 13 VictoriaMetrics 复用（默认端口 8428）
	promHandler := &infraHandler.DatabasePrometheusHandler{}
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypePrometheus), promHandler)
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeVictoriaMetrics), &infraHandler.DatabasePrometheusHandler{DefaultPort: "8428"})

	// OpenSearch / Elasticsearch：只读 _search/_count/SQL，无写；共用同一 handler
	osHandler := &infraHandler.DatabaseOpenSearchHandler{}
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeOpenSearch), osHandler)    // 11
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeElasticsearch), osHandler) // 12

	// DatabaseTdEngineHandler：type 14（只读 REST SQL，无写）
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeTDengine), &infraHandler.DatabaseTdEngineHandler{})

	// MySQL 协议兼容族：共用 DatabaseMysqlHandler
	mysqlCompat := &infraHandler.DatabaseMysqlHandler{}
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeMariaDB), mysqlCompat)   // 4
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeTiDB), mysqlCompat)      // 5
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeOceanBase), mysqlCompat) // 6 MySQL 模式
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeDoris), mysqlCompat)     // 7
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeStarRocks), mysqlCompat) // 8

	// CommandUrlHandler：TaskTypeURL
	app.CommonMap.RegisterCommandHandler(ctx, int32(entity.TaskTypeURL), &infraHandler.CommandUrlHandler{})

	// CommandDataBaseHandler：TaskTypeDatabase，注入 commonMap 连接
	app.CommonMap.RegisterCommandHandler(ctx, int32(entity.TaskTypeDatabase), &infraHandler.CommandDataBaseHandler{
		GetConn: func(id int64) (interface{}, bool) {
			return app.CommonMap.GetDatabaseConn(context.Background(), id)
		},
		GetHandler: func(dsType int32) (infraHandler.DatabaseHandlerAdapter, bool) {
			h, ok := app.CommonMap.GetDatabaseHandler(context.Background(), dsType)
			if !ok {
				return nil, false
			}
			return h, true
		},

		GetDatabaseType: func(id int64) (int32, error) {
			db, err := repository.MonitorDatabaseRepository.GetById(id)
			if err != nil || db == nil {
				return 0, err
			}

			return int32(db.Type), nil
		},
	})

	// DatabaseTimeseriesHandler：TaskTypeDrilldown（=4，特殊值值）。
	// 下钻任务以它为 CommandHandler，executeCollect 复用 cmd.ExecuteCommand → buildCollectPoints 主流程，
	// 无需单独分支。它从依赖的聚合任务 VM 中按标签过滤取 float64，其余采样/告警与正常任务完全一致。
	app.CommonMap.RegisterCommandHandler(ctx, int32(entity.TaskTypeDrilldown), &infraHandler.DatabaseTimeseriesHandler{
		TimeSeries: allConfig.TimeSeries,
		GetSourceTask: func(id int64) (*entity.MonitorTask, error) {
			return repository.MonitorTaskRepository.GetById(id)
		},
	})

	// 通道
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelEmailHandler{})
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelWechatHandler{})
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelDingtalkHandler{})
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelFeishuHandler{})
}

// databaseHealthFailThreshold 连续探活失败达到此次数才标异常并摘连接（防网络抖动误杀）。
// 周期为 1 分钟时，默认 3 次 ≈ 连续 3 分钟失败才摘。
const databaseHealthFailThreshold int32 = 3

// refreshDatabaseConnections 每分钟扫数据源：建连 + 周期验活。
// - 无连接：NewInstance 建连，成功放 map 并记健康，失败累加 consecutive_fail
// - 有连接：TestConnect 轻量探活；成功归零；失败只计数，连续达阈值才摘连接并标异常
// 循环骨架（双层 recover + 周期执行）由 common.RunSafeLoop 提供
func refreshDatabaseConnections(app *application.Application, repository *persistence.Repository) {
	common.RunSafeLoop("refreshDatabaseConnections", time.Minute, func() {
		ctx := context.Background()
		list, err := repository.MonitorDatabaseRepository.SelectAll(nil)
		if err != nil {
			logger.Warn(ctx, "scan monitor database fail", err)
			return
		}
		for _, item := range list {
			refreshOneDatabase(ctx, app, repository, item)
		}
	})
}

// refreshOneDatabase 对单个数据源做建连/探活，并回写健康字段
func refreshOneDatabase(ctx context.Context, app *application.Application, repository *persistence.Repository, item entity.MonitorDatabase) {
	h, ok := app.CommonMap.GetDatabaseHandler(ctx, int32(item.Type))
	if !ok {
		return
	}

	now := &common.LocalTime{Time: time.Now()}
	_, hasConn := app.CommonMap.GetDatabaseConn(ctx, item.Id)

	if !hasConn {
		// 无连接：尝试建连（NewInstance 内部会 GetDecodePassword，勿提前改写 Password）
		conn, err := h.NewInstance(item)
		if err != nil {
			failCount := item.ConsecutiveFail + 1
			logger.WarnFormat(ctx, "connect database %d fail count=%d: %v", item.Id, failCount, err)
			// 建连失败本身已无连接可摘；达阈值才标红，未达则保持未知（避免抖动刷异常）
			status := entity.DatabaseHealthUnknown
			if failCount >= databaseHealthFailThreshold {
				status = entity.DatabaseHealthBad
			}
			markDatabaseHealth(ctx, repository, item, status, now, err.Error(), failCount)
			return
		}
		app.CommonMap.PutDatabaseConn(ctx, item.Id, conn)
		logger.InfoFormat(ctx, "database conn ready id=%d name=%s", item.Id, item.Name)
		markDatabaseHealth(ctx, repository, item, entity.DatabaseHealthOK, now, " ", 0)
		return
	}

	// 有连接：轻量探活（TestConnect 独立建短连/Ping，不污染业务连接池）
	if err := h.TestConnect(item); err != nil {
		failCount := item.ConsecutiveFail + 1
		logger.WarnFormat(ctx, "probe database %d fail count=%d: %v", item.Id, failCount, err)
		if failCount >= databaseHealthFailThreshold {
			// 连续失败达阈值：摘掉死连接，下一轮走建连路径重建
			app.CommonMap.RemoveDatabaseConn(ctx, item.Id, int32(item.Type))
			markDatabaseHealth(ctx, repository, item, entity.DatabaseHealthBad, now, err.Error(), failCount)
			return
		}
		// 未达阈值：保留连接，只累加失败计数与错误信息，状态仍视为正常（防抖）
		markDatabaseHealth(ctx, repository, item, entity.DatabaseHealthOK, now, err.Error(), failCount)
		return
	}
	markDatabaseHealth(ctx, repository, item, entity.DatabaseHealthOK, now, " ", 0)
}

// markDatabaseHealth 回写探活结果；写库失败只打日志，不影响下一轮
func markDatabaseHealth(ctx context.Context, repository *persistence.Repository, item entity.MonitorDatabase, status int32, checkTime *common.LocalTime, lastError string, consecutiveFail int32) {
	if lastError == "" {
		lastError = " "
	}
	const maxErrLen = 900
	if len(lastError) > maxErrLen {
		lastError = lastError[:maxErrLen] + "..."
	}
	if err := repository.MonitorDatabaseRepository.UpdateHealth(item.Id, status, checkTime, lastError, consecutiveFail); err != nil {
		logger.WarnFormat(ctx, "update database health id=%d fail: %v", item.Id, err)
	}
}
