package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonitorDashboardTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorDashboardTaskRepositoryImpl(db *gorm.DB) *MonitorDashboardTaskRepositoryImpl {
	return &MonitorDashboardTaskRepositoryImpl{db: db}
}

// FindByDashboardId 按面板 id 查询关联
func (repo *MonitorDashboardTaskRepositoryImpl) FindByDashboardId(dashboardId int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	err := repo.db.Model(&entity.MonitorDashboardTask{}).
		Where("dashboard_id = ?", dashboardId).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").
		Find(&data).Error
	return data, err
}

// FindByTaskId 按任务 id 查询关联
func (repo *MonitorDashboardTaskRepositoryImpl) FindByTaskId(taskId int64) ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	err := repo.db.Model(&entity.MonitorDashboardTask{}).
		Where("task_id = ?", taskId).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("sort desc").
		Find(&data).Error
	return data, err
}

// FindMonitorDashBoardsByTaskId 按任务 id 查询关联的面板列表
func (repo *MonitorDashboardTaskRepositoryImpl) FindMonitorDashBoardsByTaskId(taskId int64) ([]entity.MonitorDashboard, error) {
	dashboardTasks, err := repo.FindByTaskId(taskId)
	if err != nil {
		return nil, err
	}

	dashboardIds := make([]int64, 0)
	for _, item := range dashboardTasks {
		dashboardIds = append(dashboardIds, item.DashboardId)
	}
	if len(dashboardIds) == 0 {
		return make([]entity.MonitorDashboard, 0), nil
	}

	var data []entity.MonitorDashboard
	err = repo.db.Model(&entity.MonitorDashboard{}).
		Where("id in ?", dashboardIds).
		Not(&entity.MonitorDashboard{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// BatchModifySort 批量修改排序（insert on duplicate key update）
func (repo *MonitorDashboardTaskRepositoryImpl) BatchModifySort(items []entity.MonitorDashboardTask) error {
	return repo.db.Model(&entity.MonitorDashboardTask{}).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"sort"}),
	}).Create(items).Error
}

// SelectAll 查询全部关联
func (repo *MonitorDashboardTaskRepositoryImpl) SelectAll() ([]entity.MonitorDashboardTask, error) {
	var data []entity.MonitorDashboardTask
	err := repo.db.Model(&entity.MonitorDashboardTask{}).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}
