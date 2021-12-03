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
	MonitorAlert     MonitorAlertCheckApplication
	MonitorEvent     MonitorEventCheckApplication
}

func NewApplication(
	config config.Config,
	repository *persistence.Repository,
) *Application {
	alertConfApp := AlertConfApplication{repository: repository, sequence: config.Sequence}
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
		AlertConf: alertConfApp,

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

		// 监控报警
		MonitorAlert: MonitorAlertCheckApplication{
			sequence:   config.Sequence,
			repository: repository,
			influxdb:   config.InfluxDbOption,
			xxlExec:    config.XxlJobExec,
			grafana:    config.Grafana,
			alertConf:  alertConfApp,
		},

		// 事件处理
		MonitorEvent: MonitorEventCheckApplication{
			sequence:   config.Sequence,
			repository: repository,
			xxlExec:    config.XxlJobExec,
		},
	}
}

func (app *Application) RegisterJobExec() {
	app.MonitorExec.RegisterExecJob()
	app.MonitorAlert.RegisterExecJob()
}
