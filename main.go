package main

import (
	"butterfly-monitor/src/app/application"
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/interfaces"
	"github.com/pwh19920920/butterfly"
)
import "github.com/pwh19920920/butterfly-admin/src/app/starter"

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

	// 注册定时任务
	app.RegisterJobExec()
}

func main() {
	butterfly.Run()
}
