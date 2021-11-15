package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type MonitorDashboardRepository interface {

	// Save 保存
	Save(monitorDashboard *entity.MonitorDashboard) error

	// UpdateById 更新
	UpdateById(id int64, monitorDashboard *entity.MonitorDashboard) error

	// GetById 获取数据
	GetById(id int64) (*entity.MonitorDashboard, error)

	SelectByIds(ids []int64) ([]entity.MonitorDashboard, error)

	// Select 分页查询
	Select(req *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error)

	// SelectSimpleAll 简单查询
	SelectSimpleAll() ([]entity.MonitorDashboard, error)
}
