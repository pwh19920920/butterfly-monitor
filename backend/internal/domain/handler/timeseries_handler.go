package handler

import (
	"context"
	"time"
)

// TimeSeriesPoint 与后端无关的写入点
type TimeSeriesPoint struct {
	Metric    string
	Tags      map[string]string
	Value     float64
	Timestamp time.Time
}

// SeriesData 带标签的时间序列查询结果：一个 series 对应一组标签 + 一组按时间排列的值
type SeriesData struct {
	Labels map[string]string // 该 series 的标签集合，如 {"region":"华东","dept":"技术部"}
	Values []float64         // 按时间顺序排列的值
}

// TimeSeriesStore 时序读写抽象（VictoriaMetrics / 将来 Influx 等）
type TimeSeriesStore interface {
	// WritePoints 批量写入
	WritePoints(ctx context.Context, points []TimeSeriesPoint) error
	// BatchWrite 分片写入
	BatchWrite(ctx context.Context, points []TimeSeriesPoint, chunkSize int) error
	// QueryMean 查询区间均值
	QueryMean(ctx context.Context, metric string, start, end time.Time) (*float64, error)
	// QueryRangeValues 查询区间内原始 value 列表
	QueryRangeValues(ctx context.Context, metric string, start, end time.Time) ([]float64, error)
	// QueryRangeWithTags 查询区间内的多 series 数据，并按 tagFilters 过滤标签。
	// tagFilters 为空(map[nil])表示不过滤，返回该 metric 下所有 series。
	QueryRangeWithTags(ctx context.Context, metric string, start, end time.Time, tagFilters map[string][]string) ([]SeriesData, error)
}

// MetricQueryDialect 指标命名 + Grafana 查询方言
// 与 TimeSeriesStore 通常由同一实现类型承担，保证读写与面板一致
type MetricQueryDialect interface {
	// RealtimeMetric 实时指标名
	RealtimeMetric(taskKey string) string
	// SampleRawMetric 样本原料指标名（*_sampling）
	SampleRawMetric(taskKey string) string
	// SmoothMetric 平滑基线指标名（*_sample）
	SmoothMetric(taskKey string) string

	// DatasourceType Grafana 数据源类型（prometheus / influxdb …）
	DatasourceType() string
	// RealtimeExpr 实时曲线查询表达式
	RealtimeExpr(taskKey string) string
	// SmoothExpr 样本曲线查询表达式
	SmoothExpr(taskKey string) string
	// BuildPanelTarget 构建 Grafana panel target
	BuildPanelTarget(refID, legend, expr string) map[string]interface{}
}
