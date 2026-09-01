package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseClickHouseHandler ClickHouse 数据源（独立方言，不与 MySQL 协议族混用）
type DatabaseClickHouseHandler struct{}

func (h *DatabaseClickHouseHandler) TestConnect(database entity.MonitorDatabase) error {
	dsn, err := buildClickHouseDSN(database)
	if err != nil {
		return err
	}

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error())
	}
	return nil
}

func (h *DatabaseClickHouseHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	dsn, err := buildClickHouseDSN(database)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Errorf("clickHouse NewInstance fail: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// CH 原生连接偏重，连接池略保守
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// Close 关闭 ClickHouse 连接池
func (h *DatabaseClickHouseHandler) Close(db interface{}) error {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return nil
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// buildClickHouseDSN 组装 clickhouse-go v2 URL 形式 DSN。
// Url 约定 host:port（无端口默认 9000 原生协议）；密码经 url.UserPassword 编码。
// 例：clickhouse://default:pass@127.0.0.1:9000/default?dial_timeout=10s&readonly=1
func buildClickHouseDSN(database entity.MonitorDatabase) (string, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return "", fmt.Errorf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error())
	}

	hostPort := strings.TrimSpace(database.GetUrl())
	if hostPort == "" {
		return "", fmt.Errorf("%s - %s db connect open failure: empty url", database.GetUrl(), database.Database)
	}

	host, port, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		host = hostPort
		port = "9000"
	}

	u := &url.URL{
		Scheme: "clickhouse",
		Host:   net.JoinHostPort(host, port),
	}
	if database.Database != "" {
		u.Path = "/" + database.Database
	}
	if database.Username != "" {
		if plain != "" {
			u.User = url.UserPassword(database.Username, plain)
		} else {
			u.User = url.User(database.Username)
		}
	}

	params := strings.TrimSpace(database.GetParams())
	if params == "" {
		// readonly=1：连接级只读，适配监控采集；dial_timeout 避免卡死
		params = "dial_timeout=10s&readonly=1"
	}
	params = strings.Join(strings.Fields(params), "&")
	u.RawQuery = params
	return u.String(), nil
}

func (h *DatabaseClickHouseHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return 0, errors.New("invalid clickhouse connection")
	}
	gdb = gdb.WithContext(ctx)

	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	// CH 无 MySQL 式 SESSION TRANSACTION；连接默认 readonly=1，再叠加静态 SQL 校验
	var raw interface{}
	queryErr := gdb.Raw(command).Row().Scan(&raw)
	if errors.Is(queryErr, sql.ErrNoRows) {
		return 0, gorm.ErrRecordNotFound
	}
	if queryErr != nil {
		return 0, queryErr
	}
	val, err := clickhouseToFloat64(raw)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ExecuteQueryMultiRows 分组聚合采集：只读查询多行，按列名映射 RowResult。
func (h *DatabaseClickHouseHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return nil, errors.New("invalid clickhouse connection")
	}
	gdb = gdb.WithContext(ctx)

	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	hasLimit := limitRe.MatchString(strings.ToUpper(command))
	if !hasLimit {
		command = command + fmt.Sprintf(" LIMIT %d", constant.MaxAggregateRows)
	}
	rows, err := gdb.Raw(command).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]domainHandler.RowResult, 0)
	scanBuf := make([]interface{}, len(cols))
	scanPtr := make([]interface{}, len(cols))
	for i := range scanBuf {
		scanPtr[i] = &scanBuf[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanPtr...); err != nil {
			return nil, err
		}
		row := domainHandler.RowResult{Columns: make(map[string]interface{}, len(cols))}
		for i, col := range cols {
			v := scanBuf[i]
			switch t := v.(type) {
			case []byte:
				row.Columns[col] = string(t)
			default:
				row.Columns[col] = v
			}
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// clickhouseToFloat64 将 CH 常见数值类型（含 UInt64 / Decimal 字符串）转为 float64。
// CH 的 count() 常返回 UInt64，直接 Scan 进 float64 可能失败，故走 interface{} 再转换。
// 不与同包 timeseries_tdengine_handler 的 toFloat64(v) (float64, bool) 同名。
func clickhouseToFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, gorm.ErrRecordNotFound
	}
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint8:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		if t > uint64(math.MaxInt64) {
			// 超出 int64 安全域仍按 float64 近似（监控场景可接受）
			return float64(t), nil
		}
		return float64(t), nil
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, fmt.Errorf("clickHouse value not numeric: %q", t)
		}
		return f, nil
	case []byte:
		f, err := strconv.ParseFloat(string(t), 64)
		if err != nil {
			return 0, fmt.Errorf("clickHouse value not numeric: %q", string(t))
		}
		return f, nil
	default:
		return 0, fmt.Errorf("clickHouse unsupported numeric type %T", v)
	}
}
