package persistence

import (
	"errors"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
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

// GetById 按主键查询关联
func (repo *MonitorDashboardTaskRepositoryImpl) GetById(id int64) (*entity.MonitorDashboardTask, error) {
	var data entity.MonitorDashboardTask
	err := repo.db.Model(&entity.MonitorDashboardTask{}).
		Where("id = ?", id).
		Not(&entity.MonitorDashboardTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// BatchModifySort 批量更新排序。
// 仅更新已存在的关联：用 CASE WHEN 按 id 一次 UPDATE，避免 OnConflict 对已删/无效 id 插入脏行。
func (repo *MonitorDashboardTaskRepositoryImpl) BatchModifySort(items []entity.MonitorDashboardTask) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(items))
	caseSQL := "CASE id "
	args := make([]interface{}, 0, len(items)*2)
	for _, it := range items {
		ids = append(ids, it.Id)
		caseSQL += "WHEN ? THEN ? "
		args = append(args, it.Id, it.Sort)
	}
	caseSQL += "ELSE sort END"
	return repo.db.Model(&entity.MonitorDashboardTask{}).
		Where("id in ?", ids).
		Update("sort", gorm.Expr(caseSQL, args...)).Error
}
