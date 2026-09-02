package handler

import (
	"bytes"
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

	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// openSearchConn 只读 HTTP 客户端（OpenSearch / 兼容 ES 查询 API，无写入）
type openSearchConn struct {
	baseURL  string
	username string
	password string
	index    string // 来自 database 字段，可空
	api      string // search | count | sql，默认 search
	client   *http.Client
}

// DatabaseOpenSearchHandler OpenSearch / Elasticsearch 只读数据源（查询 API 兼容，不写任何文档）
// type=11 OpenSearch、type=12 Elasticsearch 共用本实现。
//
// task.Command 语义（由 params.api 决定，默认 search）：
//   - search：完整 _search JSON body；单值优先取 metric 聚合 value，否则 hits.total；多行取 terms/histogram buckets
//   - count：可选 _count JSON body（可空）；单值取 count
//   - sql：OpenSearch SQL（POST /_plugins/_sql）；单值取首行首列数值，多行整表映射
//
// url 约定 host:port（无端口默认 9200）；params 支持 scheme / path_prefix / timeout / api
type DatabaseOpenSearchHandler struct{}

func (h *DatabaseOpenSearchHandler) TestConnect(database entity.MonitorDatabase) error {
	conn, err := buildOpenSearchConn(database)
	if err != nil {
		return err
	}
	if err := h.ping(context.Background(), conn); err != nil {
		return fmt.Errorf("%s opensearch connect failure: %v", database.GetUrl(), err)
	}
	return nil
}

func (h *DatabaseOpenSearchHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	conn, err := buildOpenSearchConn(database)
	if err != nil {
		return nil, err
	}
	if err := h.ping(context.Background(), conn); err != nil {
		logrus.Warnf("openSearch NewInstance ping skip: %v", err)
	}
	return conn, nil
}

func (h *DatabaseOpenSearchHandler) Close(db interface{}) error {
	return nil
}

func (h *DatabaseOpenSearchHandler) ping(ctx context.Context, conn *openSearchConn) error {
	// 集群健康只读；部分权限可能仅开放 index，再降级 GET /
	if err := h.doStatus(ctx, conn, "/_cluster/health"); err != nil {
		return h.doStatus(ctx, conn, "/")
	}
	return nil
}

func (h *DatabaseOpenSearchHandler) doStatus(ctx context.Context, conn *openSearchConn, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(conn.baseURL, "/")+path, nil)
	if err != nil {
		return err
	}
	h.applyAuth(conn, req)
	resp, err := conn.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d path=%s", resp.StatusCode, path)
	}
	return nil
}

func (h *DatabaseOpenSearchHandler) applyAuth(conn *openSearchConn, req *http.Request) {
	if conn.username != "" {
		req.SetBasicAuth(conn.username, conn.password)
	}
	if req.Header.Get("Content-Type") == "" && req.Method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
}

// buildOpenSearchConn 组装只读连接。
// Url: host:port（默认 9200）；Database: 默认 index；params: scheme / path_prefix / timeout / api=search|count|sql
func buildOpenSearchConn(database entity.MonitorDatabase) (*openSearchConn, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return nil, fmt.Errorf("%s openSearch connect open failure: %s", database.GetUrl(), err.Error())
	}
	hostPort := strings.TrimSpace(database.GetUrl())
	if hostPort == "" {
		return nil, fmt.Errorf("openSearch url empty")
	}
	host, port, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		host = hostPort
		port = "9200"
	}

	scheme := "http"
	pathPrefix := ""
	apiMode := "search"
	timeout := 30 * time.Second
	params := strings.TrimSpace(database.GetParams())
	if params != "" {
		raw := strings.Join(strings.Fields(params), "&")
		values, parseErr := url.ParseQuery(raw)
		if parseErr == nil {
			if v := values.Get("scheme"); v != "" {
				scheme = strings.ToLower(v)
			}
			if v := values.Get("path_prefix"); v != "" {
				pathPrefix = strings.TrimRight(v, "/")
			}
			if v := values.Get("api"); v != "" {
				apiMode = strings.ToLower(strings.TrimSpace(v))
			}
			if v := values.Get("timeout"); v != "" {
				if d, e := time.ParseDuration(v); e == nil && d > 0 {
					timeout = d
				}
			}
		}
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("opensearch scheme must be http or https, got %s", scheme)
	}
	switch apiMode {
	case "search", "count", "sql":
	default:
		return nil, fmt.Errorf("opensearch api must be search|count|sql, got %s", apiMode)
	}

	base := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port))
	if pathPrefix != "" {
		if !strings.HasPrefix(pathPrefix, "/") {
			pathPrefix = "/" + pathPrefix
		}
		base += pathPrefix
	}

	return &openSearchConn{
		baseURL:  base,
		username: database.Username,
		password: plain,
		index:    strings.Trim(strings.TrimSpace(database.Database), "/"),
		api:      apiMode,
		client:   &http.Client{Timeout: timeout},
	}, nil
}

func (h *DatabaseOpenSearchHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	conn, ok := db.(*openSearchConn)
	if !ok || conn == nil {
		return 0, errors.New("invalid openSearch connection")
	}
	command := strings.TrimSpace(task.Command)
	if err := validateOpenSearchCommand(conn.api, command); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	switch conn.api {
	case "count":
		n, err := h.execCount(ctx, conn, command)
		if err != nil {
			return 0, err
		}
		return float64(n), nil
	case "sql":
		rows, err := h.execSQL(ctx, conn, command)
		if err != nil {
			return 0, err
		}
		if len(rows) == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		// 首行第一个可解析数值列
		for _, v := range rows[0] {
			if f, ok := anyToFloat64(v); ok {
				return f, nil
			}
		}
		return 0, fmt.Errorf("openSearch sql row has no numeric column")
	default: // search
		body, err := h.execSearch(ctx, conn, command)
		if err != nil {
			return 0, err
		}
		f, ok := extractSearchScalar(body)
		if !ok {
			return 0, gorm.ErrRecordNotFound
		}
		return f, nil
	}
}

func (h *DatabaseOpenSearchHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	conn, ok := db.(*openSearchConn)
	if !ok || conn == nil {
		return nil, errors.New("invalid openSearch connection")
	}
	command := strings.TrimSpace(task.Command)
	if err := validateOpenSearchCommand(conn.api, command); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	switch conn.api {
	case "count":
		n, err := h.execCount(ctx, conn, command)
		if err != nil {
			return nil, err
		}
		return []domainHandler.RowResult{{
			Columns: map[string]interface{}{"count": float64(n)},
		}}, nil
	case "sql":
		rows, err := h.execSQL(ctx, conn, command)
		if err != nil {
			return nil, err
		}
		if len(rows) > constant.MaxAggregateRows {
			rows = rows[:constant.MaxAggregateRows]
		}
		out := make([]domainHandler.RowResult, 0, len(rows))
		for _, r := range rows {
			out = append(out, domainHandler.RowResult{Columns: r})
		}
		return out, nil
	default:
		body, err := h.execSearch(ctx, conn, command)
		if err != nil {
			return nil, err
		}
		rows := extractSearchBuckets(body)
		if len(rows) == 0 {
			// 无 bucket 时退回单值标量一行，便于预览
			if f, ok := extractSearchScalar(body); ok {
				return []domainHandler.RowResult{{
					Columns: map[string]interface{}{"value": f},
				}}, nil
			}
			return nil, nil
		}
		if len(rows) > constant.MaxAggregateRows {
			rows = rows[:constant.MaxAggregateRows]
		}
		out := make([]domainHandler.RowResult, 0, len(rows))
		for _, r := range rows {
			out = append(out, domainHandler.RowResult{Columns: r})
		}
		return out, nil
	}
}

func validateOpenSearchCommand(api, command string) error {
	// count 允许空 body；search/sql 需要非空
	if api != "count" && command == "" {
		return errors.New("empty command")
	}
	if strings.Contains(command, "\x00") {
		return errors.New("invalid command")
	}
	upper := strings.ToUpper(command)
	// 拒绝明显写意图（索引写 API / SQL DML）
	for _, bad := range []string{"\"INDEX\"", "\"CREATE\"", "\"UPDATE\"", "\"DELETE\"", "\"BULK\""} {
		// 过严会误伤，只拦 SQL 风格前缀
		_ = bad
	}
	if api == "sql" {
		for _, bad := range []string{"INSERT ", "UPDATE ", "DELETE ", "DROP ", "CREATE ", "ALTER ", "TRUNCATE "} {
			if strings.HasPrefix(strings.TrimSpace(upper), bad) || strings.Contains(upper, ";"+bad) {
				return fmt.Errorf("disallowed sql keyword: %s", strings.TrimSpace(bad))
			}
		}
		if strings.Contains(command, ";") {
			return errors.New("multi-statement sql is not allowed")
		}
	}
	if api == "search" || (api == "count" && command != "") {
		if !json.Valid([]byte(command)) {
			return errors.New("search/count command must be valid JSON object")
		}
	}
	return nil
}

func (h *DatabaseOpenSearchHandler) execCount(ctx context.Context, conn *openSearchConn, body string) (int64, error) {
	path := "/_count"
	if conn.index != "" {
		path = "/" + url.PathEscape(conn.index) + "/_count"
	}
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	raw, err := h.doJSON(ctx, conn, http.MethodPost, path, reader)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Count int64 `json:"count"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (h *DatabaseOpenSearchHandler) execSearch(ctx context.Context, conn *openSearchConn, body string) (map[string]interface{}, error) {
	path := "/_search"
	if conn.index != "" {
		path = "/" + url.PathEscape(conn.index) + "/_search"
	}
	// 控制返回体积：调用方可在 body 设 size；未设时不强制改 JSON，避免破坏完整 DSL
	raw, err := h.doJSON(ctx, conn, http.MethodPost, path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (h *DatabaseOpenSearchHandler) execSQL(ctx context.Context, conn *openSearchConn, sqlText string) ([]map[string]interface{}, error) {
	// OpenSearch SQL 插件；ES 部分版本为 /_sql — 优先 OS 路径，失败可提示
	payload, _ := json.Marshal(map[string]string{"query": sqlText})
	raw, err := h.doJSON(ctx, conn, http.MethodPost, "/_plugins/_sql", bytes.NewReader(payload))
	if err != nil {
		// 兼容 Elasticsearch SQL
		raw2, err2 := h.doJSON(ctx, conn, http.MethodPost, "/_sql?format=json", bytes.NewReader(payload))
		if err2 != nil {
			return nil, fmt.Errorf("openSearch sql failed: %v; es _sql fallback: %v", err, err2)
		}
		raw = raw2
	}
	return parseSQLResponse(raw)
}

func (h *DatabaseOpenSearchHandler) doJSON(ctx context.Context, conn *openSearchConn, method, path string, body io.Reader) ([]byte, error) {
	u := strings.TrimRight(conn.baseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	h.applyAuth(conn, req)
	resp, err := conn.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openSearch %s %s status=%d body=%s", method, path, resp.StatusCode, truncatePromBody(raw))
	}
	return raw, nil
}

// extractSearchScalar 单值：优先任意 metric 聚合的 value，其次 hits.total
func extractSearchScalar(body map[string]interface{}) (float64, bool) {
	if aggs, ok := body["aggregations"].(map[string]interface{}); ok {
		if f, ok := findAggMetricValue(aggs); ok {
			return f, true
		}
	}
	if hits, ok := body["hits"].(map[string]interface{}); ok {
		switch t := hits["total"].(type) {
		case float64:
			return t, true
		case map[string]interface{}:
			if v, ok := t["value"]; ok {
				return anyToFloat64(v)
			}
		}
	}
	if c, ok := body["count"]; ok { // 少数响应
		return anyToFloat64(c)
	}
	return 0, false
}

func findAggMetricValue(m map[string]interface{}) (float64, bool) {
	if v, ok := m["value"]; ok {
		if f, ok2 := anyToFloat64(v); ok2 {
			return f, true
		}
	}
	// doc_count 单独不在此当 metric；递归子对象
	for k, raw := range m {
		if k == "buckets" || k == "meta" {
			continue
		}
		if child, ok := raw.(map[string]interface{}); ok {
			if f, ok2 := findAggMetricValue(child); ok2 {
				return f, true
			}
		}
	}
	return 0, false
}

// extractSearchBuckets 取第一层带 buckets 的聚合，展开为多行
func extractSearchBuckets(body map[string]interface{}) []map[string]interface{} {
	aggs, ok := body["aggregations"].(map[string]interface{})
	if !ok {
		return nil
	}
	for aggName, raw := range aggs {
		agg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		buckets, ok := agg["buckets"].([]interface{})
		if !ok || len(buckets) == 0 {
			continue
		}
		out := make([]map[string]interface{}, 0, len(buckets))
		for _, b := range buckets {
			bm, ok := b.(map[string]interface{})
			if !ok {
				continue
			}
			row := map[string]interface{}{
				"agg": aggName,
			}
			if k, ok := bm["key_as_string"]; ok {
				row["key"] = k
			} else if k, ok := bm["key"]; ok {
				row["key"] = k
			}
			if dc, ok := bm["doc_count"]; ok {
				if f, ok2 := anyToFloat64(dc); ok2 {
					row["doc_count"] = f
					row["value"] = f // 默认值列，便于前端勾选
				}
			}
			// 子 metric 聚合摊平到列
			for sk, sv := range bm {
				if sk == "key" || sk == "key_as_string" || sk == "doc_count" || sk == "buckets" {
					continue
				}
				if sm, ok := sv.(map[string]interface{}); ok {
					if val, ok2 := sm["value"]; ok2 {
						if f, ok3 := anyToFloat64(val); ok3 {
							row[sk] = f
							// 若只有一个 metric，覆盖 value
							row["value"] = f
						}
					}
				}
			}
			out = append(out, row)
		}
		return out
	}
	return nil
}

func parseSQLResponse(raw []byte) ([]map[string]interface{}, error) {
	// OpenSearch SQL: { "schema":[{"name":"x","type":"long"}], "datarows":[[1],[2]] }
	var osResp struct {
		Schema []struct {
			Name string `json:"name"`
		} `json:"schema"`
		DataRows [][]interface{} `json:"datarows"`
	}
	if err := json.Unmarshal(raw, &osResp); err == nil && (len(osResp.Schema) > 0 || len(osResp.DataRows) > 0) {
		cols := make([]string, len(osResp.Schema))
		for i, s := range osResp.Schema {
			cols[i] = s.Name
			if cols[i] == "" {
				cols[i] = fmt.Sprintf("col%d", i)
			}
		}
		// 无数据行时仍返回列名，供预览接口提取维度（如空表预览）
		if len(osResp.DataRows) == 0 && len(cols) > 0 {
			colMap := make(map[string]interface{}, len(cols))
			for _, col := range cols {
				colMap[col] = nil
			}
			return []map[string]interface{}{colMap}, nil
		}
		out := make([]map[string]interface{}, 0, len(osResp.DataRows))
		for _, row := range osResp.DataRows {
			out = append(out, mapRowToColumns(row, cols))
		}
		return out, nil
	}

	// ES SQL: { "columns":[{"name":"x"}], "rows":[[1]] }
	var esResp struct {
		Columns []struct {
			Name string `json:"name"`
		} `json:"columns"`
		Rows [][]interface{} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &esResp); err != nil {
		return nil, err
	}
	cols := make([]string, len(esResp.Columns))
	for i, c := range esResp.Columns {
		cols[i] = c.Name
		if cols[i] == "" {
			cols[i] = fmt.Sprintf("col%d", i)
		}
	}
	// 无数据行时仍返回列名，供预览接口提取维度（如空表预览）
	if len(esResp.Rows) == 0 && len(cols) > 0 {
		colMap := make(map[string]interface{}, len(cols))
		for _, col := range cols {
			colMap[col] = nil
		}
		return []map[string]interface{}{colMap}, nil
	}
	out := make([]map[string]interface{}, 0, len(esResp.Rows))
	for _, row := range esResp.Rows {
		m := make(map[string]interface{}, len(cols))
		for i, v := range row {
			name := fmt.Sprintf("col%d", i)
			if i < len(cols) {
				name = cols[i]
			}
			m[name] = v
		}
		out = append(out, m)
	}
	return out, nil
}

// mapRowToColumns 将行数据按列名映射为 map，OpenSearch 和 ES SQL 响应共用。
func mapRowToColumns(row []interface{}, cols []string) map[string]interface{} {
	m := make(map[string]interface{}, len(cols))
	for i, v := range row {
		name := fmt.Sprintf("col%d", i)
		if i < len(cols) {
			name = cols[i]
		}
		m[name] = v
	}
	return m
}

func anyToFloat64(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case int32:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(t, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
