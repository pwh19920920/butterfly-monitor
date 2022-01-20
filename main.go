package main

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/interfaces"
	"butterfly-monitor/job"
	"github.com/pwh19920920/butterfly"
)
import "github.com/pwh19920920/butterfly-admin/starter"

func init() {
	adminConfig, adminApp := starter.InitButterflyAdmin()
	allConfig := config.InitAll(adminConfig)
	repository := persistence.NewRepository(allConfig)
	app := application.NewApplication(adminApp, allConfig, repository)
	timerJob := job.NewJob(allConfig, repository, app)

	// 初始化路由
	interfaces.InitMonitorTaskHandler(app, timerJob)
	interfaces.InitMonitorDatabaseHandler(app)
	interfaces.InitMonitorTestHandler(app)
	interfaces.InitMonitorDashboardHandler(app)
	interfaces.InitMonitorHealthHandler(app)
	interfaces.InitAlertConfHandler(app)
	interfaces.InitAlertGroupHandler(app)
	interfaces.InitAlertChannelHandler(app)
	interfaces.InitMonitorTaskEventHandler(app)

	// 注册定时任务
	timerJob.RegisterJobExec()
}

func main() {
	butterfly.Run()
}
