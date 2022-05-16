package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type AlertChannelRepository interface {

	// Select 查询
	Select(req *types.AlertChannelQueryRequest) (int64, []entity.AlertChannel, error)

	// SelectAll 查询全部
	SelectAll() ([]entity.AlertChannel, error)

	// GetById 获取
	GetById(id int64) (*entity.AlertChannel, error)

	// Delete 删除
	Delete(id int64) error

	// Save 保存
	Save(alertChannel *entity.AlertChannel) error

	// Modify 更新
	Modify(id int64, alertChannel *entity.AlertChannel) error
}
