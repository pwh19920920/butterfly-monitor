package handler

import (
	"context"

	"butterfly-monitor/internal/domain/entity"
)

// DatabaseHandler 数据源连接与查询抽象
type DatabaseHandler interface {
	// TestConnect 测试连通性
	TestConnect(database entity.MonitorDatabase) error
	// NewInstance 创建连接实例
	NewInstance(database entity.MonitorDatabase) (interface{}, error)
	// ExecuteQuery 执行查询，返回单值。
	// ctx 用于任务级超时控制，handler 应透传到 driver 层（如 gorm WithContext / mongo Collection.Aggregate(ctx)），
	// 使慢 SQL 可被截断，避免拖垮下一个采集批次。
	ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error)
	// ExecuteQueryMultiRows 执行聚合查询，返回多行（分组数据收集用）。
	ExecuteQueryMultiRows(ctx context.Context, db interface{}, task entity.MonitorTask) ([]RowResult, error)
	// Close 关闭连接实例，用于数据源被删除或配置变更时释放底层连接池
	Close(db interface{}) error
}
