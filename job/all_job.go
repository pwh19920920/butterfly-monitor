package job

import (
	"butterfly-monitor/application"
	"butterfly-monitor/config"
	"butterfly-monitor/infrastructure/persistence"
)

type Job struct {
	// 定时任务
	MonitorDataCollectJob MonitorDataCollectJob
	MonitorAlertCheckJob  MonitorAlertCheckJob
	MonitorEventCheckJob  MonitorEventCheckJob
}

func NewJob(config config.Config,
	repository *persistence.Repository,
	application *application.Application,
) *Job {
	return &Job{
		// 定时执行器
		MonitorDataCollectJob: MonitorDataCollectJob{
			sequence:       config.Sequence,
			repository:     repository,
			xxlExec:        config.XxlJobExec,
			influxDbOption: config.InfluxDbOption,
			grafana:        config.Grafana,
			commonMap:      application.CommonMapApp,
		},

		// 规则检查
		MonitorAlertCheckJob: MonitorAlertCheckJob{
			sequence:   config.Sequence,
			repository: repository,
			influxdb:   config.InfluxDbOption,
			xxlExec:    config.XxlJobExec,
			grafana:    config.Grafana,
			alertConf:  application.AlertConfApp,
		},

		// 事件检查
		MonitorEventCheckJob: MonitorEventCheckJob{
			sequence:     config.Sequence,
			repository:   repository,
			xxlExec:      config.XxlJobExec,
			alertConf:    application.AlertConfApp,
			alertChannel: application.AlertChannelApp,
			commonMap:    application.CommonMapApp,
		},
	}
}

func (job *Job) RegisterJobExec() {
	job.MonitorDataCollectJob.RegisterExecJob()
	job.MonitorAlertCheckJob.RegisterExecJob()
	job.MonitorEventCheckJob.RegisterExecJob()
}
