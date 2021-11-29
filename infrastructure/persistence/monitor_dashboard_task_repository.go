package persistence

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonitorDashboardTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDashboardTaskRepositoryImpl(db *gorm.DB) *MonitorDashboardTaskRepositoryImpl {
	return &MonitorDashboardTaskRepositoryImpl{db: db}
}

func (repo *MonitorDashboardTaskRepositoryImpl) SelectByTaskIds(taskIds []int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	tx := repo.db.Model(&entity.MonitorDashboardTask{})
	err := tx.Where("task_id in ?", taskIds).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").Find(&data).Error
	return data, err
}

func (repo *MonitorDashboardTaskRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	tx := repo.db.Model(&entity.MonitorDashboardTask{})
	err := tx.Where("id in ?", ids).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").Find(&data).Error
	return data, err
}

func (repo *MonitorDashboardTaskRepositoryImpl) SelectByDashboardId(dashboardId int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	tx := repo.db.Model(&entity.MonitorDashboardTask{})
	err := tx.Where("dashboard_id = ?", dashboardId).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").Find(&data).Error
	return data, err
}

func (repo *MonitorDashboardTaskRepositoryImpl) BatchModifySort(data []entity.MonitorDashboardTask) error {
	// insert into on duplicate key update
	return repo.db.Model(&entity.MonitorDashboardTask{}).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"sort"}),
	}).Create(data).Error
}
