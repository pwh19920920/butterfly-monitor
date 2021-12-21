package application

import (
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
	"butterfly-monitor/infrastructure/support"
	adminApp "github.com/pwh19920920/butterfly-admin/application"
)

type Application struct {
	MonitorDatabaseApp  MonitorDatabaseApplication
	MonitorDashboardApp MonitorDashboardApplication
	MonitorTaskApp      MonitorTaskApplication
	AlertGroupApp       AlertGroupApplication
	AlertConfApp        AlertConfApplication
	AlertChannelApp     AlertChannelApplication
	MonitorTaskEventApp MonitorTaskEventApplication
	CommonMapApp        CommonMapApplication
	AllConfig           config.Config
	AdminApp            *adminApp.Application
}

func NewApplication(app *adminApp.Application, config config.Config, repository *persistence.Repository) *Application {
	commonMapApp := NewCommonMapApplication(repository)
	return &Application{
		// adminApp
		AdminApp: app,

		// 配置表
		CommonMapApp: commonMapApp,

		// 全局配置
		AllConfig: config,

		// 监控数据库
		MonitorDatabaseApp: MonitorDatabaseApplication{
			sequence:   config.Sequence,
			repository: repository,
			commonMap:  commonMapApp,
		},

		// 监控任务
		MonitorTaskApp: MonitorTaskApplication{
			sequence:       config.Sequence,
			repository:     repository,
			grafanaHandler: support.NewGrafanaOptionHandler(config.Grafana),
		},

		// 主板配置
		MonitorDashboardApp: MonitorDashboardApplication{
			sequence:       config.Sequence,
			repository:     repository,
			Grafana:        config.Grafana,
			grafanaHandler: support.NewGrafanaOptionHandler(config.Grafana),
		},

		// 报警配置
		AlertConfApp: AlertConfApplication{
			repository: repository,
			sequence:   config.Sequence,
		},

		// 分组
		AlertGroupApp: AlertGroupApplication{
			repository: repository,
			sequence:   config.Sequence,
		},

		// 通道
		AlertChannelApp: AlertChannelApplication{
			repository: repository,
			sequence:   config.Sequence,
			commonMap:  commonMapApp,
		},

		// 事件
		MonitorTaskEventApp: MonitorTaskEventApplication{
			repository: repository,
			sequence:   config.Sequence,
		},
	}
}
