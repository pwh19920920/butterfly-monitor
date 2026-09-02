package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"butterfly-monitor/internal/common/constant"
	"butterfly-monitor/internal/domain/entity"
	domainHandler "butterfly-monitor/internal/domain/handler"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// limitRe 匹配 LIMIT 关键字（大小写不敏感），用于判断 SQL 是否已自带行数限制。
var limitRe = regexp.MustCompile(`\bLIMIT\s+\d+`)

// DatabaseMysqlHandler MySQL 数据源
type DatabaseMysqlHandler struct{}

func (h *DatabaseMysqlHandler) TestConnect(database entity.MonitorDatabase) error {
	dsn, err := buildMysqlDSN(database)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// Close 返回 error，需显式处理，测试连通性场景可忽略关闭失败
	defer func() { _ = db.Close() }()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error())
	}
	return nil
}

func (h *DatabaseMysqlHandler) NewInstance(database entity.MonitorDatabase) (interface{}, error) {
	dsn, err := buildMysqlDSN(database)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		logrus.Errorf("mysql NewInstance fail: %v", err)
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

// Close 关闭 MySQL 连接池
func (h *DatabaseMysqlHandler) Close(db interface{}) error {
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

// buildMysqlDSN 解密密码，并按 GetUrl/GetParams 组装 DSN
func buildMysqlDSN(database entity.MonitorDatabase) (string, error) {
	plain, err := database.GetDecodePassword()
	if err != nil {
		return "", fmt.Errorf("%s - %s db connect open failure: %s", database.GetUrl(), database.Database, err.Error())
	}

	params := database.GetParams()
	if params == "" {
		params = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", database.Username, plain, database.GetUrl(), database.Database, params), nil
}

func (h *DatabaseMysqlHandler) ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return 0, errors.New("invalid mysql connection")
	}

	// 任务级超时透传到 gorm：慢 SQL 在 ctx 截断时由 driver 返回 ctx 错误，避免拖垮下一批次
	gdb = gdb.WithContext(ctx)

	// task.Command 由用户在监控任务界面录入，作为完整 SQL 执行。
	// 监控采集只需返回单个标量，故强制只读：仅允许 SELECT/SHOW/WITH 开头，
	// 拒绝多语句（;）与写/DDL 关键字，避免注入后被监控库被执行任意写或 DROP。
	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return 0, fmt.Errorf("monitor task command rejected: %w", err)
	}

	// 会话级只读事务作为纵深防御：即使静态校验被绕过，DB 层也会拒绝写操作。
	// 仅作用于本次会话，连接归还连接池后由下个使用者重新设置，互不污染。
	execErr := gdb.Exec("SET SESSION TRANSACTION READ ONLY").Error
	if execErr != nil {
		// 数据源账号若无权限设置只读（如老版 MySQL），退化为仅靠语句校验
		logrus.Warnf("set session read-only fail, fallback to static guard only: %v", execErr)
	}

	// 单行查询：row.Scan 在无返回行时返回 sql.ErrNoRows，据此区分"无数据"与真实错误
	var result float64
	row := gdb.Raw(command).Row()
	queryErr := row.Scan(&result)
	// 查询成功但无返回行：以 gorm.ErrRecordNotFound 哨兵表示"无数据"，由上层回落默认值
	if errors.Is(queryErr, sql.ErrNoRows) {
		queryErr = gorm.ErrRecordNotFound
	}

	// 无论查询成功与否，都尝试恢复会话为可写，避免影响连接池中该连接的后续复用
	if resetErr := gdb.Exec("SET SESSION TRANSACTION READ WRITE").Error; resetErr != nil && execErr == nil {
		logrus.Warnf("reset session read-write fail: %v", resetErr)
	}

	if queryErr != nil {
		return 0, queryErr
	}
	return result, nil
}

// validateReadOnlySQL 校验监控采集 SQL 只读：仅允许 SELECT/SHOW/WITH 开头，
// 禁止语句分隔符 ; 与 DML/DDL/事务控制等可能改库的关键字前缀。
func validateReadOnlySQL(sqlText string) error {
	if sqlText == "" {
		return errors.New("empty command")
	}

	// 拒绝多语句：出现分号即认为尝试批量执行（正常采集只需单条查询）
	// 允许分号出现在字符串字面量内的概率极低，且监控 SQL 通常无字面量分号，从严处理
	if strings.Contains(sqlText, ";") {
		return errors.New("multi-statement ';' is not allowed")
	}

	// 取首个有效关键字：剥离去前导空白与括号（如 "(SELECT ...)" 子查询形式也允许）
	trimmed := strings.TrimLeft(sqlText, " \t\r\n(")
	upper := strings.ToUpper(trimmed)

	allowedPrefixes := []string{"SELECT", "SHOW", "WITH", "DESC", "DESCRIBE", "EXPLAIN"}
	allowed := false
	for _, p := range allowedPrefixes {
		if upper == p || strings.HasPrefix(upper, p+" ") || strings.HasPrefix(upper, p+"\t") || strings.HasPrefix(upper, p+"\n") {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("only read-only queries (SELECT/SHOW/WITH/DESC/EXPLAIN) are allowed")
	}

	// 拦截危险子句：SELECT ... INTO（写表）、FOR UPDATE / LOCK IN SHARE MODE（持锁）等
	dangerSubclauses := []string{
		" INTO ", " FOR UPDATE", " LOCK IN SHARE MODE",
	}
	for _, danger := range dangerSubclauses {
		// danger 以空格开头，确保匹配的是独立关键字而非字段名片段；
		// 尾部可能是语句结束无空格，故对 " FOR UPDATE" 这类无尾空格的也放行
		if strings.Contains(upper, danger) {
			return fmt.Errorf("disallowed clause keyword: %s", strings.TrimSpace(danger))
		}
	}

	return nil
}

// ExecuteQueryMultiRows 分组聚合采集：执行只读查询并返回多行，每行按列名映射为 RowResult。
// 复用 ExecuteQuery 的只读校验与会话级只读防御，仅把单值 Scan 改为多行 Rows 扫描。
func (h *DatabaseMysqlHandler) ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	gdb, ok := db.(*gorm.DB)
	if !ok || gdb == nil {
		return nil, errors.New("invalid mysql connection")
	}
	gdb = gdb.WithContext(ctx)

	command := strings.TrimSpace(task.Command)
	if err := validateReadOnlySQL(command); err != nil {
		return nil, fmt.Errorf("monitor task command rejected: %w", err)
	}

	execErr := gdb.Exec("SET SESSION TRANSACTION READ ONLY").Error
	if execErr != nil {
		logrus.Warnf("set session read-only fail, fallback to static guard only: %v", execErr)
	}

	// SQL 未含 LIMIT 时追加 LIMIT 子句，数据库端限制返回行数，避免 OOM
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

	// 无数据行时仍返回列名，供预览接口提取维度（如空表预览）
	if len(results) == 0 && len(cols) > 0 {
		colMap := make(map[string]interface{}, len(cols))
		for _, col := range cols {
			colMap[col] = nil
		}
		results = append(results, domainHandler.RowResult{Columns: colMap})
	}

	if resetErr := gdb.Exec("SET SESSION TRANSACTION READ WRITE").Error; resetErr != nil && execErr == nil {
		logrus.Warnf("reset session read-write fail: %v", resetErr)
	}
	return results, nil
}
