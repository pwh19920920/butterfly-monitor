package job

import (
	"context"
	"strconv"
	"sync"
	"time"

	"butterfly-monitor/internal/application"
	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"
	"butterfly-monitor/internal/infrastructure/persistence"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorDataSamplingJob 样本平滑任务
type MonitorDataSamplingJob struct {
	sequence      *snowflake.Node
	repository    *persistence.Repository
	xxlExec       xxl.Executor
	timeSeries    domainHandler.TimeSeriesStore
	metricQuery   domainHandler.MetricQueryDialect
	alertConf     application.AlertConfApplication
	volatilityDay application.MonitorVolatilityDayApplication
}

func (job *MonitorDataSamplingJob) RegisterExecJob() {
	if job.xxlExec == nil {
		logrus.Warn("xxl executor is nil, skip register dataSampling")
		return
	}
	job.xxlExec.RegTask("dataSampling", job.ExecDataSampling)
}

func (job *MonitorDataSamplingJob) ExecDataSampling(ctx context.Context, param *xxl.RunReq) string {
	conf, err := job.alertConf.Cover2AlertConf(ctx)
	if err != nil {
		return "exec failure conf"
	}
	specialDays, _ := job.volatilityDay.SelectAll(ctx)

	// 字段已由 Cover2AlertConf 装配并规范化（含默认值），直接取用
	pageSize := conf.SimplePageSize
	maxSecond := conf.SimpleMaxSecond
	// 冻结基线回溯天数：敏感任务波动日从「前一日」起找普通日同时刻 _sample
	freezeLookBackDays := conf.FreezeSampleLookBackDays
	// 单批内并发上限：串行改并发，缩短单轮耗时，同时用信号量压住 VM 瞬时 QPS
	concurrency := conf.SamplingConcurrency

	var lastId int64
	for {
		tasks, err := job.repository.MonitorTaskRepository.FindSamplingJobBySharding(pageSize, lastId, param.BroadcastIndex, param.BroadcastTotal)
		if err != nil || len(tasks) == 0 {
			break
		}
		// 批内并发采样：按 id asc 排序，本批全部完成后取最大 id 推进游标，保证不遗漏
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		batchMax := lastId
		for _, task := range tasks {
			wg.Add(1)
			sem <- struct{}{} // 占坑；并发满则阻塞排队
			go func(t entity.MonitorTask) {
				defer wg.Done()
				defer func() { <-sem }() // 释放坑位
				job.sampleOne(ctx, t, maxSecond, freezeLookBackDays, specialDays)
			}(task)
			if task.Id > batchMax {
				batchMax = task.Id
			}
		}
		wg.Wait()
		lastId = batchMax
		if int64(len(tasks)) < pageSize {
			break
		}
	}
	return "execute complete"
}

// sampleOne 对单个任务按 time_span 逐格补样本，直到追到 now+1d。
//
// 并发约束：只要任务处于开启状态（task_status=open），dataSampling 就会持续捞它执行，
// 与是否有新数据无关。因此 sampleOne 必须「一轮追完」而非「每轮只补一格」，
// 否则任何一次抖动让 pre_sample_time 落后超过 maxSecond 后，旧逻辑的 begin 钳制
// 会把起点吸回 now-maxSecond 并丢掉真实追赶距离，导致 pre_sample_time 永远停在
// now-maxSecond 附近震荡 —— 即「持续执行却追不上」的死循环。
//
// 修正后：
//   - begin 始终取真实 PreSampleTime，绝不钳制，落后就一气追完
//   - 按 TimeSpan 逐格产出 *_sample 点，直到下一格会超过 now+1d
//   - 无数据格跳过写点但 begin 仍推进，避免无数据造成新积压
//   - maxSecond 不再钳制起点，转为「单轮最多补点格数」安全阀（≈ maxSecond/TimeSpan），
//     防止服务停几天后一轮打爆；追不动部分留到下一轮，但不丢段
//   - 取数侧 FindSamplingJobBySharding 也按 now+1d 判断到期，否则 PreSampleTime 一过
//     当前时刻就不再入选，+1d 预生成无法跑起来
//
// 大促（仅 PromoSensitive=2）：
//   - 原料按 day tag 还原来源日，特殊日来源点不参与聚合
//   - 当前格 end 落在特殊日：从 _sample 取「前一日起、最近普通日同时刻」写入冻结基线
//   - 不回退标量（错时刻对照比缺样本更危险）
//   - freezeLookBackDays 来自 alertConf.freezeSampleLookBackDays（默认 14）
func (job *MonitorDataSamplingJob) sampleOne(ctx context.Context, task entity.MonitorTask, maxSecond, freezeLookBackDays int64, specialDays []entity.MonitorVolatilityDay) {
	defer recoverLog("dataSampling")
	// 聚合任务只收集多维点，不生成样本基线（查询侧已排除，这里再做入口守卫）
	if task.DataType == entity.DataTypeAggregate {
		return
	}
	if job.timeSeries == nil {
		return
	}
	spanSec := int64(task.TimeSpan)
	if spanSec <= 0 {
		return
	}

	now := time.Now()
	begin := now
	if task.PreSampleTime != nil {
		begin = task.PreSampleTime.Time
	}
	span := time.Duration(spanSec) * time.Second

	sampleRawMetric, smooth := job.resolveSampleMetrics(task.TaskKey)
	promoSensitive := task.PromoSensitive == entity.PromoSensitiveOn
	isSpecial := func(t time.Time) bool {
		return application.MatchVolatilityDay(specialDays, t) != nil
	}

	// 单轮最多补点格数安全阀：maxSecond/TimeSpan 向上取整，至少 1 格。
	maxLoop := (maxSecond + spanSec - 1) / spanSec
	if maxLoop < 1 {
		maxLoop = 1
	}
	cutoff := now.Add(24 * time.Hour)
	// sampleErr 非 nil 时：本格失败，begin 停在失败格起点，与采集侧「写失败不推进时间」对齐
	var sampleErr error

	for i := int64(0); i < maxLoop; i++ {
		end := begin.Add(span)
		if end.After(cutoff) {
			break
		}

		// 敏感任务 + 当前格命中波动日：用前序普通日同时刻 _sample 冻结，不聚合大促原料。
		if hit := application.MatchVolatilityDay(specialDays, end); hit != nil && promoSensitive {
			freezeVal, err := job.resolveFreezeFromPrevSample(ctx, smooth, end, span, freezeLookBackDays, specialDays)
			if err != nil {
				// 时序查询失败不推进，下一轮重试
				sampleErr = err
				break
			}
			if freezeVal != nil {
				if err := job.writeFreezeBaseline(ctx, task.TaskKey, smooth, *freezeVal, end, hit); err != nil {
					sampleErr = err
					break
				}
			} else {
				// 回溯窗口内无普通日同时刻样本（冷启动/历史空洞）：本格不写，推进窗口避免卡死；
				// 告警侧 sample 缺失会对样本差规则保守 Pending。
				logger.InfoFormat(ctx, "采样命中波动日但无前序普通日同时刻样本, 本格跳过冻结 task=%s rule=[%s] cellEnd=%s",
					task.TaskKey, hit.Name, end.Format("2006-01-02 15:04:05"))
			}
			begin = end
			continue
		}

		// 原算法聚合原料（非敏感 / 普通日）
		vals, err := job.loadSampleRawValues(ctx, sampleRawMetric, begin, end, promoSensitive, isSpecial)
		if err != nil {
			// 查询失败不推进：下一轮从本格重试，避免静默跳格导致基线空洞
			logger.ErrorFormat(ctx, "query sample range fail task=%s: %v", task.TaskKey, err)
			sampleErr = err
			break
		}
		if len(vals) > 0 {
			avg := averageSampleValues(vals)
			if err := job.writeSamplePoint(ctx, task.TaskKey, smooth, avg, end); err != nil {
				// 写失败不推进 begin，本格留给下一轮重试
				sampleErr = err
				break
			}
		}
		// 无数据格：跳过写点但 begin 仍推进，避免无数据造成新积压
		begin = end
	}

	// 回写进度：成功格已推进的 begin；失败时停在失败格起点，便于重试
	upd := &entity.MonitorTask{
		PreSampleTime: &common.LocalTime{Time: begin},
	}
	if sampleErr != nil {
		upd.SampleErrMsg = sampleErr.Error()
	} else {
		// 空格清空历史错误（GORM Updates 跳过空串，用空格占位）
		upd.SampleErrMsg = " "
	}
	if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, upd); err != nil {
		logger.ErrorFormat(ctx, "update PreSampleTime fail task=%s: %v", task.TaskKey, err)
	}
}

// resolveSampleMetrics 解析样本原料/平滑指标名。metricQuery 为 nil 时回退默认后缀命名。
func (job *MonitorDataSamplingJob) resolveSampleMetrics(taskKey string) (sampleRawMetric, smooth string) {
	sampleRawMetric = taskKey + "_sampling"
	smooth = taskKey + "_sample"
	if job.metricQuery != nil {
		sampleRawMetric = job.metricQuery.SampleRawMetric(taskKey)
		smooth = job.metricQuery.SmoothMetric(taskKey)
	}
	return sampleRawMetric, smooth
}

// resolveFreezeFromPrevSample 取冻结基线：从 end 的前一日起，找最近一个「普通日、同时刻」的 _sample。
// lookBackDays 为最大回溯天数（来自 alertConf.freezeSampleLookBackDays，默认 14）；<=0 时按 14。
// 跳过仍为波动日的日期，避免连续大促把冻结值接到脏基线/另一冻结值上。
// 不回退标量。查无返回 (nil, nil)；时序错误返回 err。
func (job *MonitorDataSamplingJob) resolveFreezeFromPrevSample(
	ctx context.Context,
	smoothMetric string,
	end time.Time,
	span time.Duration,
	lookBackDays int64,
	specialDays []entity.MonitorVolatilityDay,
) (*float64, error) {
	if lookBackDays <= 0 {
		lookBackDays = 14
	}
	for d := int64(1); d <= lookBackDays; d++ {
		refEnd := end.AddDate(0, 0, -int(d))
		if application.MatchVolatilityDay(specialDays, refEnd) != nil {
			continue
		}
		refStart := refEnd.Add(-span)
		v, err := job.timeSeries.QueryMean(ctx, smoothMetric, refStart, refEnd)
		if err != nil {
			logger.ErrorFormat(ctx, "query freeze prev sample fail metric=%s refEnd=%s: %v",
				smoothMetric, refEnd.Format("2006-01-02 15:04:05"), err)
			return nil, err
		}
		if v != nil {
			return v, nil
		}
	}
	return nil, nil
}

// writeFreezeBaseline 写波动日冻结基线（值为前序普通日同时刻 _sample）。
// 写失败返回 error，由 sampleOne 决定不推进 PreSampleTime。
func (job *MonitorDataSamplingJob) writeFreezeBaseline(ctx context.Context, taskKey, metric string, value float64, end time.Time, hit *entity.MonitorVolatilityDay) error {
	logger.InfoFormat(ctx, "采样命中波动日, 使用前序普通日同时刻样本冻结 task=%s rule=[%s] type=%s period=%s~%s freezeValue=%v cellEnd=%s",
		taskKey, hit.Name, hit.Type,
		hit.StartTime.Time.Format("2006-01-02 15:04:05"), hit.EndTime.Time.Format("2006-01-02 15:04:05"),
		value, end.Format("2006-01-02 15:04:05"))
	if err := job.timeSeries.WritePoints(ctx, []domainHandler.TimeSeriesPoint{{
		Metric: metric, Value: value, Timestamp: end,
	}}); err != nil {
		logger.ErrorFormat(ctx, "write freeze baseline fail task=%s: %v", taskKey, err)
		return err
	}
	return nil
}

// writeSamplePoint 写单个样本平滑点到时序库。
func (job *MonitorDataSamplingJob) writeSamplePoint(ctx context.Context, taskKey, metric string, value float64, end time.Time) error {
	if err := job.timeSeries.WritePoints(ctx, []domainHandler.TimeSeriesPoint{{
		Metric: metric, Value: value, Timestamp: end,
	}}); err != nil {
		logger.ErrorFormat(ctx, "write smooth baseline fail task=%s: %v", taskKey, err)
		return err
	}
	return nil
}

// loadSampleRawValues 拉取本格原料值。
// 敏感任务：按 series 的 day tag 还原来源日，剔除特殊日来源后再展平；
// 非敏感或无 day 信息：退回 QueryRangeValues 全量。
func (job *MonitorDataSamplingJob) loadSampleRawValues(
	ctx context.Context,
	metric string,
	begin, end time.Time,
	promoSensitive bool,
	isSpecial func(time.Time) bool,
) ([]float64, error) {
	if !promoSensitive {
		return job.timeSeries.QueryRangeValues(ctx, metric, begin, end)
	}
	// 按 tag 拉 series，用 day 标签过滤来源特殊日
	series, err := job.timeSeries.QueryRangeWithTags(ctx, metric, begin, end, nil)
	if err != nil {
		return nil, err
	}
	if len(series) == 0 {
		return nil, nil
	}

	points, hasDay := seriesToSamplePoints(series)
	if !hasDay {
		// 无 day tag（如 TDengine 当前实现）：退回全量，避免误杀
		return pointsToValues(points), nil
	}

	filtered := filterSpecialSourceDays(points, end, isSpecial)
	// 剔光后仍要保线：回退未过滤值（调用方 averageSampleValues 仍会写）
	if len(filtered) == 0 {
		return pointsToValues(points), nil
	}
	return filtered, nil
}

// seriesToSamplePoints 将多 series 数据展平为带 day tag 的原料点列表。
// hasDay 表示是否至少有一个 series 携带了有效的 day 标签。
func seriesToSamplePoints(series []domainHandler.SeriesData) ([]samplePoint, bool) {
	points := make([]samplePoint, 0)
	hasDay := false
	for _, s := range series {
		dayOff := 0
		if s.Labels != nil {
			if d, ok := s.Labels["day"]; ok && d != "" {
				if n, e := strconv.Atoi(d); e == nil && n > 0 {
					dayOff = n
					hasDay = true
				}
			}
		}
		for _, v := range s.Values {
			points = append(points, samplePoint{Value: v, DayOffset: dayOff})
		}
	}
	return points, hasDay
}

// pointsToValues 从原料点列表提取纯值切片。
func pointsToValues(points []samplePoint) []float64 {
	vals := make([]float64, 0, len(points))
	for _, p := range points {
		vals = append(vals, p.Value)
	}
	return vals
}
