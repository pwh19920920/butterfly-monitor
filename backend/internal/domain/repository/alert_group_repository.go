package repository

import (
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
)

type AlertGroupRepository interface {
	Save(group *entity.AlertGroup, users []entity.AlertGroupUser) error
	Modify(id int64, group *entity.AlertGroup, users []entity.AlertGroupUser) error
	Select(req *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error)
	SelectAll() ([]entity.AlertGroup, error)
	GetById(id int64) (*entity.AlertGroup, error)
	Count() (int64, error)
}
