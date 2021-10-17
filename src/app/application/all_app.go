package application

import (
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/infrastructure/persistence"
)

type Application struct {
	MonitorExec     MonitorExecApplication
	MonitorDatabase MonitorDatabaseApplication
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
		),

		// 任务数据库
		MonitorDatabase: MonitorDatabaseApplication{
			sequence:   config.Sequence,
			repository: repository,
		},
	}
}

func (app *Application) RegisterJobExec() {
	app.MonitorExec.RegisterExecJob()
}
