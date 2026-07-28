package job

import (
	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/config"
	"dragonfly-monitor/internal/infrastructure/persistence"
)

// Job 定时任务聚合
type Job struct {
	DataCollect  MonitorDataCollectJob
	DataSampling MonitorDataSamplingJob
	AlertCheck   MonitorAlertCheckJob
	EventCheck   MonitorEventCheckJob
}

func NewJob(cfg config.Config, repository *persistence.Repository, app *application.Application) *Job {
	return &Job{
		DataCollect: MonitorDataCollectJob{
			sequence:    cfg.Sequence,
			repository:  repository,
			xxlExec:     cfg.XxlJobExec,
			timeSeries:  cfg.TimeSeries,
			metricQuery: cfg.MetricQuery,
			commonMap:   app.CommonMap,
			alertConf:   app.AlertConf,
		},
		DataSampling: MonitorDataSamplingJob{
			sequence:    cfg.Sequence,
			repository:  repository,
			xxlExec:     cfg.XxlJobExec,
			timeSeries:  cfg.TimeSeries,
			metricQuery: cfg.MetricQuery,
		},
		AlertCheck: MonitorAlertCheckJob{
			sequence:    cfg.Sequence,
			repository:  repository,
			xxlExec:     cfg.XxlJobExec,
			timeSeries:  cfg.TimeSeries,
			metricQuery: cfg.MetricQuery,
			alertConf:   app.AlertConf,
		},
		EventCheck: MonitorEventCheckJob{
			sequence:   cfg.Sequence,
			repository: repository,
			xxlExec:    cfg.XxlJobExec,
			alertConf:  app.AlertConf,
			commonMap:  app.CommonMap,
		},
	}
}

func (j *Job) RegisterJobExec() {
	j.DataCollect.RegisterExecJob()
	j.DataSampling.RegisterExecJob()
	j.AlertCheck.RegisterExecJob()
	j.EventCheck.RegisterExecJob()
}
