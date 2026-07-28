package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type MonitorDashboardRepository interface {
	Save(dashboard *entity.MonitorDashboard) error
	UpdateById(id int64, dashboard *entity.MonitorDashboard) error
	Select(req *types.MonitorDashboardQueryRequest) (int64, []entity.MonitorDashboard, error)
	SelectSimpleAll() ([]entity.MonitorDashboard, error)
	GetById(id int64) (*entity.MonitorDashboard, error)
	SelectByIds(ids []int64) ([]entity.MonitorDashboard, error)
	Count() (*int64, error)
	SelectAll() ([]entity.MonitorDashboard, error)
}
