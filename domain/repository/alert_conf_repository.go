package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type AlertConfRepository interface {

	// SelectAll 查询全部
	SelectAll() ([]entity.AlertConf, error)

	// Select 获取报警配置
	Select(req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error)

	// Delete 删除
	Delete(id int64) error

	// Save 保存
	Save(alertConf *entity.AlertConf) error

	// Modify 更新
	Modify(id int64, alertConf *entity.AlertConf) error
}
