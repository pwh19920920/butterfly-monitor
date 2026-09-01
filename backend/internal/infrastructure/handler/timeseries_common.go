package handler

import (
	"context"
	"fmt"
	"strings"
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
		end := min(i+chunkSize, len(points))
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

// buildPromTagFilteredExpr 构造带标签过滤器的 PromQL 表达式（VM / remote_write 共享）。
//
// 语义：
//   - 空过滤器返回裸 metric 名；
//   - 单值用精确匹配 k="v"；
//   - 多值用正则 k=~"v1|v2"。
//
// 安全：metric 名与 tag key 经 escapePromMetric / sanitizePromLabelName 清洗，
// tag value 走 escapePromLabelValue 防注入；多值分支额外对每个值转义正则元字符，
// 防止 . * + 等被当作正则运算符。
func buildPromTagFilteredExpr(metric string, tagFilters map[string][]string) string {
	safeMetric := escapePromMetric(metric)
	if len(tagFilters) == 0 {
		return safeMetric
	}
	var b strings.Builder
	b.WriteString(safeMetric)
	b.WriteByte('{')
	first := true
	for k, vals := range tagFilters {
		if len(vals) == 0 {
			continue
		}
		safeKey := sanitizePromLabelName(k)
		if safeKey == "" {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		if len(vals) == 1 {
			b.WriteString(fmt.Sprintf("%s=\"%s\"", safeKey, escapePromLabelValue(fmt.Sprint(vals[0]))))
		} else {
			escaped := make([]string, 0, len(vals))
			for _, v := range vals {
				escaped = append(escaped, escapePromLabelValueForRegex(fmt.Sprint(v)))
			}
			b.WriteString(fmt.Sprintf("%s=~\"%s\"", safeKey, strings.Join(escaped, "|")))
		}
	}
	b.WriteByte('}')
	return b.String()
}
