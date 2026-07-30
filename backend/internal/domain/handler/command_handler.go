package handler

import (
	"context"

	"dragonfly-monitor/internal/domain/entity"
)

// RowResult 多行查询的单行结果：列名 → 值。
// 用列名而非固定顺序，使聚合收集层能按 labelColumns/valueColumns 取数，与 SQL 列序解耦。
type RowResult struct {
	Columns map[string]interface{}
}

// CommandHandler 按任务类型执行采集指令
type CommandHandler interface {
	// ExecuteCommand 执行任务指令，返回数值结果。
	// ctx 用于任务级超时控制（如采集单任务 25s 上限），handler 应透传到 driver 层，
	// 使慢查询可被 ctx 截断，避免拖垮下一个采集批次。
	ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error)

	// ExecuteMultiRows 执行聚合查询，返回多行结果（分组数据收集用）。
	// 仅聚合任务(DataType=Aggregate)调用；正常任务不会调用此方法。
	ExecuteMultiRows(ctx context.Context, task entity.MonitorTask) ([]RowResult, error)
}
