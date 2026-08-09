package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/config/tdengine"
	domainHandler "butterfly-monitor/internal/domain/handler"
)

// 超级表：统一存 value，metric/day 作 TAG，便于按指标查询与 Grafana SQL 面板
const tdStableName = "ts_points"

// TimeSeriesTDengineHandler TDengine 实现：TimeSeriesStore + MetricQueryDialect
// 通过 REST /rest/sql 写入与查询；Grafana 数据源类型 tdengine-datasource
type TimeSeriesTDengineHandler struct {
	MetricNamer
	addr     string
	username string
	password string
	database string
	client   *http.Client

	initMu      sync.Mutex
	initialized bool
}

// NewTimeSeriesTDengineHandler 根据配置创建 TDengine 时序客户端
func NewTimeSeriesTDengineHandler(cfg tdengine.Config) *TimeSeriesTDengineHandler {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		addr = "http://127.0.0.1:6041"
	}
	db := strings.TrimSpace(cfg.Database)
	if db == "" {
		db = "monitor"
	}
	user := strings.TrimSpace(cfg.Username)
	if user == "" {
		user = "root"
	}
	return &TimeSeriesTDengineHandler{
		addr:     strings.TrimRight(addr, "/"),
		username: user,
		password: cfg.Password,
		database: db,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ---------------- TimeSeriesStore ----------------

func (h *TimeSeriesTDengineHandler) WritePoints(ctx context.Context, points []domainHandler.TimeSeriesPoint) error {
	if len(points) == 0 {
		return nil
	}
	if err := h.ensureSchema(ctx); err != nil {
		return err
	}

	// 批量 INSERT：INSERT INTO t1 USING ts_points TAGS (...) VALUES (... ) t2 USING ... VALUES (...)
	var b strings.Builder
	b.WriteString("INSERT INTO")
	for _, p := range points {
		tb := childTableName(p.Metric, p.Tags)
		day := ""
		if p.Tags != nil {
			day = p.Tags["day"]
		}
		tsMs := p.Timestamp.UnixMilli()
		b.WriteString(fmt.Sprintf(
			" %s USING %s TAGS ('%s', '%s') VALUES (%d, %s)",
			quoteIdent(tb),
			tdStableName,
			escapeSQLString(p.Metric),
			escapeSQLString(day),
			tsMs,
			formatFloat(p.Value),
		))
	}
	return h.execSQL(ctx, b.String())
}

func (h *TimeSeriesTDengineHandler) BatchWrite(ctx context.Context, points []domainHandler.TimeSeriesPoint, chunkSize int) error {
	return batchWriteChunked(ctx, points, chunkSize, h.WritePoints)
}

func (h *TimeSeriesTDengineHandler) QueryMean(ctx context.Context, metric string, start, end time.Time) (*float64, error) {
	if err := h.ensureSchema(ctx); err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(
		`SELECT AVG(val) FROM %s WHERE metric='%s' AND ts >= %d AND ts <= %d`,
		tdStableName,
		escapeSQLString(metric),
		start.UnixMilli(),
		end.UnixMilli(),
	)
	vals, err := h.queryFloatColumn(ctx, sql)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	v := vals[0]
	return common.Ptr(v), nil
}

func (h *TimeSeriesTDengineHandler) QueryRangeValues(ctx context.Context, metric string, start, end time.Time) ([]float64, error) {
	if err := h.ensureSchema(ctx); err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(
		`SELECT val FROM %s WHERE metric='%s' AND ts >= %d AND ts <= %d ORDER BY ts`,
		tdStableName,
		escapeSQLString(metric),
		start.UnixMilli(),
		end.UnixMilli(),
	)
	return h.queryFloatColumn(ctx, sql)
}

// QueryRangeWithTags TDengine 实现：当前超级表仅按 metric 列存储，不支持任意动态标签维度，
// 故忽略 tagFilters，按 metric 查询区间内全部值并以空 Labels 返回单个 SeriesData。
// 多维度分组下钻在 TDengine 后端暂不支持，仅 VictoriaMetrics 后端提供完整能力。
func (h *TimeSeriesTDengineHandler) QueryRangeWithTags(ctx context.Context, metric string, start, end time.Time, tagFilters map[string][]string) ([]domainHandler.SeriesData, error) {
	vals, err := h.QueryRangeValues(ctx, metric, start, end)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	return []domainHandler.SeriesData{{Labels: map[string]string{}, Values: vals}}, nil
}

// ---------------- MetricQueryDialect ----------------

func (h *TimeSeriesTDengineHandler) DatasourceType() string {
	// Grafana 官方/常用插件 id
	return "tdengine-datasource"
}

// RealtimeExpr Grafana TDengine SQL（$from/$to 插件宏；$interval 来自大盘分组跨度变量）
func (h *TimeSeriesTDengineHandler) RealtimeExpr(taskKey string) string {
	return fmt.Sprintf(
		"SELECT _wstart AS ts, AVG(val) AS val FROM %s WHERE metric='%s' AND ts >= $from AND ts < $to INTERVAL($interval) FILL(NULL)",
		tdStableName,
		escapeSQLString(h.RealtimeMetric(taskKey)),
	)
}

func (h *TimeSeriesTDengineHandler) SmoothExpr(taskKey string) string {
	return fmt.Sprintf(
		"SELECT _wstart AS ts, AVG(val) AS val FROM %s WHERE metric='%s' AND ts >= $from AND ts < $to INTERVAL($interval) FILL(NULL)",
		tdStableName,
		escapeSQLString(h.SmoothMetric(taskKey)),
	)
}

func (h *TimeSeriesTDengineHandler) BuildPanelTarget(refID, legend, expr string) map[string]interface{} {
	return map[string]interface{}{
		"refId":     refID,
		"sql":       expr,
		"queryType": "SQL",
		"alias":     legend,
		"datasource": map[string]string{
			"type": h.DatasourceType(),
			"uid":  "${datasource}",
		},
	}
}

// ---------------- REST / schema ----------------

func (h *TimeSeriesTDengineHandler) ensureSchema(ctx context.Context) error {
	h.initMu.Lock()
	defer h.initMu.Unlock()

	// 已成功初始化：直接返回，避免每次写入/查询都重复建库建表
	if h.initialized {
		return nil
	}

	// 建库（不带 db path）
	if err := h.execSQLRaw(ctx, "", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quoteIdent(h.database))); err != nil {
		// 失败不缓存：下次调用会重试，避免因临时抖动/TDengine 重启导致时序功能永久卡死
		return err
	}

	// 超级表
	createStable := fmt.Sprintf(
		"CREATE STABLE IF NOT EXISTS %s (ts TIMESTAMP, val DOUBLE) TAGS (metric NCHAR(256), day NCHAR(32))",
		tdStableName,
	)
	if err := h.execSQLRaw(ctx, h.database, createStable); err != nil {
		return err
	}

	h.initialized = true
	return nil
}

func (h *TimeSeriesTDengineHandler) execSQL(ctx context.Context, sql string) error {
	return h.execSQLRaw(ctx, h.database, sql)
}

func (h *TimeSeriesTDengineHandler) execSQLRaw(ctx context.Context, db, sql string) error {
	_, err := h.doSQL(ctx, db, sql)
	return err
}

func (h *TimeSeriesTDengineHandler) queryFloatColumn(ctx context.Context, sql string) ([]float64, error) {
	resp, err := h.doSQL(ctx, h.database, sql)
	if err != nil {
		return nil, err
	}
	vals := make([]float64, 0)
	// data 行：各列值；找第一个可解析为 float 的列（AVG 结果或 val）
	for _, row := range resp.Data {
		for _, col := range row {
			if col == nil {
				continue
			}
			if f, ok := toFloat64(col); ok {
				vals = append(vals, f)
				break
			}
		}
	}
	return vals, nil
}

type tdSQLResponse struct {
	Code       int             `json:"code"`
	Desc       string          `json:"desc"`
	ColumnMeta [][]interface{} `json:"column_meta"`
	Data       [][]interface{} `json:"data"`
	Rows       int             `json:"rows"`
}

func (h *TimeSeriesTDengineHandler) doSQL(ctx context.Context, db, sql string) (*tdSQLResponse, error) {
	path := "/rest/sql"
	if db != "" {
		path = "/rest/sql/" + urlPathEscape(db)
	}
	u := h.addr + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBufferString(sql))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tdengine http status=%d body=%s", resp.StatusCode, string(body))
	}

	var result tdSQLResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tdengine decode response: %w body=%s", err, string(body))
	}
	// code=0 成功；部分版本成功也可能无 code 字段，以 desc 兜底
	if result.Code != 0 {
		return nil, fmt.Errorf("tdengine sql error code=%d desc=%s sql=%s", result.Code, result.Desc, truncate(sql, 200))
	}
	return &result, nil
}

// ---------------- helpers ----------------

func childTableName(metric string, tags map[string]string) string {
	// TDengine 表名：字母数字下划线，以字母开头
	base := sanitizeIdent(metric)
	if base == "" {
		base = "m"
	}
	if tags != nil {
		if day := tags["day"]; day != "" {
			base = base + "_d" + sanitizeIdent(day)
		}
	}
	// 避免纯数字开头
	if base[0] >= '0' && base[0] <= '9' {
		base = "m_" + base
	}
	return base
}

func sanitizeIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "") + "`"
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func urlPathEscape(s string) string {
	// 库名通常为标识符，简单 path 拼接；特殊字符做转义
	return strings.ReplaceAll(s, " ", "%20")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
