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
	interfaces.InitAlertGroupHandler(app)
	interfaces.InitAlertChannelHandler(app)

	// 注册定时任务
	app.RegisterJobExec()
}

type AlertConfObject struct {
	ScanSpan   int64  `json:"scanSpan"`   // 扫描间隔
	AlertSpan  int64  `json:"alertSpan"`  // 报警间隔
	FirstDelay int64  `json:"firstDelay"` // 首次延迟
	Template   string `json:"template"`   // 报警模板
}

type AlertConfObjectInstance struct {
	Alert AlertConfObject `json:"alert"`
}

func main() {
	butterfly.Run()
}
