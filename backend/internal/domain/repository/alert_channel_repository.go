package repository

import (
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
)

type AlertChannelRepository interface {
	Save(channel *entity.AlertChannel) error
	Modify(id int64, channel *entity.AlertChannel) error
	Select(req *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error)
	SelectAll() ([]entity.AlertChannel, error)
	GetById(id int64) (*entity.AlertChannel, error)
	Count() (int64, error)
}
