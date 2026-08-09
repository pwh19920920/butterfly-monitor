package job

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"
	"butterfly-monitor/internal/infrastructure/persistence"
	"butterfly-monitor/internal/types"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorAlertCheckJob 告警规则检查任务
type MonitorAlertCheckJob struct {
	sequence      *snowflake.Node
	repository    *persistence.Repository
	xxlExec       xxl.Executor
	timeSeries    domainHandler.TimeSeriesStore
	metricQuery   domainHandler.MetricQueryDialect
	alertConf     application.AlertConfApplication
	volatilityDay application.MonitorVolatilityDayApplication
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

	specialDays, _ := job.volatilityDay.SelectAll(ctx)
	// 并发上限：信号量限制同时执行的规则数，压住 VM 瞬时 QPS（大规模时避免瞬间打爆）
	concurrency := conf.AlertCheckConcurrency
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for _, check := range checks {
		wg.Add(1)
		sem <- struct{}{} // 占坑；并发满则阻塞排队
		go func(c entity.MonitorTaskAlert) {
			defer func() { <-sem }() // 释放坑位
			// execCheck 内部已 defer wg.Done()（含 recoverLog 兜底），此处不再 Done，避免负计数
			job.execCheck(ctx, c, conf, specialDays, &wg)
		}(check)
	}
	wg.Wait()
	return "execute complete"
}

// execCheck 单条告警规则检查主流程。
// 流程：取任务 → 解析规则 → (波动日放大阈值) → 并发取实时/样本均值 → 判定命中 → 持续时长 → 落库状态机
func (job *MonitorAlertCheckJob) execCheck(ctx context.Context, check entity.MonitorTaskAlert, conf *types.AlertConfObject, specialDays []entity.MonitorVolatilityDay, wg *sync.WaitGroup) {
	defer wg.Done()
	defer recoverLog("alertCheck")

	task, err := job.repository.MonitorTaskRepository.GetById(check.TaskId)
	if err != nil || task == nil {
		// 任务不存在/已删：仍推进 PreCheckTime，避免无效规则每轮被捞起空跑
		job.advancePreCheckTime(ctx, check.Id, "task missing")
		return
	}
	// 任务/告警任一关闭则跳过（同样推进检查时间）
	if task.AlertStatus == entity.MonitorAlertStatusClose || task.TaskStatus == entity.MonitorTaskStatusClose {
		job.advancePreCheckTime(ctx, check.Id, "task/alert closed")
		return
	}
	// 解析规则组；空配置或非法 JSON 直接跳过
	var params []entity.MonitorAlertCheckParams
	if check.Params == "" {
		job.advancePreCheckTime(ctx, check.Id, "empty params")
		return
	}
	if err := json.Unmarshal([]byte(check.Params), &params); err != nil {
		job.advancePreCheckTime(ctx, check.Id, "invalid params")
		return
	}

	now := time.Now()
	// 敏感任务命中波动日：仅本轮判定副本放大样本差阈值（不改库、不改用户配置）
	params = job.applyPromoRatioIfHit(ctx, params, task, conf, specialDays, now)

	// 查询窗口：以 timeSpan 为粒度向前回溯，取最近一个完整窗口
	start := now.Add(-time.Duration(2*task.TimeSpan) * time.Second)
	end := now.Add(-time.Duration(task.TimeSpan) * time.Second)
	sampleMetric, realtimeMetric := job.resolveMetrics(task.TaskKey)

	sampleVal, realVal := job.querySampleAndReal(ctx, sampleMetric, realtimeMetric, start, end)

	// 实时值缺失：视为异常，标 Pending（待数据恢复后再判定）
	if realVal == nil {
		job.markPending(ctx, check.Id, now, "missing")
		return
	}
	// 规则依赖样本但样本缺失：无法可靠判定，保守标 Pending，避免样本差规则被整体跳过而静默误报正常
	if sampleVal == nil && rulesNeedSample(params) {
		logger.InfoFormat(ctx, "告警检查样本缺失, 规则依赖样本, 保守标Pending task=%s checkId=%d", task.TaskKey, check.Id)
		job.markPending(ctx, check.Id, now, "sample missing")
		return
	}

	// 持续时长：从「本轮首次发现异常」到现在
	// - 当前仍是 Normal：本轮第一次命中，起点=现在，已持续 0 秒
	// - 已是 Pending/Firing：用 FirstFlagTime 作为起点
	elapsedSec := calcElapsedSec(check, now)
	hit, level, msg := evaluateRules(params, sampleVal, *realVal, now, elapsedSec)
	if !hit {
		job.markNormal(ctx, check.Id, now)
		return
	}
	// 未达到配置的持续阈值，仅标 Pending，不发事件
	if elapsedSec < check.Duration {
		job.markPending(ctx, check.Id, now, "duration")
		return
	}

	// Firing：创建/更新告警事件
	firstDelay := firstAlertDelay(conf)
	// 命中波动日且阈值被放大时，告警文案附上大促上下文，便于值班核对
	event := &entity.MonitorTaskEvent{
		BaseEntity:    common.BaseEntity{Id: job.sequence.Generate().Int64()},
		AlertId:       check.Id,
		TaskId:        task.Id,
		AlertMsg:      msg + promoAlertNote(application.MatchVolatilityDay(specialDays, now), conf, task.PromoSensitive == entity.PromoSensitiveOn),
		DealStatus:    entity.MonitorTaskEventDealStatusPending,
		NextAlertTime: &common.LocalTime{Time: now.Add(time.Duration(firstDelay) * time.Second)},
		EventLevel:    level,
	}
	if err := job.repository.MonitorTaskAlertRepository.ModifyByFiring(check.Id, now, event); err != nil {
		logger.ErrorFormat(ctx, "alert ModifyByFiring fail checkId=%d: %v", check.Id, err)
	}
}

// promoAlertNote 命中波动日且阈值被放大时，返回附在告警文案上的大促上下文；否则空串。
// 倍数<=1 视为未放大，不加注，避免告警文案出现无效的「策略生效」。
func promoAlertNote(hit *entity.MonitorVolatilityDay, conf *types.AlertConfObject, promoSensitive bool) string {
	if hit == nil || conf == nil || !promoSensitive {
		return ""
	}
	var ratio float64
	if hit.Type == entity.VolatilityDayTypeTrough {
		ratio = conf.PromoTroughRatio
	} else {
		ratio = conf.PromoPeakRatio
	}
	if ratio <= 1 {
		return ""
	}
	return fmt.Sprintf(", 波动策略[%s %s, 阈值放大 %.2fx]", hit.Name, hit.Type, ratio)
}

// applyPromoRatioIfHit 敏感任务命中波动日时，按波动日类型放大阈值（仅本轮副本，不写库）。
// 非敏感任务 / 未命中波动日 / ratio<=1 时原样返回。
func (job *MonitorAlertCheckJob) applyPromoRatioIfHit(ctx context.Context, params []entity.MonitorAlertCheckParams, task *entity.MonitorTask, conf *types.AlertConfObject, specialDays []entity.MonitorVolatilityDay, now time.Time) []entity.MonitorAlertCheckParams {
	hit := application.MatchVolatilityDay(specialDays, now)
	if hit == nil || task.PromoSensitive == entity.PromoSensitiveOff {
		return params
	}

	peakRatio := conf.PromoPeakRatio
	if peakRatio <= 0 {
		peakRatio = 1
	}
	troughRatio := conf.PromoTroughRatio
	if troughRatio <= 0 {
		troughRatio = 1
	}
	logger.InfoFormat(ctx, "告警检查命中波动日, 走放大阈值处理 task=%s rule=[%s] type=%s period=%s~%s peakRatio=%v troughRatio=%v",
		task.TaskKey, hit.Name, hit.Type,
		hit.StartTime.Time.Format("2006-01-02 15:04:05"), hit.EndTime.Time.Format("2006-01-02 15:04:05"),
		peakRatio, troughRatio)
	return applyPromoAlertRatio(params, hit.Type, peakRatio, troughRatio)
}

// resolveMetrics 解析样本/实时指标名。metricQuery 为 nil 时回退默认后缀命名。
func (job *MonitorAlertCheckJob) resolveMetrics(taskKey string) (sampleMetric, realtimeMetric string) {
	sampleMetric = taskKey + "_sample"
	realtimeMetric = taskKey
	if job.metricQuery != nil {
		sampleMetric = job.metricQuery.SmoothMetric(taskKey)
		realtimeMetric = job.metricQuery.RealtimeMetric(taskKey)
	}
	return sampleMetric, realtimeMetric
}

// querySampleAndReal 并发查询样本与实时均值。timeSeries 为 nil 时两者均返回 nil。
func (job *MonitorAlertCheckJob) querySampleAndReal(ctx context.Context, sampleMetric, realtimeMetric string, start, end time.Time) (sampleVal, realVal *float64) {
	if job.timeSeries == nil {
		return nil, nil
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sampleVal, _ = job.timeSeries.QueryMean(ctx, sampleMetric, start, end)
	}()
	go func() {
		defer wg.Done()
		realVal, _ = job.timeSeries.QueryMean(ctx, realtimeMetric, start, end)
	}()
	wg.Wait()
	return sampleVal, realVal
}

// calcElapsedSec 计算本轮异常已持续秒数。
// Normal 或无首次异常时间：本轮首次命中，已持续 0 秒；否则用 FirstFlagTime 作为起点。
func calcElapsedSec(check entity.MonitorTaskAlert, now time.Time) int64 {
	if check.AlertStatus == entity.MonitorTaskAlertStatusNormal || check.FirstFlagTime == nil {
		return 0
	}
	elapsed := int64(now.Sub(check.FirstFlagTime.Time).Seconds())
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

// firstAlertDelay 首次告警事件的下一次外发延迟（秒）。
func firstAlertDelay(conf *types.AlertConfObject) int64 {
	if conf == nil {
		return types.DefaultFirstDelay
	}
	return conf.FirstDelay
}

// markPending 标记 Pending，失败统一打日志。reason 仅用于日志区分缺失/持续时长等场景。
func (job *MonitorAlertCheckJob) markPending(ctx context.Context, checkId int64, now time.Time, reason string) {
	if err := job.repository.MonitorTaskAlertRepository.ModifyByPending(checkId, now); err != nil {
		logger.ErrorFormat(ctx, "alert ModifyByPending(%s) fail checkId=%d: %v", reason, checkId, err)
	}
}

// markNormal 恢复正常，失败统一打日志。
func (job *MonitorAlertCheckJob) markNormal(ctx context.Context, checkId int64, now time.Time) {
	if err := job.repository.MonitorTaskAlertRepository.ModifyForNormal(checkId, now); err != nil {
		logger.ErrorFormat(ctx, "alert ModifyForNormal fail checkId=%d: %v", checkId, err)
	}
}

// advancePreCheckTime 仅推进检查时间，不改 AlertStatus。
// 用于「本轮不可检」的跳过路径（任务关闭/缺失、规则空或非法），避免 FindCheckJob 每轮空捞。
func (job *MonitorAlertCheckJob) advancePreCheckTime(ctx context.Context, checkId int64, reason string) {
	now := time.Now()
	if err := job.repository.MonitorTaskAlertRepository.Modify(checkId, &entity.MonitorTaskAlert{
		PreCheckTime: &common.LocalTime{Time: now},
	}); err != nil {
		logger.ErrorFormat(ctx, "alert advancePreCheckTime(%s) fail checkId=%d: %v", reason, checkId, err)
	}
}
