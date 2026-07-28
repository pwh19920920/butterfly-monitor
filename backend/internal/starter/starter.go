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
	interfaces.InitMonitorHomeHandler(app)
	interfaces.InitMonitorHealthHandler(app)
	interfaces.InitMonitorSystemHandler(app)

	// 2. 注册 CommonMap handlers
	registerHandlers(app, repository)

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
func registerHandlers(app *application.Application, repository *persistence.Repository) {
	if app.CommonMap == nil {
		return
	}
	ctx := context.Background()
	// DatabaseMysqlHandler：type 2
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeMysql), &infraHandler.DatabaseMysqlHandler{})

	// DatabaseMongoHandler：type 1
	app.CommonMap.RegisterDatabaseHandler(ctx, int32(entity.DataSourceTypeMongo), &infraHandler.DatabaseMongoHandler{})

	// CommandUrlHandler：TaskTypeURL
	app.CommonMap.RegisterCommandHandler(ctx, int32(entity.TaskTypeURL), &infraHandler.CommandUrlHandler{})

	// CommandDataBaseHandler：TaskTypeDatabase，注入 commonMap 连接
	app.CommonMap.RegisterCommandHandler(ctx, int32(entity.TaskTypeDatabase), &infraHandler.CommandDataBaseHandler{
		GetConn: func(id int64) (interface{}, bool) {
			return app.CommonMap.GetDatabaseConn(context.Background(), id)
		},
		GetHandler: func(dsType int32) (func(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error), bool) {
			h, ok := app.CommonMap.GetDatabaseHandler(context.Background(), dsType)
			if !ok {
				return nil, false
			}
			return h.ExecuteQuery, true
		},

		GetDatabaseType: func(id int64) (int32, error) {
			db, err := repository.MonitorDatabaseRepository.GetById(id)
			if err != nil || db == nil {
				return 0, err
			}

			return int32(db.Type), nil
		},
	})

	// 通道
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelEmailHandler{})
	app.CommonMap.RegisterChannelHandler(ctx, &infraHandler.ChannelWechatHandler{})
}

// refreshDatabaseConnections 每分钟扫数据源建连
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
			if _, ok := app.CommonMap.GetDatabaseConn(ctx, item.Id); ok {
				continue
			}
			h, ok := app.CommonMap.GetDatabaseHandler(ctx, int32(item.Type))
			if !ok {
				continue
			}
			// NewInstance 内部会 GetDecodePassword，勿提前改写 Password
			conn, err := h.NewInstance(item)
			if err != nil {
				logger.WarnFormat(ctx, "connect database %d fail: %v", item.Id, err)
				continue
			}
			app.CommonMap.PutDatabaseConn(ctx, item.Id, conn)
			logger.InfoFormat(ctx, "database conn ready id=%d name=%s", item.Id, item.Name)
		}
	})
}
