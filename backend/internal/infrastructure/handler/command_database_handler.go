package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"dragonfly-monitor/internal/domain/entity"
)

// ConnProvider 数据源连接提供者
type ConnProvider interface {
	GetDatabaseConn(id int64) (interface{}, bool)
	GetDatabaseHandler(dsType int32) (DatabaseHandlerAdapter, bool)
}

// DatabaseHandlerAdapter 适配 domain DatabaseHandler
type DatabaseHandlerAdapter interface {
	ExecuteQuery(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error)
}

// CommandDataBaseHandler 数据库任务采集
type CommandDataBaseHandler struct {
	// GetConn 取连接
	GetConn func(id int64) (interface{}, bool)

	// GetHandler 按类型取 handler
	GetHandler func(dsType int32) (func(ctx context.Context, db interface{}, task entity.MonitorTask) (float64, error), bool)

	// GetDatabase 按 id 取数据源元信息（用于取 type）
	GetDatabaseType func(id int64) (int32, error)
}

func (h *CommandDataBaseHandler) ExecuteCommand(ctx context.Context, task entity.MonitorTask) (float64, error) {
	var params struct {
		DatabaseId   *int64   `json:"databaseId,string"`
		DefaultValue *float64 `json:"defaultValue"`
	}
	if task.ExecParams == "" {
		return 0, errors.New("execParams 为空")
	}

	if err := json.Unmarshal([]byte(task.ExecParams), &params); err != nil {
		return 0, err
	}

	if params.DatabaseId == nil {
		return 0, errors.New("databaseId 为空")
	}

	conn, ok := h.GetConn(*params.DatabaseId)
	if !ok || conn == nil {
		return 0, fmt.Errorf("dbMap is not contain %d", *params.DatabaseId)
	}

	dsType, err := h.GetDatabaseType(*params.DatabaseId)
	if err != nil {
		return 0, err
	}

	exec, ok := h.GetHandler(dsType)
	if !ok {
		return 0, fmt.Errorf("database handler not found for type %d", dsType)
	}

	val, err := exec(ctx, conn, task)
	if err != nil && params.DefaultValue != nil {
		return *params.DefaultValue, nil
	}
	return val, err
}
