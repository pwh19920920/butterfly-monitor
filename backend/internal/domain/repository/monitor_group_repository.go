package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type MonitorGroupRepository interface {
	Save(group *entity.MonitorGroup) error
	Modify(id int64, oldRoute string, group *entity.MonitorGroup) error
	Select(req *types.MonitorGroupQueryRequest) (int64, []entity.MonitorGroup, error)
	SelectAll() ([]entity.MonitorGroup, error)
	GetById(id int64) (*entity.MonitorGroup, error)
	SelectByIds(ids []int64) ([]entity.MonitorGroup, error)
}
