package main

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/interfaces"
	"github.com/pwh19920920/butterfly"
)
import "github.com/pwh19920920/butterfly-admin/starter"

func init() {
	adminConfig := starter.InitButterflyAdmin()
	allConfig := config.InitAll(adminConfig)
	repository := persistence.NewRepository(allConfig)
	app := application.NewApplication(
		allConfig,
		repository,
	)

	// 初始化路由
	interfaces.InitMonitorDatabaseHandler(app)
	interfaces.InitMonitorTaskHandler(app)
	interfaces.InitMonitorTestHandler(app)
	interfaces.InitMonitorDashboardHandler(app)
	interfaces.InitMonitorHealthHandler(app)
	interfaces.InitAlertConfHandler(app)

	// 注册定时任务
	app.RegisterJobExec()
}

func main() {
	butterfly.Run()
}
