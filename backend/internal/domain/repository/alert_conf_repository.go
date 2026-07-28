package repository

import (
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"
)

type AlertConfRepository interface {
	Save(conf *entity.AlertConf) error
	Modify(id int64, conf *entity.AlertConf) error
	Select(req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error)
	SelectAll() ([]entity.AlertConf, error)
}
