package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"butterfly-monitor/internal/config/promremotewrite"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

const (
	// remote_write 重试参数
	rwInitialBackoff = 1 * time.Second
	rwMaxBackoff     = 30 * time.Second
	rwBackoffMul     = 2

	// HTTP 响应体读取上限
	rwMaxErrorBodyBytes = 1024    // 错误响应体最多读 1KB
	rwMaxResponseBytes  = 4 << 20 // query 响应体最多读 4MB
	rwMinStep           = 15 * time.Second
	rwAdaptiveMaxPoints = 500

	// remote_write 协议常量
	rwContentTypeHeader   = "application/x-protobuf"
	rwContentEncoding     = "snappy"
	rwUserAgent           = "butterfly-monitor/1.0"
	rwRemoteWriteVersion  = "0.1.0"
	rwPrometheusNameLabel = "__name__"
)

// TimeSeriesPromRemoteWriteHandler 通过 Prometheus remote_write 协议将指标推送到远端
// 兼容端点（Prometheus 自身不接收 remote_write，需搭配 vmagent / Mimir / Thanos Receive /
// Grafana Cloud / VictoriaMetrics remote_write receiver / OpenObserve 等前置接收组件）。
//
// 协议要点（写路径）：
//   - HTTP POST {URL}
//   - Content-Type: application/x-protobuf
//   - Content-Encoding: snappy
//   - Body: snappy-compressed prompb.WriteRequest（包含若干 TimeSeries）
//
// 读路径：remote_write 接收端通常同时暴露 Prometheus HTTP API（/api/v1/query_range），
// QueryMean / QueryRangeValues / QueryRangeWithTags 直接打远端，与 VM 后端同语义。
// 这样三种 backend 互不隐式依赖：TDengine 打 TDengine、VM 打 VM、remote_write 打远端。
type TimeSeriesPromRemoteWriteHandler struct {
	MetricNamer

	url            string
	queryBase      string // 显式 query 端点 base；空则从 url 剥 /api/v1/write 推导
	bearerToken    string
	username       string
	password       string
	tenantID       string
	timeout        time.Duration
	maxRetries     int
	externalLabels []prompb.Label

	// 远端推送 / 远端查询双用途 client；和远端是同一服务，连接复用不分家
	client *http.Client
}

// NewTimeSeriesPromRemoteWriteHandler 创建纯出站写入的 handler
func NewTimeSeriesPromRemoteWriteHandler(cfg promremotewrite.Config) *TimeSeriesPromRemoteWriteHandler {
	ext := make([]prompb.Label, 0, len(cfg.ExternalLabels))
	for k, v := range cfg.ExternalLabels {
		ext = append(ext, prompb.Label{Name: k, Value: v})
	}
	return &TimeSeriesPromRemoteWriteHandler{
		url:            strings.TrimRight(strings.TrimSpace(cfg.URL), "/"),
		queryBase:      strings.TrimRight(strings.TrimSpace(cfg.QueryBase), "/"),
		bearerToken:    strings.TrimSpace(cfg.BearerToken),
		username:       cfg.Username,
		password:       cfg.Password,
		tenantID:       strings.TrimSpace(cfg.TenantID),
		timeout:        cfg.Timeout(),
		maxRetries:     cfg.MaxRetries,
		externalLabels: ext,
		client:         &http.Client{Timeout: cfg.Timeout()},
	}
}

// ---------------- TimeSeriesStore ----------------

// WritePoints 编码 points → prompb.WriteRequest → snappy → POST {url}/api/v1/write
// 失败按指数退避重试（由 maxRetries 控制）；ctx 取消时立刻返回。
func (h *TimeSeriesPromRemoteWriteHandler) WritePoints(ctx context.Context, points []domainHandler.TimeSeriesPoint) error {
	if len(points) == 0 {
		return nil
	}
	wr := h.encodeWriteRequest(points)
	body, err := encodeSnappy(wr)
	if err != nil {
		return fmt.Errorf("remote_write encode: %w", err)
	}

	var lastErr error
	attempts := h.maxRetries + 1
	backoff := rwInitialBackoff
	const maxBackoff = rwMaxBackoff

	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := h.post(ctx, body)
		if err == nil {
			return nil
		}
		lastErr = err
		// 4xx（除 408/429）视作客户端错误，不重试
		var rwe *remoteWriteError
		if errors.As(err, &rwe) && rwe.Status >= 400 && rwe.Status < 500 && rwe.Status != http.StatusRequestTimeout && rwe.Status != http.StatusTooManyRequests {
			return err
		}
		// 最后一次失败直接退出
		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= rwBackoffMul
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return fmt.Errorf("remote_write post after %d attempts: %w", attempts, lastErr)
}

// BatchWrite 分片写入，沿用公共 chunked 框架（片间 sleep 200ms）
func (h *TimeSeriesPromRemoteWriteHandler) BatchWrite(ctx context.Context, points []domainHandler.TimeSeriesPoint, chunkSize int) error {
	return batchWriteChunked(ctx, points, chunkSize, h.WritePoints)
}

// QueryMean / QueryRangeValues / QueryRangeWithTags
// remote_write 接收端通常同时暴露 Prometheus HTTP API，读路径直接打远端，与 VM 后端同语义。
// 这样三种 backend 互不隐式依赖：TDengine 打 TDengine、VM 打 VM、remote_write 打远端。
//
// QueryMean 与 VM handler 逐字对齐：instant query（/api/v1/query + time=end），
// 在 end 时刻回看 [start,end] 窗口做 avg(avg_over_time(...[窗口]))。
// 必须用 instant 而非 query_range：range 查询在多个 step 点各回看一次窗口，
// 返回点数与 step 对齐相关，取末点会因 step 错位引入告警判定漂移。
func (h *TimeSeriesPromRemoteWriteHandler) QueryMean(ctx context.Context, metric string, start, end time.Time) (*float64, error) {
	window := end.Sub(start)
	var query string
	if window <= 0 {
		// 瞬时值：与 VM 的 avg(metric) 退化分支一致
		query = fmt.Sprintf("avg(%s)", escapePromMetric(metric))
	} else {
		query = fmt.Sprintf("avg(avg_over_time(%s[%ds]))", escapePromMetric(metric), int(window.Seconds()))
	}
	return h.queryInstant(ctx, query, end)
}

func (h *TimeSeriesPromRemoteWriteHandler) QueryRangeValues(ctx context.Context, metric string, start, end time.Time) ([]float64, error) {
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

func (h *TimeSeriesPromRemoteWriteHandler) QueryRangeWithTags(ctx context.Context, metric string, start, end time.Time, tagFilters map[string][]string) ([]domainHandler.SeriesData, error) {
	// 复用 buildTagFilteredExpr，与 VM handler 完全相同的 PromQL 拼接逻辑
	raw, err := h.queryRange(ctx, buildPromTagFilteredExpr(metric, tagFilters), start, end)
	if err != nil {
		return nil, err
	}
	series := make([]domainHandler.SeriesData, 0, len(raw))
	for _, r := range raw {
		series = append(series, domainHandler.SeriesData{Labels: r.Metric, Values: r.Values})
	}
	return series, nil
}

// ---------------- 编码 / 发送 ----------------

// encodeWriteRequest 将 TimeSeriesPoint 列表装入 prompb.WriteRequest。
// 关键点：
//   - __name__ 走 Label.Name 字段，其余 tags 顺序拼接，remote_write 接收端按字典序无关；
//   - externalLabels 在每条 series 头部插入，便于接收端识别来源；
//   - Sample.Value 限定为合法 float64（NaN/Inf 在 Prometheus 端会被拒），
//     平台数据来源是 float64，但 Job 在算式里可能产出 NaN/Inf，故此处显式过滤。
func (h *TimeSeriesPromRemoteWriteHandler) encodeWriteRequest(points []domainHandler.TimeSeriesPoint) *prompb.WriteRequest {
	series := make([]prompb.TimeSeries, 0, len(points))
	for _, p := range points {
		// NaN / Inf 在 Prometheus remote_write 协议里属于非法样本，跳过以免整批被拒
		if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
			continue
		}
		// 预分配：外部标签 + __name__ + 业务 tag
		labels := make([]prompb.Label, 0, len(h.externalLabels)+1+len(p.Tags))
		labels = append(labels, h.externalLabels...)
		labels = append(labels, prompb.Label{Name: rwPrometheusNameLabel, Value: p.Metric})
		for k, v := range p.Tags {
			// 跳过空值标签；remote_write 接收端通常拒绝空 label
			if v == "" {
				continue
			}
			labels = append(labels, prompb.Label{Name: k, Value: v})
		}
		series = append(series, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{{Value: p.Value, Timestamp: p.Timestamp.UnixMilli()}},
		})
	}
	return &prompb.WriteRequest{Timeseries: series}
}

// encodeSnappy 把 WriteRequest marshal 成 protobuf 字节，再 snappy 压缩
func encodeSnappy(wr *prompb.WriteRequest) ([]byte, error) {
	raw, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, raw), nil
}

// post 发送一次；非 2xx 返回 remoteWriteError 携带 status/body 摘要
func (h *TimeSeriesPromRemoteWriteHandler) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", rwContentTypeHeader)
	req.Header.Set("Content-Encoding", rwContentEncoding)
	req.Header.Set("User-Agent", rwUserAgent)
	// Prometheus 自身会识别这个头；保持与官方 prometheus client 一致
	req.Header.Set("X-Prometheus-Remote-Write-Version", rwRemoteWriteVersion)

	h.setAuthHeaders(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// 读取最多 1KB 错误体，避免撑大日志
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, rwMaxErrorBodyBytes))
	return &remoteWriteError{
		Status: resp.StatusCode,
		Body:   strings.TrimSpace(string(respBody)),
	}
}

// setAuthHeaders 设置通用鉴权头（Bearer Token / Basic Auth / 多租户 OrgID），post 和 doGet 共用。
func (h *TimeSeriesPromRemoteWriteHandler) setAuthHeaders(req *http.Request) {
	if h.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.bearerToken)
	} else if h.username != "" {
		req.SetBasicAuth(h.username, h.password)
	}
	if h.tenantID != "" {
		// Mimir / Cortex 多租户约定
		req.Header.Set("X-Scope-OrgID", h.tenantID)
	}
}

// queryRange 向远端发起 /api/v1/query_range 查询，承载 QueryRangeValues / QueryRangeWithTags。
//
// URL 推导：优先用显式配置的 queryBase；未配置时从 write URL 剥掉尾部 /api/v1/write
// 得到 base 再拼 /api/v1/query_range。vmagent / Mimir / VM remote_write receiver 等标准部署
// 都满足该约定；非标准路径（如 Grafana Cloud 的 /api/prom/push）需显式配置 queryBase。
//
// 鉴权：复用 bearerToken / username / password / tenantID（远端 query 与 write 通常共用同一鉴权）。
func (h *TimeSeriesPromRemoteWriteHandler) queryRange(ctx context.Context, query string, start, end time.Time) ([]promRemoteSeries, error) {
	endpoint := h.resolveEndpoint("/api/v1/query_range")

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("remote_write query_range parse url: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	// step 自适应：目标 ≤ ~500 点，避免长窗口（如多天采样原料回放）把远端打出上万点响应。
	// 不用固定 15s——VM handler 的 15s 服务 Grafana 面板展示；这里服务采样/下钻拉值，
	// 长窗口固定 15s 会产生海量点，短窗口则必须防 0 除。
	q.Set("step", adaptiveStep(start, end))
	u.RawQuery = q.Encode()

	body, err := h.doGet(ctx, u.String())
	if err != nil {
		return nil, err
	}
	return parseRangeSeries(body)
}

// promRemoteSeries queryRange 返回的单条 series；与 vmSeries 同结构，独立命名以避免误共享
type promRemoteSeries struct {
	Metric map[string]string
	Values []float64
}

// queryInstant 瞬时查询 /api/v1/query（与 VM handler 的 queryInstant 同语义），
// 供 QueryMean 使用：在 ts 时刻回看窗口，单点返回。
// instant 响应的 result[].value 是 [ts, "v"] 两元数组，与 range 的 values 列表结构不同，单独解析。
func (h *TimeSeriesPromRemoteWriteHandler) queryInstant(ctx context.Context, query string, ts time.Time) (*float64, error) {
	endpoint := h.resolveEndpoint("/api/v1/query")

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("remote_write query parse url: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("time", fmt.Sprintf("%d", ts.Unix()))
	u.RawQuery = q.Encode()

	body, err := h.doGet(ctx, u.String())
	if err != nil {
		return nil, err
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value []any `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("remote_write query decode: %w body=%s", err, truncateBody(body))
	}
	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return nil, nil
	}
	s, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return nil, nil
	}
	return parseFloat64(s)
}

// resolveEndpoint 推导 query 端点：显式 queryBase 优先；否则从 write URL 剥 /api/v1/write 尾缀。
// 未匹配到尾缀时（如 Grafana Cloud 的 /api/prom/push）直接以 write URL 为 base 拼接，
// 该场景必须显式配置 queryBase，否则 query 会 404——README 已注明。
func (h *TimeSeriesPromRemoteWriteHandler) resolveEndpoint(path string) string {
	if h.queryBase != "" {
		return h.queryBase + path
	}
	base := h.url
	for _, suffix := range []string{"/api/v1/write", "/api/v1/write/"} {
		if before, ok := strings.CutSuffix(base, suffix); ok {
			base = before
			break
		}
	}
	return base + path
}

// adaptiveStep 根据查询窗口自适应 step：目标 ≤500 点、下限 15s、无 0 除。
// 窗口 ≤15s 时 step=15s（Prometheus 最小 step 粒度，避免过小被拒或除零）。
func adaptiveStep(start, end time.Time) string {
	window := end.Sub(start)
	if window <= rwMinStep {
		return fmt.Sprintf("%ds", int(rwMinStep.Seconds()))
	}
	step := max(window/rwAdaptiveMaxPoints, rwMinStep)
	return fmt.Sprintf("%ds", int(step.Seconds()))
}

// doGet 统一 GET 请求：设置鉴权头 → 发送 → 校验状态码 → 读 body（限 4MB 防异常大响应撑爆内存）
func (h *TimeSeriesPromRemoteWriteHandler) doGet(ctx context.Context, fullURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	h.setAuthHeaders(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, rwMaxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remote_write query status=%d body=%s", resp.StatusCode, truncateBody(body))
	}
	return body, nil
}

// parseRangeSeries 解析 query_range 响应：status.data.result[]，每条 series 含 metric labels + values。
// 空 body / 空 result 都返回空切片而非错误（与 VM handler 行为一致，语义为「无数据」）。
func parseRangeSeries(body []byte) ([]promRemoteSeries, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("remote_write query_range empty body")
	}
	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][]any           `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("remote_write query_range decode: %w body=%s", err, truncateBody(body))
	}
	series := make([]promRemoteSeries, 0, len(result.Data.Result))
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
			if f, err := parseFloat64(s); err == nil {
				vals = append(vals, *f)
			}
		}
		series = append(series, promRemoteSeries{Metric: r.Metric, Values: vals})
	}
	return series, nil
}

// parseFloat64 将 Prometheus 响应中的字符串值转为 float64，queryInstant 和 parseRangeSeries 共用。
func parseFloat64(s string) (*float64, error) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// remoteWriteError 携带 HTTP 状态码与摘要 body，供上层区分 4xx / 5xx
type remoteWriteError struct {
	Status int
	Body   string
}

func (e *remoteWriteError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("remote_write http status=%d", e.Status)
	}
	return fmt.Sprintf("remote_write http status=%d body=%s", e.Status, truncateBody([]byte(e.Body)))
}

// ---------------- MetricQueryDialect（与 VM 一致，便于在远端复刻同一套 PromQL） ----------------

// DatasourceType Grafana Prometheus 数据源
func (h *TimeSeriesPromRemoteWriteHandler) DatasourceType() string {
	return "prometheus"
}

// RealtimeExpr 实时曲线（与 VM handler 保持字面一致；远端任何 PromQL 兼容存储都识别）
func (h *TimeSeriesPromRemoteWriteHandler) RealtimeExpr(taskKey string) string {
	return h.buildAvgOverTimeExpr(h.RealtimeMetric(taskKey))
}

// SmoothExpr 样本曲线
func (h *TimeSeriesPromRemoteWriteHandler) SmoothExpr(taskKey string) string {
	return h.buildAvgOverTimeExpr(h.SmoothMetric(taskKey))
}

// buildAvgOverTimeExpr 构造 avg(avg_over_time(metric[$interval])) 表达式
func (h *TimeSeriesPromRemoteWriteHandler) buildAvgOverTimeExpr(metric string) string {
	return fmt.Sprintf("avg(avg_over_time(%s[$interval]))", escapePromMetric(metric))
}

// BuildPanelTarget Prometheus plugin target
func (h *TimeSeriesPromRemoteWriteHandler) BuildPanelTarget(refID, legend, expr string) map[string]any {
	return map[string]any{
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

// 编译期断言：实现 TimeSeriesStore + MetricQueryDialect
var (
	_ domainHandler.TimeSeriesStore    = (*TimeSeriesPromRemoteWriteHandler)(nil)
	_ domainHandler.MetricQueryDialect = (*TimeSeriesPromRemoteWriteHandler)(nil)
)
