package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
)

type MonitorDatabaseRepository interface {

	// SelectAll 查询全部数据库
	SelectAll(lastTime *common.LocalTime) ([]entity.MonitorDatabase, error)

	// Save 保存
	Save(monitorDatabase *entity.MonitorDatabase) error

	// UpdateById 更新
	UpdateById(id int64, jobDatabase *entity.MonitorDatabase) error

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error)

	// SelectSimpleAll 简单查询
	SelectSimpleAll() ([]entity.MonitorDatabase, error)

	// Count 统计总数
	Count() (*int64, error)
}
