package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

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
		Not(&entity.AlertConf{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Select 分页查询
func (repo *AlertConfRepositoryImpl) Select(req *types.AlertConfQueryRequest) (int64, []entity.AlertConf, error) {
	var count int64 = 0
	notCase := &entity.AlertConf{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	whereSql := "1 = 1"
	whereArg := make([]interface{}, 0)
	if req.ConfKey != "" {
		whereSql += " and conf_key like ?"
		whereArg = append(whereArg, "%"+req.ConfKey+"%")
	}
	_ = repo.db.Model(&entity.AlertConf{}).Where(whereSql, whereArg...).Not(notCase).Count(&count)

	var data []entity.AlertConf
	err := repo.db.Model(&entity.AlertConf{}).
		Order("id desc").
		Where(whereSql, whereArg...).
		Not(notCase).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// Save 保存
func (repo *AlertConfRepositoryImpl) Save(conf *entity.AlertConf) error {
	return repo.db.Model(&entity.AlertConf{}).Create(&conf).Error
}

// Modify 按主键更新
func (repo *AlertConfRepositoryImpl) Modify(id int64, conf *entity.AlertConf) error {
	return repo.db.Model(&entity.AlertConf{}).
		Where(&entity.AlertConf{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&conf).Error
}
