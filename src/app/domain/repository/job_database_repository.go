package repository

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/types"
	"github.com/pwh19920920/butterfly-admin/src/app/common"
)

type JobDatabaseRepository interface {

	// SelectAll 查询全部数据库
	SelectAll(lastTime *common.LocalTime) ([]entity.JobDatabase, error)

	// Save 保存
	Save(jobDatabase *entity.JobDatabase) error

	// UpdateById 更新
	UpdateById(id int64, jobDatabase *entity.JobDatabase) error

	// Delete 删除
	Delete(id int64) error

	// Select 分页查询
	Select(req *types.JobDatabaseQueryRequest) (int64, []entity.JobDatabase, error)
}
