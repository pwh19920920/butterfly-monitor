package application

import (
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/infrastructure/persistence"
	"butterfly-monitor/src/app/infrastructure/support"
)

type Application struct {
	MonitorExec      MonitorExecApplication
	MonitorDatabase  MonitorDatabaseApplication
	MonitorDashboard MonitorDashboardApplication
	MonitorTask      MonitorTaskApplication
}

func NewApplication(
	config config.Config,
	repository *persistence.Repository,
) *Application {
	return &Application{
		// 定时执行器
		MonitorExec: NewMonitorExecApplication(
			config.Sequence,
			repository,
			config.XxlJobExec,
			config.InfluxDbOption,
			config.Grafana,
		),

		// 监控数据库
		MonitorDatabase: MonitorDatabaseApplication{
			sequence:   config.Sequence,
			repository: repository,
		},

		// 监控任务
		MonitorTask: MonitorTaskApplication{
			sequence:       config.Sequence,
			repository:     repository,
			grafanaHandler: support.NewGrafanaOptionHandler(config.Grafana),
		},

		// 主板配置
		MonitorDashboard: MonitorDashboardApplication{
			sequence:       config.Sequence,
			repository:     repository,
			Grafana:        config.Grafana,
			grafanaHandler: support.NewGrafanaOptionHandler(config.Grafana),
		},
	}
}

func (app *Application) RegisterJobExec() {
	app.MonitorExec.RegisterExecJob()
}
