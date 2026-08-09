package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabasePostgresHandler PostgreSQL 数据源
type DatabasePostgresHandler struct{}

func (h *DatabasePostgresHandler) TestConnect(database entity.MonitorDatabase) error {
	dsn, err := buildPostgresDSN(database)
	if err != nil {
		return err
	}

	// 用 gorm 打开后立即 Ping，避免额外 blank import pgx/stdlib
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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

func (h *DatabasePostgresHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	dsn, err := buildPostgresDSN(database)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Errorf("postgres NewInstance fail: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// Close 关闭 PostgreSQL 连接池
func (h *DatabasePostgresHandler) Close(db interface{}) error {
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

// buildPostgresDSN 解密密码，按 GetUrl/GetParams 组装 postgres URL 形式 DSN。
// Url 约定为 host:port（无端口时默认 5432）；密码经 url.UserPassword 编码，避免特殊字符破坏 DSN。
// 例：postgres://user:pass@127.0.0.1:5432/dbname?sslmode=disable
func buildPostgresDSN(database entity.MonitorDatabase) (string, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return "", errors.New(fmt.Sprintf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error()))
	}

	hostPort := strings.TrimSpace(database.GetUrl())
	if hostPort == "" {
		return "", fmt.Errorf("%s - %s db connect open failure: empty url", database.GetUrl(), database.Database)
	}

	host, port, splitErr := net.SplitHostPort(hostPort)
	if splitErr != nil {
		// 无端口：整段当 host，默认 5432
		host = hostPort
		port = "5432"
	}

	u := &url.URL{
		Scheme: "postgres",
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
		params = "sslmode=disable"
	}
	// 兼容空格分隔的 libpq 写法：sslmode=disable TimeZone=Local → a=b&c=d
	params = strings.Join(strings.Fields(params), "&")
	u.RawQuery = params
	return u.String(), nil
}

func (h *DatabasePostgresHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return 0, errors.New("invalid postgres connection")
	}

	// 任务级超时透传到 gorm：慢 SQL 在 ctx 截断时由 driver 返回 ctx 错误，避免拖垮下一批次
	gdb = gdb.WithContext(ctx)

	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	// PG 会话级只读：default_transaction_read_only 对后续事务生效（含隐式单语句事务）。
	// SET TRANSACTION READ ONLY 必须在已开启事务内，对连接池场景不适用。
	execErr := gdb.Exec("SET default_transaction_read_only = on").Error
	if execErr != nil {
		logrus.Warnf("set postgres read-only fail, fallback to static guard only: %v", execErr)
	}

	var result float64
	row := gdb.Raw(command).Row()
	queryErr := row.Scan(&result)
	if errors.Is(queryErr, sql.ErrNoRows) {
		queryErr = gorm.ErrRecordNotFound
	}

	// 无论查询成功与否，都恢复会话，避免污染连接池复用
	if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
		logrus.Warnf("reset postgres read-write fail: %v", resetErr)
	}

	if queryErr != nil {
		return 0, queryErr
	}
	return result, nil
}

// ExecuteQueryMultiRows 分组聚合采集：执行只读查询并返回多行，每行按列名映射为 RowResult。
func (h *DatabasePostgresHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return nil, errors.New("invalid postgres connection")
	}
	gdb = gdb.WithContext(ctx)

	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	execErr := gdb.Exec("SET default_transaction_read_only = on").Error
	if execErr != nil {
		logrus.Warnf("set postgres read-only fail, fallback to static guard only: %v", execErr)
	}

	hasLimit := limitRe.MatchString(strings.ToUpper(command))
	if !hasLimit {
		command = command + fmt.Sprintf(" LIMIT %d", constant.MaxAggregateRows)
	}
	rows, err := gdb.Raw(command).Rows()
	if err != nil {
		// 查询失败也尝试恢复会话
		if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
			logrus.Warnf("reset postgres read-write fail: %v", resetErr)
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
			logrus.Warnf("reset postgres read-write fail: %v", resetErr)
		}
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
			if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
				logrus.Warnf("reset postgres read-write fail: %v", resetErr)
			}
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
		if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
			logrus.Warnf("reset postgres read-write fail: %v", resetErr)
		}
		return nil, err
	}

	if resetErr := gdb.Exec("SET default_transaction_read_only = off").Error; resetErr != nil && execErr == nil {
		logrus.Warnf("reset postgres read-write fail: %v", resetErr)
	}
	return results, nil
}
