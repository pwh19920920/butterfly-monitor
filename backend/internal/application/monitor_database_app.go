package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/pwh19920920/butterfly/pkg/logger"

	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// MonitorDatabaseApplication 数据源应用服务
type MonitorDatabaseApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
	commonMap  *CommonMapApplication
}

// NewMonitorDatabaseApplication 创建数据源应用服务
func NewMonitorDatabaseApplication(sequence *snowflake.Node, repository *persistence.Repository, commonMap *CommonMapApplication) MonitorDatabaseApplication {
	return MonitorDatabaseApplication{sequence: sequence, repository: repository, commonMap: commonMap}
}

// Query 分页查询
func (app *MonitorDatabaseApplication) Query(ctx context.Context, req *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error) {
	total, data, err := app.repository.MonitorDatabaseRepository.Select(req)
	if err != nil {
		logger.Error(ctx, "MonitorDatabaseRepository.Select() happen error for", err)
		return total, nil, err
	}
	for i := range data {
		data[i].Password = ""
		data[i].Salt = ""
	}
	return total, data, nil
}

// QueryAll 全量查询（下拉）
func (app *MonitorDatabaseApplication) QueryAll(ctx context.Context) ([]entity.MonitorDatabase, error) {
	data, err := app.repository.MonitorDatabaseRepository.SelectSimpleAll()
	if err != nil {
		logger.Error(ctx, "MonitorDatabaseRepository.SelectSimpleAll() happen error for", err)
	}
	return data, err
}

// testConnect 按类型取 handler 做连通性测试；失败则不落库
func (app *MonitorDatabaseApplication) testConnect(ctx context.Context, db *entity.MonitorDatabase) error {
	h, ok := app.commonMap.GetDatabaseHandler(ctx, int32(db.Type))
	if !ok {
		return errors.New("不支持的数据源类型")
	}
	if err := h.TestConnect(*db); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}
	return nil
}

// Create 创建数据源：加密密码 → 测试连通 → 保存
func (app *MonitorDatabaseApplication) Create(ctx context.Context, db *entity.MonitorDatabase) error {
	if db.Password == "" {
		return errors.New("密码不能为空")
	}
	if err := db.ResetEncodePasswordAndSalt(db.Password); err != nil {
		return err
	}
	if err := app.testConnect(ctx, db); err != nil {
		return err
	}
	db.Id = app.sequence.Generate().Int64()
	return app.repository.MonitorDatabaseRepository.Save(db)
}

// Modify 修改数据源：密码为空则保留旧密码；先测连通再落库
func (app *MonitorDatabaseApplication) Modify(ctx context.Context, db *entity.MonitorDatabase) error {
	if db.Id == 0 {
		return errors.New("id 不能为空")
	}
	old, err := app.repository.MonitorDatabaseRepository.GetById(db.Id)
	if err != nil {
		return err
	}
	if old == nil {
		return errors.New("数据源不存在")
	}

	if db.Password == "" {
		// 前端未改密时不回传密码，沿用库中密文与盐
		db.Password = old.Password
		db.Salt = old.Salt
	} else {
		if err := db.ResetEncodePasswordAndSalt(db.Password); err != nil {
			return err
		}
	}

	// 使用最终将落库的账号密码做连通性校验，失败则拒绝更新
	if err := app.testConnect(ctx, db); err != nil {
		return err
	}

	if err := app.repository.MonitorDatabaseRepository.UpdateById(db.Id, db); err != nil {
		return err
	}

	// 数据源配置可能已变更（类型/地址/账号密码等），移除旧连接让 refreshDatabaseConnections
	// 在下一轮用新配置重建，避免悬空连接与陈旧连接池泄漏
	app.commonMap.RemoveDatabaseConn(ctx, db.Id, int32(old.Type))
	return nil
}

// GetById 按主键查询
func (app *MonitorDatabaseApplication) GetById(ctx context.Context, id int64) (*entity.MonitorDatabase, error) {
	return app.repository.MonitorDatabaseRepository.GetById(id)
}

// Count 统计
func (app *MonitorDatabaseApplication) Count(ctx context.Context) (int64, error) {
	c, err := app.repository.MonitorDatabaseRepository.Count()
	if err != nil || c == nil {
		return 0, err
	}
	return *c, nil
}
