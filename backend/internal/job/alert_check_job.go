package job

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"
	"dragonfly-monitor/internal/infrastructure/persistence"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorAlertCheckJob 告警规则检查任务
type MonitorAlertCheckJob struct {
	sequence    *snowflake.Node
	repository  *persistence.Repository
	xxlExec     xxl.Executor
	timeSeries  domainHandler.TimeSeriesStore
	metricQuery domainHandler.MetricQueryDialect
	alertConf   application.AlertConfApplication
}

func (job *MonitorAlertCheckJob) RegisterExecJob() {
	if job.xxlExec == nil {
		return
	}
	job.xxlExec.RegTask("alertCheck", job.alertCheck)
}

func (job *MonitorAlertCheckJob) alertCheck(ctx context.Context, param *xxl.RunReq) string {
	checks, err := job.repository.MonitorTaskAlertRepository.FindCheckJob(param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		return "exec failure"
	}
	conf, err := job.alertConf.Cover2AlertConf(ctx)
	if err != nil {
		return "exec failure conf"
	}
	var wg sync.WaitGroup
	for _, check := range checks {
		wg.Add(1)
		go job.execCheck(ctx, check, conf.FirstDelay, &wg)
	}
	wg.Wait()
	return "execute complete"
}

func (job *MonitorAlertCheckJob) execCheck(ctx context.Context, check entity.MonitorTaskAlert, firstDelay int64, wg *sync.WaitGroup) {
	defer wg.Done()
	defer recoverLog("alertCheck")

	task, err := job.repository.MonitorTaskRepository.GetById(check.TaskId)
	if err != nil || task == nil {
		return
	}
	if task.AlertStatus == entity.MonitorAlertStatusClose || task.TaskStatus == entity.MonitorTaskStatusClose {
		return
	}
	if check.Params == "" {
		return
	}
	var params []entity.MonitorAlertCheckParams
	if err := json.Unmarshal([]byte(check.Params), &params); err != nil {
		return
	}
	now := time.Now()
	start := now.Add(-time.Duration(2*task.TimeSpan) * time.Second)
	end := now.Add(-time.Duration(task.TimeSpan) * time.Second)

	sampleMetric := task.TaskKey + "_sample"
	realtimeMetric := task.TaskKey
	if job.metricQuery != nil {
		sampleMetric = job.metricQuery.SmoothMetric(task.TaskKey)
		realtimeMetric = job.metricQuery.RealtimeMetric(task.TaskKey)
	}
	var sampleVal, realVal *float64
	if job.timeSeries != nil {
		sampleVal, _ = job.timeSeries.QueryMean(ctx, sampleMetric, start, end)
		realVal, _ = job.timeSeries.QueryMean(ctx, realtimeMetric, start, end)
	}
	if realVal == nil {
		// 实时值缺失视为异常 low
		if err := job.repository.MonitorTaskAlertRepository.ModifyByPending(check.Id, now); err != nil {
			logger.ErrorFormat(ctx, "alert ModifyByPending(missing) fail checkId=%d: %v", check.Id, err)
		}
		return
	}
	// 先判定是否命中规则（文案稍后用实际持续秒数生成）
	hit, level, _ := evaluateRules(params, sampleVal, *realVal, now, 0)
	if !hit {
		if err := job.repository.MonitorTaskAlertRepository.ModifyForNormal(check.Id, now); err != nil {
			logger.ErrorFormat(ctx, "alert ModifyForNormal fail checkId=%d: %v", check.Id, err)
		}
		return
	}
	// 持续时长：从「本轮首次发现异常」到现在
	// - 当前仍是 Normal：本轮第一次命中，起点=现在，已持续 0 秒
	// - 已是 Pending/Firing：用 FirstFlagTime 作为起点
	var elapsedSec int64
	if check.AlertStatus == entity.MonitorTaskAlertStatusNormal || check.FirstFlagTime == nil {
		elapsedSec = 0
	} else {
		elapsedSec = int64(now.Sub(check.FirstFlagTime.Time).Seconds())
		if elapsedSec < 0 {
			elapsedSec = 0
		}
	}
	// 未达到配置的持续阈值，仅标记 Pending，不发事件
	if elapsedSec < check.Duration {
		if err := job.repository.MonitorTaskAlertRepository.ModifyByPending(check.Id, now); err != nil {
			logger.ErrorFormat(ctx, "alert ModifyByPending(duration) fail checkId=%d: %v", check.Id, err)
		}
		return
	}
	// 用「现在 - 首次异常」的实际持续秒数生成 hitRule 文案
	_, _, msg := evaluateRules(params, sampleVal, *realVal, now, elapsedSec)
	// Firing
	if firstDelay <= 0 {
		firstDelay = 60
	}
	event := &entity.MonitorTaskEvent{
		BaseEntity:    common.BaseEntity{Id: job.sequence.Generate().Int64()},
		AlertId:       check.Id,
		TaskId:        task.Id,
		AlertMsg:      msg,
		DealStatus:    entity.MonitorTaskEventDealStatusPending,
		NextAlertTime: &common.LocalTime{Time: now.Add(time.Duration(firstDelay) * time.Second)},
		EventLevel:    level,
	}
	if err := job.repository.MonitorTaskAlertRepository.ModifyByFiring(check.Id, now, event); err != nil {
		logger.ErrorFormat(ctx, "alert ModifyByFiring fail checkId=%d: %v", check.Id, err)
	}
}
