package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"dragonfly-monitor/internal/domain/entity"
	domainHandler "dragonfly-monitor/internal/domain/handler"
)

// ConnProvider 数据源连接提供者
type ConnProvider interface {
	GetDatabaseConn(id int64) (interface{}, bool)
	GetDatabaseHandler(dsType int32) (DatabaseHandlerAdapter, bool)
}

// DatabaseHandlerAdapter 适配 domain DatabaseHandler（含单值与多行聚合查询）
type DatabaseHandlerAdapter = domainHandler.DatabaseHandler

// CommandDataBaseHandler 数据库任务采集
type CommandDataBaseHandler struct {
	// GetConn 取连接
	GetConn func(id int64) (interface{}, bool)

	// GetHandler 按类型取 handler
	GetHandler func(dsType int32) (DatabaseHandlerAdapter, bool)

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

	val, err := exec.ExecuteQuery(ctx, conn, task)
	if err != nil && params.DefaultValue != nil {
		return *params.DefaultValue, nil
	}
	return val, err
}

// ExecuteMultiRows 聚合任务多行取数：分发到具体数据源 handler 的 ExecuteQueryMultiRows。
func (h *CommandDataBaseHandler) ExecuteMultiRows(ctx context.Context, task entity.MonitorTask) ([]domainHandler.RowResult, error) {
	var params struct {
		DatabaseId *int64 `json:"databaseId,string"`
	}
	if task.ExecParams == "" {
		return nil, errors.New("execParams 为空")
	}
	if err := json.Unmarshal([]byte(task.ExecParams), &params); err != nil {
		return nil, err
	}
	if params.DatabaseId == nil {
		return nil, errors.New("databaseId 为空")
	}

	conn, ok := h.GetConn(*params.DatabaseId)
	if !ok || conn == nil {
		return nil, fmt.Errorf("dbMap is not contain %d", *params.DatabaseId)
	}
	dsType, err := h.GetDatabaseType(*params.DatabaseId)
	if err != nil {
		return nil, err
	}

	exec, ok := h.GetHandler(dsType)
	if !ok {
		return nil, fmt.Errorf("database handler not found for type %d", dsType)
	}
	return exec.ExecuteQueryMultiRows(ctx, conn, task)
}
