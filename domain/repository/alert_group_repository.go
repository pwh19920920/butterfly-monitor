package repository

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
)

type AlertGroupRepository interface {

	// SelectAll 查询全部
	SelectAll() ([]entity.AlertGroup, error)

	// Select 获取分组
	Select(req *types.AlertGroupQueryRequest) (int64, []entity.AlertGroup, error)

	// Save 保存
	Save(alertGroup *entity.AlertGroup, alertGroupUsers []entity.AlertGroupUser) error

	// Modify 更新
	Modify(id int64, alertGroup *entity.AlertGroup, alertGroupUsers []entity.AlertGroupUser) error
}
