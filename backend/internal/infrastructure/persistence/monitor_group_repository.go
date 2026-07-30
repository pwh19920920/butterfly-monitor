package persistence

import (
	"errors"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorGroupRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorGroupRepositoryImpl(db *gorm.DB) *MonitorGroupRepositoryImpl {
	return &MonitorGroupRepositoryImpl{db: db}
}

// GetById 按主键查询
func (repo *MonitorGroupRepositoryImpl) GetById(id int64) (*entity.MonitorGroup, error) {
	var data entity.MonitorGroup
	err := repo.db.Model(&entity.MonitorGroup{}).
		Where(&entity.MonitorGroup{BaseEntity: common.BaseEntity{Id: id}}).
		Not(&entity.MonitorGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// SelectAll 查询全部
func (repo *MonitorGroupRepositoryImpl) SelectAll() ([]entity.MonitorGroup, error) {
	var data []entity.MonitorGroup
	err := repo.db.Model(&entity.MonitorGroup{}).
		Not(&entity.MonitorGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectByIds 批量查询
func (repo *MonitorGroupRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorGroup, error) {
	var data []entity.MonitorGroup
	if ids == nil || len(ids) == 0 {
		return data, nil
	}
	err := repo.db.Model(&entity.MonitorGroup{}).
		Where("id in ?", ids).
		Not(&entity.MonitorGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Select 分页查询
func (repo *MonitorGroupRepositoryImpl) Select(req *types.MonitorGroupQueryRequest) (int64, []entity.MonitorGroup, error) {
	var count int64 = 0
	notCase := &entity.MonitorGroup{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	whereSql := "1 = 1"
	whereArg := make([]interface{}, 0)
	if req.Name != "" {
		whereSql += " and name like ?"
		whereArg = append(whereArg, "%"+req.Name+"%")
	}
	_ = repo.db.Model(&entity.MonitorGroup{}).Where(whereSql, whereArg...).Not(notCase).Count(&count)

	var data []entity.MonitorGroup
	err := repo.db.Model(&entity.MonitorGroup{}).
		Order("id desc").
		Where(whereSql, whereArg...).
		Not(notCase).
		Limit(req.PageSize).Offset(req.Offset()).Find(&data).Error
	return count, data, err
}

// Save 保存
func (repo *MonitorGroupRepositoryImpl) Save(group *entity.MonitorGroup) error {
	return repo.db.Model(&entity.MonitorGroup{}).Create(&group).Error
}

// Modify 更新自身并级联子节点 route（REPLACE）
func (repo *MonitorGroupRepositoryImpl) Modify(id int64, oldRoute string, group *entity.MonitorGroup) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 级联更新子节点 route：REPLACE(route, oldRoute, newRoute)
		if err := tx.Model(&entity.MonitorGroup{}).
			Where("route like ?", oldRoute+"%").
			UpdateColumn("route", gorm.Expr("REPLACE(route, ?, ?)", oldRoute, group.Route)).Error; err != nil {
			return err
		}
		return tx.Model(&entity.MonitorGroup{}).
			Where(&entity.MonitorGroup{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&group).Error
	})
}
