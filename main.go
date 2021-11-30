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

func main() {

	//xx := handler.ChannelEmailHandler{}
	//channel := entity.AlertChannel{
	//	Params: "{\"host\":\"smtp.exmail.qq.com\",\"port\":\"465\",\"username\":\"ibg-fund@we.cn\",\"password\":\"3dkrkpzPecq59kZL\",\"ssl\":true}",
	//}
	//
	//groupUsers := make([]entity2.SysUser, 0)
	//groupUsers = append(groupUsers, entity2.SysUser{
	//	Email: "pengweihuang@we.cn",
	//})
	//err := xx.DispatchMessage(channel, groupUsers, "hello world")
	//println(err)

	//xx := handler.ChannelDingDingHandler{}
	//channel := entity.AlertChannel{
	//	Params: "{\"addr\":\"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=d1d5354d-be44-4539-b6b0-d7534bde1e33\"}",
	//}
	//
	//groupUsers := make([]entity2.SysUser, 0)
	//err := xx.DispatchMessage(channel, groupUsers, "hello world")
	//fmt.Printf("%v", err)

	butterfly.Run()
}
