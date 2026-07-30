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

	"dragonfly-monitor/internal/application"
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

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

func (job *MonitorDataCollectJob) ExecDataCollect(ctx context.Context, param *xxl.RunReq) string {
	tasks, err := job.repository.MonitorTaskRepository.FindJobBySharding(param.BroadcastIndex, param.BroadcastTotal)
	if err != nil {
		return "exec failure: load tasks"
	}
	// 单任务最大执行秒数：超过则 ctx 截断，防止拖垮下一个采集批次
	collectMaxSecond := int64(25)
	if conf, cErr := job.alertConf.Cover2AlertConf(ctx); cErr == nil && conf.CollectMaxSecond > 0 {
		collectMaxSecond = conf.CollectMaxSecond
	}
	var wg sync.WaitGroup
	for _, task := range tasks {
		// Push 任务由外部 /api/monitor/task/dataPush/:id 推入，不参与主动采集
		if task.TaskType != nil && *task.TaskType == entity.TaskTypePush {
			continue
		}
		wg.Add(1)
		go job.executeCollect(ctx, task, collectMaxSecond, &wg)
	}
	wg.Wait()
	return "execute complete"
}

// DataPush 外部数据推送：校验任务为 Push 类型后，将 items 写入实时序列与样本序列
// 复用 executeCollect 的点构造规则：1 实时点 + 未来 8 天样本点
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
	points := make([]domainHandler.TimeSeriesPoint, 0, len(req.Items)*9)
	var maxTime *time.Time

	for _, item := range req.Items {
		if item.Time == nil && item.Timestamp == nil {
			return errors.New("时间戳或者时间串不能同时为空")
		}
		itemTime := now
		if item.Time != nil {
			itemTime = item.Time.Time
		} else {
			itemTime = time.UnixMilli(*item.Timestamp)
		}
		// 距今超过 24 小时拒绝
		if diff := now.Sub(itemTime); int(diff.Hours()) > 24 {
			return fmt.Errorf("时间数据[%v]距今已超24小时, 不允许push", itemTime.Format("2006-01-02 15:04:05"))
		}
		// 未来时间拒绝
		if now.Before(itemTime) {
			return fmt.Errorf("时间数据[%v]超前, 不允许push", itemTime.Format("2006-01-02 15:04:05"))
		}
		// 记录本轮最大时间，用于回写 PreExecuteTime
		if maxTime == nil || itemTime.After(*maxTime) {
			maxTime = common.Ptr(itemTime)
		}
		// 1 实时点 + 8 天未来样本原料（复用 executeCollect 的点构造规则）
		points = append(points, job.buildCollectPoints(task.TaskKey, item.Value, itemTime)...)
	}

	if job.timeSeries != nil {
		if err := job.timeSeries.BatchWrite(ctx, points, 3000); err != nil {
			logger.ErrorFormat(ctx, "dataPush timeseries write fail task=%s: %v", task.TaskKey, err)
			return errors.New("保存数据发生错误, 请稍后重试")
		}
	}

	// 回写最大推送时间到 PreExecuteTime
	if maxTime != nil {
		if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
			PreExecuteTime: &common.LocalTime{Time: *maxTime},
			CollectErrMsg:  " ",
		}); err != nil {
			logger.ErrorFormat(ctx, "dataPush update PreExecuteTime fail task=%s: %v", task.TaskKey, err)
			return err
		}
	}
	return nil
}

// buildCollectPoints 构造单个采集点的时序写入点：1 个实时点 + 未来 1~8 天样本原料
// 实时点用于当前值展示；样本原料在 N 天后进入采样窗口，当天无历史时由 dataSampling 用实时序列兜底
func (job *MonitorDataCollectJob) buildCollectPoints(taskKey string, value float64, t time.Time) []domainHandler.TimeSeriesPoint {
	realtimeMetric := taskKey
	if job.metricQuery != nil {
		realtimeMetric = job.metricQuery.RealtimeMetric(taskKey)
	}
	sampleMetric := taskKey + "_sampling"
	if job.metricQuery != nil {
		sampleMetric = job.metricQuery.SampleRawMetric(taskKey)
	}
	points := make([]domainHandler.TimeSeriesPoint, 0, 9)
	points = append(points, domainHandler.TimeSeriesPoint{Metric: realtimeMetric, Value: value, Timestamp: t})
	for i := 1; i <= 8; i++ {
		points = append(points, domainHandler.TimeSeriesPoint{
			Metric:    sampleMetric,
			Tags:      map[string]string{"day": fmt.Sprintf("%d", i)},
			Value:     value,
			Timestamp: t.AddDate(0, 0, i),
		})
	}
	return points
}

func (job *MonitorDataCollectJob) executeCollect(ctx context.Context, task entity.MonitorTask, collectMaxSecond int64, wg *sync.WaitGroup) {
	defer wg.Done()
	defer recoverLog("dataCollect")

	if task.TaskType == nil {
		return
	}

	// 聚合任务（DataType=Aggregate）：多行分组结果写入时序库，仅收集不做采样/告警
	if task.DataType == entity.DataTypeAggregate {
		job.collectAggregate(ctx, task, collectMaxSecond)
		return
	}

	cmd, ok := job.commonMap.GetCommandHandler(ctx, int32(*task.TaskType))
	if !ok {
		logger.WarnFormat(ctx, "no command handler for taskType=%v task=%s", *task.TaskType, task.TaskKey)
		return
	}
	// 任务级超时：单个采集任务超过 collectMaxSecond 秒则 ctx 截断，
	// 避免慢 SQL / 慢 URL 拖垮同一批次后续任务。ctx 透传到 driver 层。
	if collectMaxSecond <= 0 {
		collectMaxSecond = 25
	}
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(collectMaxSecond)*time.Second)
	defer cancel()
	// 渲染模板
	end := time.Now()
	begin := end.Add(-time.Duration(task.TimeSpan) * time.Second)
	start := end.Add(-time.Duration(task.StepSpan) * time.Second)
	rendered, err := renderCommand(task.Command, begin, start, end)
	if err == nil {
		task.Command = rendered
	}
	val, err := cmd.ExecuteCommand(taskCtx, task)
	if err != nil {
		// 超时单独标记，便于排查「追不上/拖批次」类问题
		errMsg := err.Error()
		if taskCtx.Err() != nil {
			errMsg = fmt.Sprintf("collect timeout (>%ds): %v", collectMaxSecond, taskCtx.Err())
		}
		if uErr := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: errMsg}); uErr != nil {
			logger.ErrorFormat(ctx, "record collect error fail task=%s: collectErr=%v updateErr=%v", task.TaskKey, err, uErr)
		}
		return
	}
	// 1 实时 + 8 天未来原料（不写 day=0，避免当天实时值污染历史基线）
	now := time.Now()
	points := job.buildCollectPoints(task.TaskKey, val, now)
	if job.timeSeries != nil {
		if err := job.timeSeries.BatchWrite(ctx, points, 3000); err != nil {
			logger.ErrorFormat(ctx, "timeseries write fail task=%s: %v", task.TaskKey, err)
		}
	}
	if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		PreExecuteTime: &common.LocalTime{Time: now},
		CollectErrMsg:  " ",
	}); err != nil {
		// 采集时间未落库，下一轮调度可能重复采集/重复写时序，必须留痕
		logger.ErrorFormat(ctx, "update PreExecuteTime fail task=%s: %v", task.TaskKey, err)
	}
}

// collectAggregate 聚合任务收集：从数据源取多行分组结果，按 labelColumns/valueColumns 拆解为
// 多个带标签的时序点写入 VM。仅收集，不做采样与告警（DataType=Aggregate）。
// 聚合任务复用现有 CommandHandler.ExecuteMultiRows，数据源类型由 TaskType 决定（DB/URL/Push）。
func (job *MonitorDataCollectJob) collectAggregate(ctx context.Context, task entity.MonitorTask, collectMaxSecond int64) {
	if collectMaxSecond <= 0 {
		collectMaxSecond = 25
	}
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

	// 渲染命令模板（DB/URL 的 Command 可能含时间变量）
	end := time.Now()
	begin := end.Add(-time.Duration(task.TimeSpan) * time.Second)
	start := end.Add(-time.Duration(task.StepSpan) * time.Second)
	if rendered, err := renderCommand(task.Command, begin, start, end); err == nil {
		task.Command = rendered
	}

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
	points := make([]domainHandler.TimeSeriesPoint, 0)
	for _, row := range rows {
		tags := make(map[string]string, len(params.LabelColumns))
		for _, lc := range params.LabelColumns {
			if v, ok := row.Columns[lc]; ok {
				tags[lc] = fmt.Sprint(v)
			}
		}
		for _, vc := range params.ValueColumns {
			v, ok := row.Columns[vc]
			if !ok {
				continue
			}
			f, ok := toFloat64Value(v)
			if !ok {
				logger.WarnFormat(ctx, "aggregate value column %s not numeric, skip task=%s", vc, task.TaskKey)
				continue
			}
			points = append(points, domainHandler.TimeSeriesPoint{
				Metric:    task.TaskKey + "_" + vc,
				Tags:      tags,
				Value:     f,
				Timestamp: now,
			})
		}
	}

	if len(points) > 0 && job.timeSeries != nil {
		if err := job.timeSeries.BatchWrite(ctx, points, 3000); err != nil {
			logger.ErrorFormat(ctx, "aggregate timeseries write fail task=%s: %v", task.TaskKey, err)
		}
	}
	if err := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{
		PreExecuteTime: &common.LocalTime{Time: now},
		CollectErrMsg:  " ",
	}); err != nil {
		logger.ErrorFormat(ctx, "update PreExecuteTime fail aggregate task=%s: %v", task.TaskKey, err)
	}
}

// recordCollectError 记录采集错误到 CollectErrMsg
func (job *MonitorDataCollectJob) recordCollectError(ctx context.Context, task entity.MonitorTask, err error) {
	logger.ErrorFormat(ctx, "aggregate collect fail task=%s: %v", task.TaskKey, err)
	if uErr := job.repository.MonitorTaskRepository.UpdateById(task.Id, &entity.MonitorTask{CollectErrMsg: err.Error()}); uErr != nil {
		logger.ErrorFormat(ctx, "record aggregate collect error fail task=%s: collectErr=%v updateErr=%v", task.TaskKey, err, uErr)
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
