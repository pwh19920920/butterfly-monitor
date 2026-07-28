package handler

import (
	"context"

	"dragonfly-monitor/internal/domain/entity"
)

// CommandHandler 按任务类型执行采集指令
type CommandHandler interface {
	// ExecuteCommand 执行任务指令，返回数值结果。
	// ctx 用于任务级超时控制（如采集单任务 25s 上限），handler 应透传到 driver 层，
	// 使慢查询可被 ctx 截断，避免拖垮下一个采集批次。
	ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error)
}
