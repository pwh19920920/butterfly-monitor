package job

import (
	"context"
	"strconv"
	"time"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"
	"dragonfly-monitor/internal/infrastructure/persistence"

	"github.com/pwh19920920/butterfly/pkg/logger"
	"github.com/pwh19920920/snowflake"
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorDataSamplingJob 样本平滑任务
type MonitorDataSamplingJob struct {
	sequence    *snowflake.Node
	repository  *persistence.Repository
	xxlExec     xxl.Executor
	timeSeries  domainHandler.TimeSeriesStore
	metricQuery domainHandler.MetricQueryDialect
}

func (job *MonitorDataSamplingJob) RegisterExecJob() {
	if job.xxlExec == nil {
		logrus.Warn("xxl executor is nil, skip register dataSampling")
		return
	}
	job.xxlExec.RegTask("dataSampling", job.ExecDataSampling)
}

func (job *MonitorDataSamplingJob) ExecDataSampling(ctx context.Context, param *xxl.RunReq) string {
	conf, _ := job.repository.AlertConfRepository.SelectAll()
	pageSize := int64(50)
	maxSecond := int64(600)
	for _, c := range conf {
		if c.ConfKey == "simplePageSize" {
			// 用 strconv.Atoi 健壮解析配置值，失败时保留默认 pageSize
			if v, err := strconv.Atoi(c.ConfVal); err == nil {
				pageSize = int64(v)
			}
		}
		if c.ConfKey == "simpleMaxSecond" {
			// 用 strconv.Atoi 健壮解析配置值，失败时保留默认 maxSecond
			if v, err := strconv.Atoi(c.ConfVal); err == nil {
				maxSecond = int64(v)
			}
		}
	}
	var lastId int64
	for {
		tasks, err := job.repository.MonitorTaskRepository.FindSamplingJobBySharding(pageSize, lastId, param.BroadcastIndex, param.BroadcastTotal)
		if err != nil || len(tasks) == 0 {
			break
		}
		for _, task := range tasks {
			job.sampleOne(ctx, task, maxSecond)
			if task.Id > lastId {
				lastId = task.Id
			}
		}
		if int64(len(tasks)) < pageSize {
			break
		}
	}
	return "execute complete"
}

// sampleOne 对单个任务按 time_span 逐格补样本，直到追到现在。
//
// 关键约束：只要任务处于开启状态（task_status=open），dataSampling 就会持续捞它执行，
// 与是否有新数据无关。因此 sampleOne 必须「一轮追完」而非「每轮只补一格」，
// 否则任何一次抖动让 pre_sample_time 落后超过 maxSecond 后，旧逻辑的 begin 钳制
// 会把起点吸回 now-maxSecond 并丢掉真实追赶距离，导致 pre_sample_time 永远停在
// now-maxSecond 附近震荡 —— 即「持续执行却追不上」的死循环。
//
// 修正后：
//   - begin 始终取真实 PreSampleTime，绝不钳制，落后就一气追完
//   - 按 TimeSpan 逐格产出 *_sample 点，直到下一格会超过 now
//   - 无数据格跳过写点但 begin 仍推进，避免无数据造成新积压
//   - maxSecond 不再钳制起点，转为「单轮最多补点格数」安全阀（≈ maxSecond/TimeSpan），
//     防止服务停几天后一轮打爆；追不动部分留到下一轮，但不丢段
func (job *MonitorDataSamplingJob) sampleOne(ctx context.Context, task entity.MonitorTask, maxSecond int64) {
	defer recoverLog("dataSampling")
	if job.timeSeries == nil {
		return
	}
	spanSec := int64(task.TimeSpan)
	if spanSec <= 0 {
		return
	}
	now := time.Now()
	// begin 始终取真实起点，不钳制；首次采样 PreSampleTime 为 nil 时从 now 起只往后补不受影响
	begin := now
	if task.PreSampleTime != nil {
		begin = task.PreSampleTime.Time
	}
	span := time.Duration(spanSec) * time.Second

	// metric 名计算下沉到循环外，单次构建复用
	// 注：realtimeMetric 不再需要 —— realtime 回退已移除，基线只来源于 *_sampling 历史原料
	sampleRawMetric := task.TaskKey + "_sampling"
	smooth := task.TaskKey + "_sample"
	if job.metricQuery != nil {
		sampleRawMetric = job.metricQuery.SampleRawMetric(task.TaskKey)
		smooth = job.metricQuery.SmoothMetric(task.TaskKey)
	}

	// 单轮最多补点格数安全阀：maxSecond/TimeSpan 向上取整，至少 1 格。
	// maxSecond 语义为「秒」，按 TimeSpan 折算为格数，保留原配置 key 兼容存量数据。
	// 超过该上限则本轮只补到上限，追不动部分留到下一轮，但不丢段。
	maxLoop := int64(1)
	if maxSecond > 0 {
		maxLoop = (maxSecond + spanSec - 1) / spanSec
		if maxLoop < 1 {
			maxLoop = 1
		}
	}

	// 逐格补点直到 now 之后一天：end <= now+1day 才可产出。
	// 不再保留安全间距：realtime 回退已移除，基线不会混入当前实时值，无需避让它；
	// 配合 _sampling 的"未来 1~8 天预填"机制，可提前把未来一天的格子也生成出来。
	cutoff := now.Add(24 * time.Hour)
	for end := begin.Add(span); !end.After(cutoff); end = begin.Add(span) {
		// 优先用历史原料 *_sampling（未来 1~8 天滚动预填的点）
		vals, err := job.timeSeries.QueryRangeValues(ctx, sampleRawMetric, begin, end)
		if err != nil {
			// 查询失败：跳过该格，不写基线（避免用不可信数据污染历史）
			logger.ErrorFormat(ctx, "query sample range fail task=%s: %v", task.TaskKey, err)
		} else if len(vals) > 0 {
			// 仅在有样本时聚合写点；无样本跳过该格（告警端 sampleVal==nil 会安全降级）
			avg := averageSampleValues(vals)
			if wErr := job.timeSeries.WritePoints(ctx, []domainHandler.TimeSeriesPoint{{
				Metric: smooth, Value: avg, Timestamp: end,
			}}); wErr != nil {
				// 平滑基线写入失败：begin 仍推进，后续告警判定基于缺失基线可能误报/漏报
				logger.ErrorFormat(ctx, "write smooth baseline fail task=%s: %v", task.TaskKey, wErr)
			}
		}
		// 无论是否写入，均推进 begin，避免在「任务一直采样」模式下无数据造成新积压
		begin = end
		// 安全阀：本轮追不动部分留到下一轮，不丢段
		if maxLoop--; maxLoop <= 0 {
			break
		}
	}

	// 推进 PreSampleTime 到本轮追到的最新位置（距 now < 一格），下一轮续追
	if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		PreSampleTime: &common.LocalTime{Time: begin},
	}); err != nil {
		logger.ErrorFormat(ctx, "update PreSampleTime fail task=%s: %v", task.TaskKey, err)
		return
	}
	// 清空样本错误标记（用空格非空串，规避 Updates 零值忽略）
	if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		SampleErrMsg: " ",
	}); err != nil {
		logger.ErrorFormat(ctx, "clear SampleErrMsg fail task=%s: %v", task.TaskKey, err)
	}
}
