package persistence

import (
	"butterfly-monitor/src/app/domain/entity"
	"butterfly-monitor/src/app/types"
	"github.com/pwh19920920/butterfly-admin/src/app/common"
	"gorm.io/gorm"
)

type JobDatabaseRepositoryImpl struct {
	db *gorm.DB
}

func NewJobDatabaseRepositoryImpl(db *gorm.DB) *JobDatabaseRepositoryImpl {
	return &JobDatabaseRepositoryImpl{db: db}
}

func (repo *JobDatabaseRepositoryImpl) SelectAll(lastTime *common.LocalTime) ([]entity.JobDatabase, error) {
	var data []entity.JobDatabase
	tx := repo.db.Model(&entity.JobDatabase{})
	if lastTime != nil {
		tx.Where("updated_at >= ?", lastTime.Time)
	}
	err := tx.Find(&data).Error
	return data, err
}

// Save 保存
func (repo *JobDatabaseRepositoryImpl) Save(jobDatabase *entity.JobDatabase) error {
	return repo.db.Model(&entity.JobDatabase{}).Create(&jobDatabase).Error
}

// UpdateById 更新
func (repo *JobDatabaseRepositoryImpl) UpdateById(id int64, jobDatabase *entity.JobDatabase) error {
	return repo.db.Model(&entity.JobDatabase{}).
		Where(&entity.JobDatabase{BaseEntity: common.BaseEntity{Id: id}}).Updates(&jobDatabase).Error
}

// Delete 删除
func (repo *JobDatabaseRepositoryImpl) Delete(id int64) error {
	err := repo.db.Model(&entity.JobDatabase{}).
		Where(&entity.JobDatabase{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&entity.JobDatabase{BaseEntity: common.BaseEntity{Deleted: 1}}).Error
	return err
}

// Select 分页查询
func (repo *JobDatabaseRepositoryImpl) Select(req *types.JobDatabaseQueryRequest) (int64, []entity.JobDatabase, error) {
	var count int64 = 0
	whereCase := &entity.JobDatabase{
		Name: req.Name,
		Type: req.Type,
	}
	repo.db.Model(&entity.JobDatabase{}).Where(whereCase).Count(&count)

	var data []entity.JobDatabase
	err := repo.db.Model(&entity.JobDatabase{}).Where(whereCase).Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}
