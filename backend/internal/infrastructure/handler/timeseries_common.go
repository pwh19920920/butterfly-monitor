package handler

import (
	"context"
	"time"

	domainHandler "butterfly-monitor/internal/domain/handler"
)

// MetricNamer 指标命名：实时/样本原料/平滑基线。
// VM 和 TDengine 共用同一套命名规则，嵌入各自 handler 以消除重复。
type MetricNamer struct{}

func (MetricNamer) RealtimeMetric(taskKey string) string {
	return taskKey
}

func (MetricNamer) SampleRawMetric(taskKey string) string {
	return taskKey + "_sampling"
}

func (MetricNamer) SmoothMetric(taskKey string) string {
	return taskKey + "_sample"
}

// batchWriteChunked 通用分片写入：按 chunkSize 切片调用 writeFn，片间 sleep 200ms 避免打满后端。
// VM 和 TDengine 的 BatchWrite 逻辑完全一致，仅 WritePoints 实现不同，故抽为公共函数。
// 片间等待响应 ctx 取消（D-027），避免 ctx 已取消时仍阻塞 200ms。
func batchWriteChunked(ctx context.Context, points []domainHandler.TimeSeriesPoint, chunkSize int, writeFn func(context.Context, []domainHandler.TimeSeriesPoint) error) error {
	if chunkSize <= 0 {
		chunkSize = 3000
	}
	for i := 0; i < len(points); i += chunkSize {
		end := i + chunkSize
		if end > len(points) {
			end = len(points)
		}
		if err := writeFn(ctx, points[i:end]); err != nil {
			return err
		}
		if end < len(points) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
	}
	return nil
}
