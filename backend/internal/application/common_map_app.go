package application

import (
	"context"
	"sync"

	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/domain/handler"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/butterfly/pkg/logger"
)

// CommonMapApplication 运行期 handler/连接 映射骨架
type CommonMapApplication struct {
	repository *persistence.Repository

	channelHandlerMap  map[string]handler.ChannelHandler
	databaseHandlerMap map[int32]handler.DatabaseHandler
	commandHandlerMap  map[int32]handler.CommandHandler
	databaseConnMap    map[int64]interface{}

	mu sync.RWMutex
}

// NewCommonMapApplication 创建骨架
func NewCommonMapApplication(repository *persistence.Repository) *CommonMapApplication {
	return &CommonMapApplication{
		repository:         repository,
		channelHandlerMap:  make(map[string]handler.ChannelHandler),
		databaseHandlerMap: make(map[int32]handler.DatabaseHandler),
		commandHandlerMap:  make(map[int32]handler.CommandHandler),
		databaseConnMap:    make(map[int64]interface{}),
	}
}

// GetAlertChannelHandlerNameMap 返回可用通道处理器名称
func (app *CommonMapApplication) GetAlertChannelHandlerNameMap(ctx context.Context) map[string]bool {
	app.mu.RLock()
	defer app.mu.RUnlock()
	result := make(map[string]bool, len(app.channelHandlerMap)+2)
	result["ChannelEmailHandler"] = true
	result["ChannelWechatHandler"] = true
	for name := range app.channelHandlerMap {
		result[name] = true
	}
	return result
}

// GetAlertChannelHandlerMap 返回通道类型与可用处理器的绑定关系，供前端下拉联动使用。
// 与 starter.registerHandlers 注册保持一致：邮件(1)→ChannelEmailHandler，Webhook(2)→ChannelWechatHandler。
func (app *CommonMapApplication) GetAlertChannelHandlerMap(ctx context.Context) []types.AlertChannelHandlerVO {
	return []types.AlertChannelHandlerVO{
		{ChannelType: int32(entity.AlertChannelTypeWebhook), Handlers: []string{"ChannelWechatHandler"}},
		{ChannelType: int32(entity.AlertChannelTypeEmail), Handlers: []string{"ChannelEmailHandler"}},
	}
}

// GetChannelHandler 按类名取通道处理器
func (app *CommonMapApplication) GetChannelHandler(ctx context.Context, className string) (handler.ChannelHandler, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	h, ok := app.channelHandlerMap[className]
	return h, ok
}

// RegisterChannelHandler 注册通道处理器
func (app *CommonMapApplication) RegisterChannelHandler(ctx context.Context, h handler.ChannelHandler) {
	if h == nil {
		return
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	app.channelHandlerMap[h.GetClassName()] = h
}

// GetDatabaseHandler 按数据源类型取 handler
func (app *CommonMapApplication) GetDatabaseHandler(ctx context.Context, dsType int32) (handler.DatabaseHandler, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	h, ok := app.databaseHandlerMap[dsType]
	return h, ok
}

// RegisterDatabaseHandler 注册数据源 handler
func (app *CommonMapApplication) RegisterDatabaseHandler(ctx context.Context, dsType int32, h handler.DatabaseHandler) {
	if h == nil {
		return
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	app.databaseHandlerMap[dsType] = h
}

// GetCommandHandler 按任务类型取 command handler
func (app *CommonMapApplication) GetCommandHandler(ctx context.Context, taskType int32) (handler.CommandHandler, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	h, ok := app.commandHandlerMap[taskType]
	return h, ok
}

// RegisterCommandHandler 注册 command handler
func (app *CommonMapApplication) RegisterCommandHandler(ctx context.Context, taskType int32, h handler.CommandHandler) {
	if h == nil {
		return
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	app.commandHandlerMap[taskType] = h
}

// GetDatabaseConn 取数据源连接
func (app *CommonMapApplication) GetDatabaseConn(ctx context.Context, id int64) (interface{}, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	conn, ok := app.databaseConnMap[id]
	return conn, ok
}

// PutDatabaseConn 写入数据源连接
func (app *CommonMapApplication) PutDatabaseConn(ctx context.Context, id int64, conn interface{}) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.databaseConnMap[id] = conn
}

// RemoveDatabaseConn 关闭并移除数据源连接，用于数据源被删除或配置（类型/连接信息）变更时
// 释放底层连接池：按 dsType 取对应 handler 调用 Close，再从 map 删除。返回是否曾存在该连接。
func (app *CommonMapApplication) RemoveDatabaseConn(ctx context.Context, id int64, dsType int32) bool {
	app.mu.Lock()
	conn, ok := app.databaseConnMap[id]
	if ok {
		delete(app.databaseConnMap, id)
	}
	h, hOK := app.databaseHandlerMap[dsType]
	app.mu.Unlock()

	if !ok {
		return false
	}
	// 锁外关闭，避免长时间持锁阻塞其他取连接的请求
	if hOK {
		if err := h.Close(conn); err != nil {
			logger.WarnFormat(ctx, "close database conn id=%d dsType=%d fail: %v", id, dsType, err)
		}
	}
	return true
}
