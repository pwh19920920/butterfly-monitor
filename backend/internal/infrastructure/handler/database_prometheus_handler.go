package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dragonfly-monitor/internal/common/constant"
	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// prometheusConn 只读 HTTP 客户端，无连接池状态（Prometheus 无 SQL 会话）
type prometheusConn struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// DatabasePrometheusHandler Prometheus / VictoriaMetrics 只读数据源（PromQL/MetricsQL 查询，不写任何指标）
// task.Command 语义：完整 PromQL 表达式
//   - 单值：期望返回 0/1 个 series 或 scalar；0 个 → ErrRecordNotFound；多个 series 取第一个（建议用 sum/avg 聚成单值）
//   - 多行：每个 series 映射为 RowResult（labels + value）
//
// DefaultPort 无端口时的默认端口；空则 9090（Prometheus）。VictoriaMetrics 注册时传 8428。
type DatabasePrometheusHandler struct {
	DefaultPort string
}

func (h *DatabasePrometheusHandler) defaultPort() string {
	if h != nil && strings.TrimSpace(h.DefaultPort) != "" {
		return strings.TrimSpace(h.DefaultPort)
	}
	return "9090"
}

func (h *DatabasePrometheusHandler) TestConnect(database entity.MonitorDatabase) error {
	conn, err := buildPrometheusConn(database, h.defaultPort())
	if err != nil {
		return err
	}
	// 优先 /-/healthy；部分发行版/代理可能没有，再试 /api/v1/status/buildinfo
	if err := h.ping(context.Background(), conn, "/-/healthy"); err != nil {
		if err2 := h.ping(context.Background(), conn, "/api/v1/status/buildinfo"); err2 != nil {
			return fmt.Errorf("%s prometheus connect failure: healthy=%v buildinfo=%v", database.GetUrl(), err, err2)
		}
	}
	return nil
}

func (h *DatabasePrometheusHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	conn, err := buildPrometheusConn(database, h.defaultPort())
	if err != nil {
		return nil, err
	}
	if err := h.ping(context.Background(), conn, "/-/healthy"); err != nil {
		// 建连不强制 healthy，部分环境仅开放 query API
		logrus.Warnf("prometheus NewInstance healthy check skip: %v", err)
	}
	return conn, nil
}

// Close Prometheus 无长连接池，http.Client 由 GC 回收即可
func (h *DatabasePrometheusHandler) Close(db interface{}) error {
	return nil
}

func (h *DatabasePrometheusHandler) ping(ctx context.Context, conn *prometheusConn, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(conn.baseURL, "/")+path, nil)
	if err != nil {
		return err
	}
	if conn.username != "" {
		req.SetBasicAuth(conn.username, conn.password)
	}
	resp, err := conn.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}
	return nil
}

// buildPrometheusConn 组装只读连接。
// Url 约定 host:port（无端口用 defaultPort，空则 9090）；params 支持 scheme=https、path_prefix=/prometheus、timeout=30s
func buildPrometheusConn(database entity.MonitorDatabase, defaultPort string) (*prometheusConn, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return nil, fmt.Errorf("%s prometheus connect open failure: %s", database.GetUrl(), err.Error())
	}

	hostPort := strings.TrimSpace(database.GetUrl())
	if hostPort == "" {
		return nil, fmt.Errorf("prometheus url empty")
	}
	if strings.TrimSpace(defaultPort) == "" {
		defaultPort = "9090"
	}
	host, port, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		host = hostPort
		port = defaultPort
	}

	scheme := "http"
	pathPrefix := ""
	timeout := 30 * time.Second
	params := strings.TrimSpace(database.GetParams())
	if params != "" {
		// 兼容空格与 & 分隔：scheme=https path_prefix=/xxx timeout=15s
		raw := strings.Join(strings.Fields(params), "&")
		values, parseErr := url.ParseQuery(raw)
		if parseErr == nil {
			if v := values.Get("scheme"); v != "" {
				scheme = strings.ToLower(v)
			}
			if v := values.Get("path_prefix"); v != "" {
				pathPrefix = strings.TrimRight(v, "/")
			}
			if v := values.Get("timeout"); v != "" {
				if d, e := time.ParseDuration(v); e == nil && d > 0 {
					timeout = d
				}
			}
		}
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("prometheus scheme must be http or https, got %s", scheme)
	}

	base := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port))
	if pathPrefix != "" {
		if !strings.HasPrefix(pathPrefix, "/") {
			pathPrefix = "/" + pathPrefix
		}
		base += pathPrefix
	}

	return &prometheusConn{
		baseURL:  base,
		username: database.Username,
		password: plain,
		client:   &http.Client{Timeout: timeout},
	}, nil
}

func (h *DatabasePrometheusHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	conn, ok := db.(*prometheusConn)
	if !ok || conn == nil {
		return 0, errors.New("invalid prometheus connection")
	}
	expr := strings.TrimSpace(task.Command)
	if expr == "" {
		return 0, errors.New("empty promql command")
	}
	if err := validatePromQLReadOnly(expr); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	samples, err := h.queryInstant(ctx, conn, expr)
	if err != nil {
		return 0, err
	}
	if len(samples) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	// 单值任务：取第一个 series；多 series 时请在 PromQL 侧 sum/avg 收敛
	return samples[0].value, nil
}

// ExecuteQueryMultiRows 分组聚合：每个 series 一行，列 = 全部 label + value
func (h *DatabasePrometheusHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	conn, ok := db.(*prometheusConn)
	if !ok || conn == nil {
		return nil, errors.New("invalid prometheus connection")
	}
	expr := strings.TrimSpace(task.Command)
	if expr == "" {
		return nil, errors.New("empty promql command")
	}
	if err := validatePromQLReadOnly(expr); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	samples, err := h.queryInstant(ctx, conn, expr)
	if err != nil {
		return nil, err
	}
	if len(samples) > constant.MaxAggregateRows {
		samples = samples[:constant.MaxAggregateRows]
	}
	results := make([]domainHandler.RowResult, 0, len(samples))
	for _, s := range samples {
		cols := make(map[string]interface{}, len(s.metric)+1)
		for k, v := range s.metric {
			// 跳过内部 __name__ 可选保留：保留便于维度选择
			cols[k] = v
		}
		cols["value"] = s.value
		results = append(results, domainHandler.RowResult{Columns: cols})
	}
	return results, nil
}

type promSample struct {
	metric map[string]string
	value  float64
}

// queryInstant 调用 GET /api/v1/query，解析 vector / scalar
func (h *DatabasePrometheusHandler) queryInstant(ctx context.Context, conn *prometheusConn, expr string) ([]promSample, error) {
	u, err := url.Parse(strings.TrimRight(conn.baseURL, "/") + "/api/v1/query")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", expr)
	// 不传 time：用服务端 now，与采集周期对齐
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if conn.username != "" {
		req.SetBasicAuth(conn.username, conn.password)
	}
	resp, err := conn.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("prometheus query status=%d body=%s", resp.StatusCode, truncatePromBody(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", result.Error)
	}

	switch result.Data.ResultType {
	case "vector":
		var rows []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}
		if err := json.Unmarshal(result.Data.Result, &rows); err != nil {
			return nil, err
		}
		out := make([]promSample, 0, len(rows))
		for _, r := range rows {
			v, ok := parsePromValue(r.Value)
			if !ok {
				continue
			}
			out = append(out, promSample{metric: r.Metric, value: v})
		}
		return out, nil
	case "scalar":
		// scalar: [ts, "1.23"]
		var pair []interface{}
		if err := json.Unmarshal(result.Data.Result, &pair); err != nil {
			return nil, err
		}
		v, ok := parsePromValue(pair)
		if !ok {
			return nil, nil
		}
		return []promSample{{metric: map[string]string{}, value: v}}, nil
	case "matrix", "string":
		return nil, fmt.Errorf("prometheus resultType %s not supported for collect; use instant vector/scalar PromQL", result.Data.ResultType)
	default:
		return nil, fmt.Errorf("prometheus unknown resultType %s", result.Data.ResultType)
	}
}

// parsePromValue 解析 Prometheus value 对 [timestamp, "number"]
func parsePromValue(pair []interface{}) (float64, bool) {
	if len(pair) < 2 {
		return 0, false
	}
	switch t := pair[1].(type) {
	case string:
		f, err := strconv.ParseFloat(t, 64)
		return f, err == nil
	case float64:
		return t, true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// validatePromQLReadOnly 轻量拒绝明显写操作；PromQL 本身只读，主要防误把 HTTP 管理 API 写进 command
func validatePromQLReadOnly(expr string) error {
	if expr == "" {
		return errors.New("empty command")
	}
	// 拒绝分号多语句与明显管理意图
	if strings.Contains(expr, ";") {
		return errors.New("multi-statement is not allowed")
	}
	upper := strings.ToUpper(strings.TrimSpace(expr))
	// PromQL 无 INSERT 等，但用户可能误填 HTTP path
	for _, bad := range []string{"DELETE ", "DROP ", "INSERT ", "CREATE ", "ALTER "} {
		if strings.HasPrefix(upper, bad) || strings.Contains(upper, " "+bad) {
			return fmt.Errorf("disallowed keyword in promql: %s", strings.TrimSpace(bad))
		}
	}
	return nil
}

func truncatePromBody(body []byte) string {
	const maxLen = 500
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "...(truncated)"
}
