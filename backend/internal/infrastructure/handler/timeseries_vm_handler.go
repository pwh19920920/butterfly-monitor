package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/config/victoriametrics"
	domainHandler "dragonfly-monitor/internal/domain/handler"
)

// TimeSeriesVmHandler VictoriaMetrics 实现：TimeSeriesStore + MetricQueryDialect
type TimeSeriesVmHandler struct {
	addr     string
	username string
	password string
	client   *http.Client
}

// NewTimeSeriesVmHandler 根据配置创建 VM 时序客户端
func NewTimeSeriesVmHandler(cfg victoriametrics.Config) *TimeSeriesVmHandler {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		addr = "http://127.0.0.1:8428"
	}
	return &TimeSeriesVmHandler{
		addr:     addr,
		username: cfg.Username,
		password: cfg.Password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ---------------- TimeSeriesStore ----------------

// WritePoints 批量写入（Prometheus 文本格式 → /api/v1/import/prometheus）
// 注意：不要用 Influx 行协议 /write。Influx 会把字段名拼进指标名，
// 例如 measurement=test field=value → 存成 test_value，导致 Grafana 的 avg(test) 查不到。
func (h *TimeSeriesVmHandler) WritePoints(ctx context.Context, points []domainHandler.TimeSeriesPoint) error {
	if len(points) == 0 {
		return nil
	}
	var b strings.Builder
	for _, p := range points {
		// Prometheus 文本：metric{tag="v"} value timestamp_ms
		b.WriteString(escapePromMetric(p.Metric))
		if len(p.Tags) > 0 {
			b.WriteByte('{')
			first := true
			for k, v := range p.Tags {
				if !first {
					b.WriteByte(',')
				}
				first = false
				b.WriteString(k)
				b.WriteByte('=')
				b.WriteByte('"')
				b.WriteString(escapePromLabelValue(v))
				b.WriteByte('"')
			}
			b.WriteByte('}')
		}
		b.WriteByte(' ')
		b.WriteString(fmt.Sprintf("%v", p.Value))
		b.WriteByte(' ')
		b.WriteString(fmt.Sprintf("%d", p.Timestamp.UnixMilli()))
		b.WriteByte('\n')
	}
	return h.post(ctx, "/api/v1/import/prometheus", "text/plain", []byte(b.String()))
}

// BatchWrite 分片写入
func (h *TimeSeriesVmHandler) BatchWrite(ctx context.Context, points []domainHandler.TimeSeriesPoint, chunkSize int) error {
	if chunkSize <= 0 {
		chunkSize = 3000
	}
	for i := 0; i < len(points); i += chunkSize {
		end := i + chunkSize
		if end > len(points) {
			end = len(points)
		}
		if err := h.WritePoints(ctx, points[i:end]); err != nil {
			return err
		}
		if end < len(points) {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return nil
}

// QueryMean 查询区间均值 avg(metric)
func (h *TimeSeriesVmHandler) QueryMean(ctx context.Context, metric string, start, end time.Time) (*float64, error) {
	query := fmt.Sprintf("avg(%s)", metric)
	return h.queryInstant(ctx, query, end)
}

// QueryRangeValues 查询区间内的原始 value 列表
func (h *TimeSeriesVmHandler) QueryRangeValues(ctx context.Context, metric string, start, end time.Time) ([]float64, error) {
	u, err := url.Parse(strings.TrimRight(h.addr, "/") + "/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", metric)
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	q.Set("step", "15s")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	h.applyAuth(req)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vm query_range status=%d body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Values [][]interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	vals := make([]float64, 0)
	for _, series := range result.Data.Result {
		for _, pair := range series.Values {
			if len(pair) < 2 {
				continue
			}
			s, ok := pair[1].(string)
			if !ok {
				continue
			}
			var f float64
			if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
				vals = append(vals, f)
			}
		}
	}
	return vals, nil
}

func (h *TimeSeriesVmHandler) queryInstant(ctx context.Context, query string, ts time.Time) (*float64, error) {
	u, err := url.Parse(strings.TrimRight(h.addr, "/") + "/api/v1/query")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("time", fmt.Sprintf("%d", ts.Unix()))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	h.applyAuth(req)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vm query status=%d body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return nil, nil
	}
	s, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return nil, nil
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return nil, err
	}
	return common.Ptr(f), nil
}

func (h *TimeSeriesVmHandler) post(ctx context.Context, path string, contentType string, body []byte) error {
	u := strings.TrimRight(h.addr, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	h.applyAuth(req)
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("vm write status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (h *TimeSeriesVmHandler) applyAuth(req *http.Request) {
	if h.username != "" {
		req.SetBasicAuth(h.username, h.password)
	}
}

// ---------------- MetricQueryDialect ----------------

// RealtimeMetric 实时指标名
func (h *TimeSeriesVmHandler) RealtimeMetric(taskKey string) string {
	return taskKey
}

// SampleRawMetric 样本原料指标名
func (h *TimeSeriesVmHandler) SampleRawMetric(taskKey string) string {
	return fmt.Sprintf("%s_sampling", taskKey)
}

// SmoothMetric 平滑基线指标名
func (h *TimeSeriesVmHandler) SmoothMetric(taskKey string) string {
	return fmt.Sprintf("%s_sample", taskKey)
}

// DatasourceType Grafana Prometheus 数据源
func (h *TimeSeriesVmHandler) DatasourceType() string {
	return "prometheus"
}

// RealtimeExpr PromQL 实时曲线（$interval 来自大盘分组跨度变量）
func (h *TimeSeriesVmHandler) RealtimeExpr(taskKey string) string {
	// avg_over_time 按 $interval 聚合，配合 dashboard templating.interval
	return fmt.Sprintf("avg(avg_over_time(%s[$interval]))", escapePromMetric(h.RealtimeMetric(taskKey)))
}

// SmoothExpr PromQL 样本曲线
func (h *TimeSeriesVmHandler) SmoothExpr(taskKey string) string {
	return fmt.Sprintf("avg(avg_over_time(%s[$interval]))", escapePromMetric(h.SmoothMetric(taskKey)))
}

// BuildPanelTarget Prometheus plugin target
func (h *TimeSeriesVmHandler) BuildPanelTarget(refID, legend, expr string) map[string]interface{} {
	return map[string]interface{}{
		"refId":        refID,
		"expr":         expr,
		"legendFormat": legend,
		"editorMode":   "code",
		"range":        true,
		"datasource": map[string]string{
			"type": h.DatasourceType(),
			"uid":  "${datasource}",
		},
	}
}

// escapePromMetric 指标名：合法字符原样返回；含特殊字符时用 {} 选择器兜底（少见）
// PromQL 选择器写法：{"metric-name"} 不通用，这里尽量只允许 [a-zA-Z_:][a-zA-Z0-9_:]*
func escapePromMetric(name string) string {
	if name == "" {
		return name
	}
	for _, c := range name {
		if !(c == '_' || c == ':' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			// 非法字符替换为下划线，保证写入/查询同一套名字
			var b strings.Builder
			for _, r := range name {
				if r == '_' || r == ':' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					b.WriteRune(r)
				} else {
					b.WriteByte('_')
				}
			}
			return b.String()
		}
	}
	return name
}

// escapePromLabelValue 标签值转义：\ " \n
func escapePromLabelValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	return v
}
