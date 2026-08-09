package repository

import (
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"
)

type AlertConfRepository interface {
	Save(conf *entity.AlertConf) error
	Modify(id int64, conf *entity.AlertConf) error
	Select(req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error)
	SelectAll() ([]entity.AlertConf, error)
}
