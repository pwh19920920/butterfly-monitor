package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
)

type AlertConfRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertConfRepositoryImpl(db *gorm.DB) *AlertConfRepositoryImpl {
	return &AlertConfRepositoryImpl{db: db}
}

// SelectAll 查询全部
func (repo *AlertConfRepositoryImpl) SelectAll() ([]entity.AlertConf, error) {
	var data []entity.AlertConf
	err := repo.db.Model(&entity.AlertConf{}).
		Find(&data).Error
	return data, err
}

// Select 获取报警配置
func (repo *AlertConfRepositoryImpl) Select(req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error) {
	var count int64 = 0
	_ = repo.db.Model(&entity.AlertConf{}).Count(&count)

	var data []entity.AlertConf
	err := repo.db.Model(&entity.AlertConf{}).
		Order("id desc").
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// Delete 删除
func (repo *AlertConfRepositoryImpl) Delete(id int64) error {
	err := repo.db.Where("id = ?", id).Delete(&entity.AlertConf{}).Error
	return err
}

// Save 保存
func (repo *AlertConfRepositoryImpl) Save(alertConf *entity.AlertConf) error {
	return repo.db.Model(&entity.AlertConf{}).Create(&alertConf).Error
}

// Modify 更新
func (repo *AlertConfRepositoryImpl) Modify(id int64, alertConf *entity.AlertConf) error {
	return repo.db.Model(&entity.AlertConf{}).
		Where(&entity.AlertConf{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&alertConf).Error
}
