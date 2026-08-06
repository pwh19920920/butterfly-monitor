package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"
)

// DatabaseTimeseriesHandler 系统下钻任务的数据源：从时序库(VM)按标签过滤取数。
// 下钻任务以它为 CommandHandler 注册到 commonMap（TaskTypeDrilldown），
// 因此 executeCollect 复用现有 cmd.ExecuteCommand → buildCollectPoints 主流程，无需特殊分支。
//
// 取数契约：
//   - 按 sourceTaskId 找到聚合任务，拼出源 metric = sourceTaskKey + "_" + queryMetric
//   - tagFilters 由 filters 全部等值精确匹配构成（维度数量需与聚合分组层数一致）
//   - 查询结果必须恰好 1 个 series，否则视为配置错误：
//     0 个 series → 返回 DefaultValue（点位不存在，前端配 0/100，默认 0）
//     >1 个 series → 报错（过滤条件未覆盖全部分组维度）
//   - 恰好 1 个 series → 取该 series 最新值返回 float64
type DatabaseTimeseriesHandler struct {
	// TimeSeries 时序库（VM），注入以避免 handler 反向依赖 config。
	TimeSeries domainHandler.TimeSeriesStore
	// GetSourceTask 按聚合任务 ID 取任务（用于反查 taskKey）。注入以避免 handler 反向依赖 repository。
	GetSourceTask func(id int64) (*entity.MonitorTask, error)
}

func (h *DatabaseTimeseriesHandler) ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error) {
	if h.TimeSeries == nil {
		return 0, fmt.Errorf("下钻任务执行失败：时序库未配置")
	}
	var params struct {
		SourceTaskId *int64   `json:"sourceTaskId,string"`
		QueryMetric  string   `json:"queryMetric"`
		DefaultValue *float64 `json:"defaultValue"`
		Filters      []struct {
			FieldName string `json:"fieldName"`
			Value     string `json:"value"`
		} `json:"filters"`
	}
	if task.ExecParams == "" {
		return 0, fmt.Errorf("下钻任务 execParams 为空")
	}
	if err := json.Unmarshal([]byte(task.ExecParams), &params); err != nil {
		return 0, fmt.Errorf("解析下钻 execParams 失败: %w", err)
	}
	if params.SourceTaskId == nil {
		return 0, fmt.Errorf("下钻任务 sourceTaskId 为空")
	}
	if h.GetSourceTask == nil {
		return 0, fmt.Errorf("下钻 handler 未配置 GetSourceTask")
	}

	sourceTask, err := h.GetSourceTask(*params.SourceTaskId)
	if err != nil || sourceTask == nil {
		return 0, fmt.Errorf("下钻依赖的聚合任务 %d 不存在", *params.SourceTaskId)
	}

	// 拼源 metric：聚合任务写入的 metric 为 taskKey + "_" + valueColumn
	queryMetric := params.QueryMetric
	if queryMetric == "" {
		// 未指定则尝试取聚合任务的第一个 valueColumn
		var srcParams struct {
			ValueColumns []string `json:"valueColumns"`
		}
		if sourceTask.ExecParams != "" {
			_ = json.Unmarshal([]byte(sourceTask.ExecParams), &srcParams)
		}
		if len(srcParams.ValueColumns) == 0 {
			return 0, fmt.Errorf("聚合任务 %d 未配置 valueColumns，下钻需指定 queryMetric", *params.SourceTaskId)
		}
		queryMetric = srcParams.ValueColumns[0]
	}
	metric := sourceTask.TaskKey + "_" + queryMetric

	// 构建标签过滤器：每个维度等值精确匹配（单值）
	tagFilters := make(map[string][]string, len(params.Filters))
	for _, f := range params.Filters {
		if f.FieldName == "" || f.Value == "" {
			continue
		}
		tagFilters[f.FieldName] = []string{f.Value}
	}

	spanSec := int64(task.TimeSpan)
	if spanSec <= 0 {
		spanSec = 60
	}
	// 时间窗：[now-2*TS, now-TS]，避免与聚合并发时查空
	now := time.Now()
	start := now.Add(-time.Duration(2*spanSec) * time.Second)
	end := now.Add(-time.Duration(spanSec) * time.Second)

	series, err := h.TimeSeries.QueryRangeWithTags(ctx, metric, start, end, tagFilters)
	if err != nil {
		return 0, fmt.Errorf("下钻查询 VM 失败: %w", err)
	}

	switch len(series) {
	case 0:
		// 点位不存在：返回默认值（前端配 0/100），不报错
		return *params.DefaultValue, nil
	case 1:
		vals := series[0].Values
		if len(vals) == 0 {
			return *params.DefaultValue, nil
		}
		return vals[len(vals)-1], nil
	default:
		return 0, fmt.Errorf("下钻过滤条件未覆盖全部分组维度，命中 %d 个 series（需恰好 1 个）", len(series))
	}
}

// ExecuteMultiRows 下钻 handler 仅取单 float64，不提供多行聚合能力
func (h *DatabaseTimeseriesHandler) ExecuteMultiRows(ctx context.Context, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	return nil, fmt.Errorf("下钻任务不支持多行取数")
}

// TestConnect 下钻数据源为时序库，无需连接测试占位
func (h *DatabaseTimeseriesHandler) TestConnect(database entity.MonitorDatabase) error {
	return nil
}

// NewInstance 下钻数据源为时序库，无需创建连接实例
func (h *DatabaseTimeseriesHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	return nil, nil
}

// Close 下钻数据源为时序库，无需关闭
func (h *DatabaseTimeseriesHandler) Close(db interface{}) error {
	return nil
}
