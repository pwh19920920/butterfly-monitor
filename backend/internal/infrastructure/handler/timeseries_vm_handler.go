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
	MetricNamer
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
				b.WriteString(sanitizePromLabelName(k))
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
	return batchWriteChunked(ctx, points, chunkSize, h.WritePoints)
}

// QueryRangeWithTags 查询区间内的多 series 数据，并按 tagFilters 过滤标签。
// tagFilters 为空(map长度为0)表示不过滤，返回该 metric 下所有 series。
func (h *TimeSeriesVmHandler) QueryRangeWithTags(ctx context.Context, metric string, start, end time.Time, tagFilters map[string][]string) ([]domainHandler.SeriesData, error) {
	raw, err := h.queryRange(ctx, h.buildTagFilteredExpr(metric, tagFilters), start, end)
	if err != nil {
		return nil, err
	}
	series := make([]domainHandler.SeriesData, 0, len(raw))
	for _, r := range raw {
		series = append(series, domainHandler.SeriesData{Labels: r.Metric, Values: r.Values})
	}
	return series, nil
}

// QueryRangeValues 查询区间内的原始 value 列表（所有 series 的值展平）
func (h *TimeSeriesVmHandler) QueryRangeValues(ctx context.Context, metric string, start, end time.Time) ([]float64, error) {
	raw, err := h.queryRange(ctx, escapePromMetric(metric), start, end)
	if err != nil {
		return nil, err
	}
	vals := make([]float64, 0)
	for _, r := range raw {
		vals = append(vals, r.Values...)
	}
	return vals, nil
}

// vmSeries queryRange 返回的单条 series
type vmSeries struct {
	Metric map[string]string
	Values []float64
}

// queryRange 公共 query_range 调用：构建 URL → GET → 解析 JSON → 提取各 series 的 float64 值。
// QueryRangeWithTags 和 QueryRangeValues 共用此方法，仅 query 表达式不同。
func (h *TimeSeriesVmHandler) queryRange(ctx context.Context, query string, start, end time.Time) ([]vmSeries, error) {
	u, err := url.Parse(strings.TrimRight(h.addr, "/") + "/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", query)
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
		return nil, fmt.Errorf("vm query_range status=%d body=%s", resp.StatusCode, truncateBody(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	series := make([]vmSeries, 0, len(result.Data.Result))
	for _, r := range result.Data.Result {
		vals := make([]float64, 0, len(r.Values))
		for _, pair := range r.Values {
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
		series = append(series, vmSeries{Metric: r.Metric, Values: vals})
	}
	return series, nil
}

// buildTagFilteredExpr 构造带标签过滤器的 PromQL 表达式。
// 空过滤器返回裸 metric 名；单值用精确匹配 k="v"，多值用正则 k=~"v1|v2"。
// metric 名与 tag key 均经清洗，防止用户可控值注入 PromQL 选择器。
// 多值分支额外对每个值转义正则元字符，防止 . * + 等被当作正则运算符。
func (h *TimeSeriesVmHandler) buildTagFilteredExpr(metric string, tagFilters map[string][]string) string {
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

// QueryMean 查询区间均值 avg(metric)
func (h *TimeSeriesVmHandler) QueryMean(ctx context.Context, metric string, start, end time.Time) (*float64, error) {
	query := fmt.Sprintf("avg(%s)", escapePromMetric(metric))
	return h.queryInstant(ctx, query, end)
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
		return nil, fmt.Errorf("vm query status=%d body=%s", resp.StatusCode, truncateBody(body))
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
		return fmt.Errorf("vm write status=%d body=%s", resp.StatusCode, truncateBody(respBody))
	}
	return nil
}

func (h *TimeSeriesVmHandler) applyAuth(req *http.Request) {
	if h.username != "" {
		req.SetBasicAuth(h.username, h.password)
	}
}

// ---------------- MetricQueryDialect ----------------

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

// truncateBody 截断错误消息中的响应体，避免完整 body 泄露内部信息或撑大日志（D-017）。
func truncateBody(body []byte) string {
	const maxLen = 500
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "...(truncated)"
}

// escapePromMetric 指标名清洗：非法字符替换为下划线，数字开头加 m_ 前缀。
// Prometheus metric 名须匹配 [a-zA-Z_:][a-zA-Z0-9_:]*，用户可控的 taskKey 可能含
// 特殊字符或以数字开头，直接拼入 PromQL/文本格式会导致注入或写入被拒，故统一清洗。
func escapePromMetric(name string) string {
	if name == "" {
		return name
	}
	var b strings.Builder
	for _, r := range name {
		if r == '_' || r == ':' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	s := b.String()
	if s == "" {
		return "m"
	}
	// 首字符为数字时 Prometheus 不接受，加前缀
	if s[0] >= '0' && s[0] <= '9' {
		s = "m_" + s
	}
	return s
}

// sanitizePromLabelName 标签名清洗：与 metric 名同规则，防止用户可控的 tag key
// （如下钻 FieldName、聚合 labelColumn）注入 PromQL 选择器。
func sanitizePromLabelName(name string) string {
	return escapePromMetric(name)
}

// escapePromLabelValue 标签值转义：\ " \n \r
func escapePromLabelValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	v = strings.ReplaceAll(v, "\r", `\r`)
	return v
}

// escapePromLabelValueForRegex 标签值转义后额外转义正则元字符，用于多值 =~ 分支。
// 在 escapePromLabelValue 基础上把 . * + ? | ( ) { } [ ] ^ $ 等正则特殊字符用 \ 转义，
// 避免值中的正则元字符被当作运算符匹配到非预期标签值。
func escapePromLabelValueForRegex(v string) string {
	v = escapePromLabelValue(v)
	// 按长度从高到低排序，优先转义多字符序列
	for _, ch := range []string{".", "*", "+", "?", "|", "(", ")", "{", "}", "[", "]", "^", "$"} {
		v = strings.ReplaceAll(v, ch, `\`+ch)
	}
	return v
}
