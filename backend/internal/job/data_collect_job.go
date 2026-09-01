package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	"github.com/sirupsen/logrus"
	"github.com/xxl-job/xxl-job-executor-go"
)

// MonitorDataCollectJob 数据采集任务
type MonitorDataCollectJob struct {
	sequence    *snowflake.Node
	repository  *persistence.Repository
	xxlExec     xxl.Executor
	timeSeries  domainHandler.TimeSeriesStore
	metricQuery domainHandler.MetricQueryDialect
	commonMap   *application.CommonMapApplication
	alertConf   application.AlertConfApplication
}

func (job *MonitorDataCollectJob) RegisterExecJob() {
	if job.xxlExec == nil {
		logrus.Warn("xxl executor is nil, skip register dataCollect")
		return
	}
	job.xxlExec.RegTask("dataCollect", job.ExecDataCollect)
}

// ExecDataCollect 数据采集调度入口。
//
// 落后恢复策略（与 dataSampling 刻意相反）：
//   - 任务卡住 / 超时 / 写失败 / 服务停机后，只要还能采，就只采「当前时刻」最近一个窗口，
//     绝不从 PreExecuteTime 往前递归补历史空窗。
//   - 原因：采集链路是实时点 + 未来原料投射，历史空窗补不回来也没业务价值；
//     一旦按 PreExecuteTime 逐格追赶，落后越大越慢，最终永远追不上。
//   - 成功后 PreExecuteTime 直接跳到 now，缺口留给实时曲线自然断层；采样侧再按自己的规则处理。
func (job *MonitorDataCollectJob) ExecDataCollect(ctx context.Context, param *xxl.RunReq) string {
	tasks, err := job.repository.MonitorTaskRepository.FindJobBySharding(param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		return "exec failure: load tasks"
	}
	conf, err := job.alertConf.Cover2AlertConf(ctx)
	if err != nil || conf == nil {
		return "exec failure conf"
	}
	collectMaxSecond := conf.CollectMaxSecond
	sampleRawDays := conf.SampleRawDays
	chunkSize := conf.BatchWriteChunkSize
	var wg sync.WaitGroup
	for _, task := range tasks {
		// Push 任务由外部 /api/monitor/task/dataPush/:id 推入，不参与主动采集
		if task.TaskType != nil && *task.TaskType == entity.TaskTypePush {
			continue
		}
		wg.Add(1)
		go job.executeCollect(ctx, task, collectMaxSecond, sampleRawDays, chunkSize, &wg)
	}
	wg.Wait()
	return "execute complete"
}

// DataPush 外部数据推送：校验任务为 Push 类型后，将 items 写入实时序列与样本序列
// 复用 executeCollect 的点构造规则：1 实时点 + 未来 N 天样本点（N = alertConf.sampleRawDays）
func (job *MonitorDataCollectJob) DataPush(ctx context.Context, req *types.MonitorTaskPushDataRequest) error {
	task, err := job.repository.MonitorTaskRepository.GetById(req.TaskId)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	if task.TaskType == nil || *task.TaskType != entity.TaskTypePush {
		return errors.New("当前任务不支持push")
	}
	if len(req.Items) == 0 {
		return errors.New("推送数据不能为空")
	}

	now := time.Now()
	conf, err := job.alertConf.Cover2AlertConf(ctx)
	if err != nil || conf == nil {
		conf = &types.AlertConfObject{SampleRawDays: types.DefaultSampleRawDays, BatchWriteChunkSize: types.DefaultBatchWriteChunkSize}
	}
	sampleRawDays := conf.SampleRawDays
	chunkSize := conf.BatchWriteChunkSize
	capHint := len(req.Items) * (int(sampleRawDays) + 1)
	points := make([]domainHandler.TimeSeriesPoint, 0, capHint)
	var maxTime *time.Time

	for _, item := range req.Items {
		itemTime, err := parseItemTime(item, now)
		if err != nil {
			return err
		}
		if err := validateItemTime(itemTime, now); err != nil {
			return err
		}
		// 记录本轮最大时间，用于回写 PreExecuteTime
		if maxTime == nil || itemTime.After(*maxTime) {
			maxTime = common.Ptr(itemTime)
		}
		// 1 实时点 + N 天未来样本原料
		points = append(points, job.buildCollectPoints(task.TaskKey, item.Value, itemTime, sampleRawDays)...)
	}

	if job.timeSeries != nil {
		if err := job.timeSeries.BatchWrite(ctx, points, int(chunkSize)); err != nil {
			logger.ErrorFormat(ctx, "dataPush timeseries write fail task=%s: %v", task.TaskKey, err)
			return errors.New("保存数据发生错误, 请稍后重试")
		}
	}

	// 回写最大推送时间到 PreExecuteTime
	if maxTime != nil {
		job.updateCollectTime(ctx, task.Id, task.TaskKey, *maxTime)
	}
	return nil
}

// parseItemTime 从推送数据项中解析时间，优先使用 Time，其次 Timestamp，均无则报错。
func parseItemTime(item types.MonitorTaskPushDataItem, now time.Time) (time.Time, error) {
	if item.Time != nil {
		return item.Time.Time, nil
	}
	if item.Timestamp != nil {
		return time.UnixMilli(*item.Timestamp), nil
	}
	return now, errors.New("时间戳或者时间串不能同时为空")
}

// validateItemTime 校验推送数据时间：距今不超过 24h 且不超前。
func validateItemTime(itemTime, now time.Time) error {
	if now.Sub(itemTime) > 24*time.Hour {
		return fmt.Errorf("时间数据[%v]距今已超24小时, 不允许push", itemTime.Format("2006-01-02 15:04:05"))
	}
	if now.Before(itemTime) {
		return fmt.Errorf("时间数据[%v]超前, 不允许push", itemTime.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// buildCollectPoints 构造单个采集点的时序写入点：1 个实时点 + 未来 1~sampleRawDays 天样本原料
// 实时点用于当前值展示；样本原料在 N 天后进入采样窗口，当天无历史时由 dataSampling 用实时序列兜底
func (job *MonitorDataCollectJob) buildCollectPoints(taskKey string, value float64, t time.Time, sampleRawDays int64) []domainHandler.TimeSeriesPoint {
	realtimeMetric := job.resolveRealtimeMetric(taskKey)
	sampleMetric := job.resolveSampleRawMetric(taskKey)
	capacity := int(sampleRawDays) + 1
	points := make([]domainHandler.TimeSeriesPoint, 0, capacity)
	points = append(points, domainHandler.TimeSeriesPoint{Metric: realtimeMetric, Value: value, Timestamp: t})
	for i := int64(1); i <= sampleRawDays; i++ {
		points = append(points, domainHandler.TimeSeriesPoint{
			Metric:    sampleMetric,
			Tags:      map[string]string{"day": fmt.Sprintf("%d", i)},
			Value:     value,
			Timestamp: t.AddDate(0, 0, int(i)),
		})
	}
	return points
}

// resolveRealtimeMetric 解析实时指标名。metricQuery 为 nil 时回退默认 taskKey 本身。
func (job *MonitorDataCollectJob) resolveRealtimeMetric(taskKey string) string {
	if job.metricQuery != nil {
		return job.metricQuery.RealtimeMetric(taskKey)
	}
	return taskKey
}

// resolveSampleRawMetric 解析样本原料指标名。metricQuery 为 nil 时回退默认 _sampling 后缀。
func (job *MonitorDataCollectJob) resolveSampleRawMetric(taskKey string) string {
	if job.metricQuery != nil {
		return job.metricQuery.SampleRawMetric(taskKey)
	}
	return taskKey + "_sampling"
}

// executeCollect 执行单任务采集。
// 始终只采「当前时刻」最近一个窗口（见 renderCommandTemplate），成功后 PreExecuteTime 跳到 now；
// 失败不推进 PreExecuteTime，下一轮仍从当前时刻重试，不回补历史缺口。
func (job *MonitorDataCollectJob) executeCollect(ctx context.Context, task entity.MonitorTask, collectMaxSecond, sampleRawDays, batchWriteChunkSize int64, wg *sync.WaitGroup) {
	defer wg.Done()
	defer recoverLog("dataCollect")

	if task.TaskType == nil {
		return
	}

	// 聚合任务（DataType=Aggregate）：多行分组结果写入时序库，仅收集不做采样/告警
	if task.DataType == entity.DataTypeAggregate {
		job.collectAggregate(ctx, task, collectMaxSecond, batchWriteChunkSize)
		return
	}

	cmd, ok := job.commonMap.GetCommandHandler(ctx, int32(*task.TaskType))
	if !ok {
		logger.WarnFormat(ctx, "no command handler for taskType=%v task=%s", *task.TaskType, task.TaskKey)
		return
	}

	// 任务级超时：单个采集任务超过 collectMaxSecond 秒则 ctx 截断，避免慢 SQL / 慢 URL 拖垮同一批次
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(collectMaxSecond)*time.Second)
	defer cancel()

	// 渲染命令模板（时间锚点 = now，与 PreExecuteTime 无关）
	job.renderCommandTemplate(&task)
	val, err := cmd.ExecuteCommand(taskCtx, task)
	if err != nil {
		errMsg := err.Error()
		if taskCtx.Err() != nil {
			errMsg = fmt.Sprintf("collect timeout (>%ds): %v", collectMaxSecond, taskCtx.Err())
		}
		job.recordCollectError(ctx, task, errors.New(errMsg))
		return
	}

	// 1 实时 + N 天未来原料（不写 day=0，避免当天实时值污染历史基线）
	// 点时间戳用 now：即使 PreExecuteTime 已落后很久，也只落当前点，不补历史空窗
	now := time.Now()
	points := job.buildCollectPoints(task.TaskKey, val, now, sampleRawDays)
	if !job.tryBatchWrite(ctx, task, points, batchWriteChunkSize, "timeseries write fail") {
		return
	}
	// 成功：PreExecuteTime 直接跳到 now，历史缺口不再追
	job.updateCollectTime(ctx, task.Id, task.TaskKey, now)
}

// collectAggregate 聚合任务收集：从数据源取多行分组结果，按 labelColumns/valueColumns 拆解为
// 多个带标签的时序点写入 VM。仅收集，不做采样与告警（DataType=Aggregate）。
func (job *MonitorDataCollectJob) collectAggregate(ctx context.Context, task entity.MonitorTask, collectMaxSecond, batchWriteChunkSize int64) {
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(collectMaxSecond)*time.Second)
	defer cancel()

	var params types.MonitorTaskExecParams
	if task.ExecParams != "" {
		if err := json.Unmarshal([]byte(task.ExecParams), &params); err != nil {
			job.recordCollectError(ctx, task, err)
			return
		}
	}
	if len(params.LabelColumns) == 0 || len(params.ValueColumns) == 0 {
		job.recordCollectError(ctx, task, fmt.Errorf("聚合任务需配置 labelColumns 与 valueColumns"))
		return
	}

	cmd, ok := job.commonMap.GetCommandHandler(ctx, int32(*task.TaskType))
	if !ok {
		job.recordCollectError(ctx, task, fmt.Errorf("no command handler for taskType=%v", *task.TaskType))
		return
	}

	// 渲染命令模板
	job.renderCommandTemplate(&task)
	rows, err := cmd.ExecuteMultiRows(taskCtx, task)
	if err != nil {
		job.recordCollectError(ctx, task, err)
		return
	}

	now := time.Now()
	// 防御：label/value 列名不可为空或仍是聚合函数调用（未取别名会导致 tag/metric 命名错乱）
	if bad := common.FindUnnamedAggColumns(append(append([]string{}, params.LabelColumns...), params.ValueColumns...)); len(bad) > 0 {
		job.recordCollectError(ctx, task, fmt.Errorf("聚合列名不规范：%s 缺少别名，聚合函数/表达式列必须用 AS 取别名（如 COUNT(*) AS cnt）", strings.Join(bad, "、")))
		return
	}

	points := job.buildAggregatePoints(ctx, task, params, rows, now)
	if !job.tryBatchWrite(ctx, task, points, batchWriteChunkSize, "aggregate timeseries write fail") {
		return
	}
	job.updateCollectTime(ctx, task.Id, task.TaskKey, now)
}

// renderCommandTemplate 渲染命令中的时间模板变量（{{.beginTime}} 等），成功后回写 task.Command。
//
// 时间锚点固定为 time.Now()，与 PreExecuteTime 无关：
//   - end   = now
//   - begin = now - TimeSpan（窗口前进间隔）
//   - start = now - StepSpan（查询区间宽度）
//
// 即使任务因故障落后多个周期，也只渲染「当前」窗口，禁止按 PreExecuteTime 回溯补采。
func (job *MonitorDataCollectJob) renderCommandTemplate(task *entity.MonitorTask) {
	end := time.Now()
	begin := end.Add(-time.Duration(task.TimeSpan) * time.Second)
	start := end.Add(-time.Duration(task.StepSpan) * time.Second)
	if rendered, err := renderCommand(task.Command, begin, start, end); err == nil {
		task.Command = rendered
	}
}

// buildAggregatePoints 将聚合多行结果拆解为带标签的时序写入点。
// 每行数据按 labelColumns 生成 tags，每个 valueColumn 生成一个独立 metric 点。
func (job *MonitorDataCollectJob) buildAggregatePoints(ctx context.Context, task entity.MonitorTask, params types.MonitorTaskExecParams, rows []domainHandler.RowResult, now time.Time) []domainHandler.TimeSeriesPoint {
	// 容量上限预估，避免逐行 append 反复扩容
	points := make([]domainHandler.TimeSeriesPoint, 0, len(rows)*len(params.ValueColumns))
	for _, row := range rows {
		points = append(points, job.buildRowPoints(ctx, task.TaskKey, row, params.LabelColumns, params.ValueColumns, now)...)
	}
	return points
}

// buildRowPoints 单行聚合结果拆解为时序点：按 labelColumns 生成 tags，每个 valueColumn 一个 metric 点。
func (job *MonitorDataCollectJob) buildRowPoints(ctx context.Context, taskKey string, row domainHandler.RowResult, labelColumns, valueColumns []string, now time.Time) []domainHandler.TimeSeriesPoint {
	tags := make(map[string]string, len(labelColumns))
	for _, lc := range labelColumns {
		if v, ok := row.Columns[lc]; ok {
			tags[lc] = fmt.Sprint(v)
		}
	}
	points := make([]domainHandler.TimeSeriesPoint, 0, len(valueColumns))
	for _, vc := range valueColumns {
		v, ok := row.Columns[vc]
		if !ok {
			continue
		}
		f, ok := toFloat64Value(v)
		if !ok {
			logger.WarnFormat(ctx, "aggregate value column %s not numeric, skip task=%s", vc, taskKey)
			continue
		}
		points = append(points, domainHandler.TimeSeriesPoint{
			Metric:    taskKey + "_" + vc,
			Tags:      tags,
			Value:     f,
			Timestamp: now,
		})
	}
	return points
}

// recordCollectError 记录采集错误到 CollectErrMsg
func (job *MonitorDataCollectJob) recordCollectError(ctx context.Context, task entity.MonitorTask, err error) {
	logger.ErrorFormat(ctx, "collect fail task=%s: %v", task.TaskKey, err)
	if uErr := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: err.Error()}); uErr != nil {
		logger.ErrorFormat(ctx, "record collect error fail task=%s: collectErr=%v updateErr=%v", task.TaskKey, err, uErr)
	}
}

// tryBatchWrite 时序写入与错误记录，executeCollect 和 collectAggregate 共用。
// 写入成功返回 true，失败返回 false（调用方跳过 updateCollectTime 以保留重试机会）。
func (job *MonitorDataCollectJob) tryBatchWrite(ctx context.Context, task entity.MonitorTask, points []domainHandler.TimeSeriesPoint, chunkSize int64, errMsg string) bool {
	if len(points) == 0 || job.timeSeries == nil {
		return true
	}
	if err := job.timeSeries.BatchWrite(ctx, points, int(chunkSize)); err != nil {
		job.recordCollectError(ctx, task, fmt.Errorf("%s: %w", errMsg, err))
		return false
	}
	return true
}

// updateCollectTime 回写 PreExecuteTime 并清空 CollectErrMsg。
func (job *MonitorDataCollectJob) updateCollectTime(ctx context.Context, taskId int64, taskKey string, t time.Time) {
	if err := job.repository.MonitorTaskRepository.UpdateById(taskId, &entity.MonitorTask{
		PreExecuteTime: &common.LocalTime{Time: t},
		CollectErrMsg:  " ",
	}); err != nil {
		// 采集时间未落库，下一轮调度可能重复采集/重复写时序，必须留痕
		logger.ErrorFormat(ctx, "update PreExecuteTime fail task=%s: %v", taskKey, err)
	}
}

// toFloat64Value 将任意数值类型转为 float64
func toFloat64Value(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
