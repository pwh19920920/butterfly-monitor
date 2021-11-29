package application

import (
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/infrastructure/support"
)

type Application struct {
	MonitorExec      MonitorExecApplication
	MonitorDatabase  MonitorDatabaseApplication
	MonitorDashboard MonitorDashboardApplication
	MonitorTask      MonitorTaskApplication
	AlertConf        AlertConfApplication
	AlertGroup       AlertGroupApplication
	AlertChannel     AlertChannelApplication
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

		// 报警配置
		AlertConf: AlertConfApplication{
			repository: repository,
			sequence:   config.Sequence,
		},

		// 分组
		AlertGroup: AlertGroupApplication{
			repository: repository,
			sequence:   config.Sequence,
		},

		// 通道
		AlertChannel: AlertChannelApplication{
			repository: repository,
			sequence:   config.Sequence,
		},
	}
}

func (app *Application) RegisterJobExec() {
	app.MonitorExec.RegisterExecJob()
}
