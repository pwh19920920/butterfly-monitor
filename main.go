package main

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/interfaces"
	"fmt"
	"github.com/pwh19920920/butterfly"
	"strings"
	"time"
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
	effectTime := "20:10:10-22:10:10"
	idx := strings.LastIndex(effectTime, "-")
	startTimeStr := effectTime[0:idx]
	endTimeStr := effectTime[idx+1 : len(effectTime)]

	// 转换开始时间
	currentTime := time.Now()
	dateStr := currentTime.Format("2006-01-02")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", dateStr, startTimeStr), time.Local)
	fmt.Println(err)

	// 转换结束时间
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", dateStr, endTimeStr), time.Local)
	fmt.Println(err)

	if currentTime.Unix() < startTime.Unix() || currentTime.Unix() > endTime.Unix() {
		fmt.Println("不在时间范围内")
	}

	println(startTime.Unix())
	println(currentTime.Unix())
	println(endTime.Unix())
	butterfly.Run()
}
