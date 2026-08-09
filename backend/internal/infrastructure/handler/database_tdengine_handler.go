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

// tdEngineConn 只读 REST SQL 客户端（TDengine taosAdapter /rest/sql）
type tdEngineConn struct {
	baseURL  string
	database string
	username string
	password string
	client   *http.Client
}

// DatabaseTdEngineHandler TDengine 只读数据源（REST SQL 查询，不建表、不 INSERT）
// task.Command 语义：只读 SQL（SELECT/SHOW/DESCRIBE/DESC/EXPLAIN）
//   - 单值：取首行第一个可解析为 float 的列；无行 → ErrRecordNotFound
//   - 多行：每行按 column_meta 列名映射为 RowResult
type DatabaseTdEngineHandler struct{}

func (h *DatabaseTdEngineHandler) TestConnect(database entity.MonitorDatabase) error {
	conn, err := buildTdEngineConn(database)
	if err != nil {
		return err
	}
	// 探活不依赖业务库：SHOW DATABASES
	if _, err := h.doSQL(context.Background(), conn, "", "SHOW DATABASES"); err != nil {
		return fmt.Errorf("%s tdengine connect failure: %w", database.GetUrl(), err)
	}
	return nil
}

func (h *DatabaseTdEngineHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	conn, err := buildTdEngineConn(database)
	if err != nil {
		return nil, err
	}
	if _, err := h.doSQL(context.Background(), conn, "", "SHOW DATABASES"); err != nil {
		// 建连不强制探活成功（部分环境只开放业务库路径）
		logrus.Warnf("tdengine NewInstance probe skip: %v", err)
	}
	return conn, nil
}

// Close TDengine 无长连接池，http.Client 由 GC 回收
func (h *DatabaseTdEngineHandler) Close(db interface{}) error {
	return nil
}

// buildTdEngineConn 组装只读连接。
// Url 约定 host:port（无端口默认 6041）；params 支持 scheme=https、timeout=30s
// database 为业务库名，查询走 /rest/sql/{database}；可空则仅集群级 SQL
func buildTdEngineConn(database entity.MonitorDatabase) (*tdEngineConn, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return nil, fmt.Errorf("%s tdengine connect open failure: %s", database.GetUrl(), err.Error())
	}

	hostPort := strings.TrimSpace(database.GetUrl())
	if hostPort == "" {
		return nil, fmt.Errorf("tdengine url empty")
	}
	host, port, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		host = hostPort
		port = "6041"
	}

	scheme := "http"
	timeout := 30 * time.Second
	params := strings.TrimSpace(database.GetParams())
	if params != "" {
		raw := strings.Join(strings.Fields(params), "&")
		values, parseErr := url.ParseQuery(raw)
		if parseErr == nil {
			if v := values.Get("scheme"); v != "" {
				scheme = strings.ToLower(v)
			}
			if v := values.Get("timeout"); v != "" {
				if d, e := time.ParseDuration(v); e == nil && d > 0 {
					timeout = d
				}
			}
		}
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("tdengine scheme must be http or https, got %s", scheme)
	}

	user := strings.TrimSpace(database.Username)
	if user == "" {
		user = "root"
	}

	return &tdEngineConn{
		baseURL:  fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(host, port)),
		database: strings.TrimSpace(database.Database),
		username: user,
		password: plain,
		client:   &http.Client{Timeout: timeout},
	}, nil
}

func (h *DatabaseTdEngineHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	conn, ok := db.(*tdEngineConn)
	if !ok || conn == nil {
		return 0, errors.New("invalid tdengine connection")
	}
	sqlText := strings.TrimSpace(task.Command)
	if err := validateTdEngineReadOnlySQL(sqlText); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	resp, err := h.doSQL(ctx, conn, conn.database, sqlText)
	if err != nil {
		return 0, err
	}
	if len(resp.Data) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	// 首行第一个可解析为 float 的列
	for _, col := range resp.Data[0] {
		if col == nil {
			continue
		}
		if f, ok := tdCellToFloat64(col); ok {
			return f, nil
		}
	}
	return 0, gorm.ErrRecordNotFound
}

// ExecuteQueryMultiRows 分组聚合：每行按 column_meta 列名映射
func (h *DatabaseTdEngineHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	conn, ok := db.(*tdEngineConn)
	if !ok || conn == nil {
		return nil, errors.New("invalid tdengine connection")
	}
	sqlText := strings.TrimSpace(task.Command)
	if err := validateTdEngineReadOnlySQL(sqlText); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	resp, err := h.doSQL(ctx, conn, conn.database, sqlText)
	if err != nil {
		return nil, err
	}

	colNames := tdColumnNames(resp.ColumnMeta)
	results := make([]domainHandler.RowResult, 0, len(resp.Data))
	for _, row := range resp.Data {
		cols := make(map[string]interface{}, len(row))
		for i, cell := range row {
			name := ""
			if i < len(colNames) {
				name = colNames[i]
			}
			if name == "" {
				name = fmt.Sprintf("col%d", i)
			}
			// 数值尽量转 float64，便于聚合 valueColumns 消费
			if f, ok := tdCellToFloat64(cell); ok {
				cols[name] = f
			} else {
				cols[name] = cell
			}
		}
		results = append(results, domainHandler.RowResult{Columns: cols})
		if len(results) >= constant.MaxAggregateRows {
			logrus.Warnf("aggregate rows truncated at %d for task=%s", constant.MaxAggregateRows, task.TaskKey)
			break
		}
	}
	return results, nil
}

type tdQueryResponse struct {
	Code       int             `json:"code"`
	Desc       string          `json:"desc"`
	ColumnMeta [][]interface{} `json:"column_meta"`
	Data       [][]interface{} `json:"data"`
	Rows       int             `json:"rows"`
}

func (h *DatabaseTdEngineHandler) doSQL(ctx context.Context, conn *tdEngineConn, db, sqlText string) (*tdQueryResponse, error) {
	path := "/rest/sql"
	if db != "" {
		path = "/rest/sql/" + url.PathEscape(db)
	}
	u := strings.TrimRight(conn.baseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBufferString(sqlText))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(conn.username, conn.password)

	resp, err := conn.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("tdengine query canceled by task timeout: %w", ctxErr)
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tdengine http status=%d body=%s", resp.StatusCode, truncateTDBody(body))
	}

	var result tdQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tdengine decode response: %w body=%s", err, truncateTDBody(body))
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("tdengine sql error code=%d desc=%s", result.Code, result.Desc)
	}
	return &result, nil
}

// validateTdEngineReadOnlySQL 只允许只读查询，禁止写/DDL/多语句
func validateTdEngineReadOnlySQL(sqlText string) error {
	if sqlText == "" {
		return errors.New("empty command")
	}
	if strings.Contains(sqlText, ";") {
		return errors.New("multi-statement ';' is not allowed")
	}
	trimmed := strings.TrimLeft(sqlText, " \t\r\n(")
	upper := strings.ToUpper(trimmed)
	allowedPrefixes := []string{"SELECT", "SHOW", "DESCRIBE", "DESC", "EXPLAIN"}
	allowed := false
	for _, p := range allowedPrefixes {
		if upper == p || strings.HasPrefix(upper, p+" ") || strings.HasPrefix(upper, p+"\t") || strings.HasPrefix(upper, p+"\n") {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("only read-only queries (SELECT/SHOW/DESC/EXPLAIN) are allowed")
	}
	// 拦截明显写意图关键字
	for _, bad := range []string{
		" INSERT ", " UPDATE ", " DELETE ", " DROP ", " CREATE ", " ALTER ",
		" TRUNCATE ", " REPLACE ", " GRANT ", " REVOKE ",
	} {
		if strings.Contains(" "+upper+" ", bad) {
			return fmt.Errorf("disallowed keyword in tdengine sql: %s", strings.TrimSpace(bad))
		}
	}
	return nil
}

// tdColumnNames 从 column_meta 提取列名；meta 项常见为 [name, type, length]
func tdColumnNames(meta [][]interface{}) []string {
	names := make([]string, 0, len(meta))
	for i, m := range meta {
		name := ""
		if len(m) > 0 {
			switch v := m[0].(type) {
			case string:
				name = v
			default:
				name = fmt.Sprint(v)
			}
		}
		if name == "" {
			name = fmt.Sprintf("col%d", i)
		}
		names = append(names, name)
	}
	return names
}

func tdCellToFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func truncateTDBody(body []byte) string {
	const maxLen = 500
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "...(truncated)"
}
