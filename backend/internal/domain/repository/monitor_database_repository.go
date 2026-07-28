package repository

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type MonitorDatabaseRepository interface {
	SelectAll(lastTime *common.LocalTime) ([]entity.MonitorDatabase, error)
	Save(monitorDatabase *entity.MonitorDatabase) error
	UpdateById(id int64, jobDatabase *entity.MonitorDatabase) error
	Delete(id int64) error
	Select(req *types.MonitorDatabaseQueryRequest) (int64, []entity.MonitorDatabase, error)
	SelectSimpleAll() ([]entity.MonitorDatabase, error)
	GetById(id int64) (*entity.MonitorDatabase, error)
	Count() (*int64, error)
}
