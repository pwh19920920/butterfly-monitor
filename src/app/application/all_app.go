package application

import (
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/infrastructure/persistence"
)

type Application struct {
	JobExec     JobExecApplication
	JobDatabase JobDatabaseApplication
}

func NewApplication(
	config config.Config,
	repository *persistence.Repository,
) *Application {
	return &Application{
		// 定时执行器
		JobExec: NewJobExecApplication(
			config.Sequence,
			repository,
			config.XxlJobExec,
			config.InfluxDbOption,
		),

		// 任务数据库
		JobDatabase: JobDatabaseApplication{
			sequence:   config.Sequence,
			repository: repository,
		},
	}
}

func (app *Application) RegisterJobExec() {
	app.JobExec.RegisterExecJob()
}
