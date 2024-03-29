package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"errors"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
	"time"
)

type MonitorTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskRepositoryImpl(db *gorm.DB) *MonitorTaskRepositoryImpl {
	return &MonitorTaskRepositoryImpl{db: db}
}

func (repo *MonitorTaskRepositoryImpl) FindJobByShardingNoPaging(shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("mod(id, ?) = ? "+
			"and task_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_execute_time", shardTotal, shardIndex, entity.MonitorTaskStatusOpen).
		Find(&data).Error
	return data, err
}

func (repo *MonitorTaskRepositoryImpl) FindJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("id > ? "+
			"and mod(id, ?) = ? "+
			"and task_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_execute_time "+
			"limit 0, ?", lastId, shardTotal, shardIndex, entity.MonitorTaskStatusOpen, pageSize).
		Find(&data).Error
	return data, err
}

func (repo *MonitorTaskRepositoryImpl) FindSamplingJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("id > ? "+
			"and mod(id, ?) = ? "+
			"and task_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_sample_time "+
			"limit 0, ?", lastId, shardTotal, shardIndex, entity.MonitorTaskStatusOpen, pageSize).
		Find(&data).Error
	return data, err
}

// Save 保存
func (repo *MonitorTaskRepositoryImpl) Save(monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert entity.MonitorTaskAlert) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MonitorDashboardTask{}).Create(&dashboardTasks).Error; err != nil {
			return err
		}

		// 监控任务
		if err := tx.Model(&entity.MonitorTask{}).Create(&monitorTask).Error; err != nil {
			return err
		}

		// 保存检测规则
		return tx.Model(&entity.MonitorTaskAlert{}).Create(&taskAlert).Error
	})
}

// UpdateById 更新
func (repo *MonitorTaskRepositoryImpl) UpdateById(id int64, monitorTask *entity.MonitorTask) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&entity.MonitorTask{}).
			Where(&entity.MonitorTask{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&monitorTask).Error
	})
}

// UpdateTaskAndDashboardTaskAndAlertById 更新
func (repo *MonitorTaskRepositoryImpl) UpdateTaskAndDashboardTaskAndAlertById(id int64, monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if dashboardTasks != nil {
			// 删除dashboard_task
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorDashboardTask{
				BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
			}).Error; err != nil {
				return err
			}

			if err := tx.Model(&entity.MonitorDashboardTask{}).Create(&dashboardTasks).Error; err != nil {
				return err
			}
		}

		// 更新taskAlert
		if taskAlert != nil {
			err := tx.Model(&entity.MonitorTaskAlert{}).Where("task_id = ?", taskAlert.TaskId).Updates(taskAlert).Error
			if err != nil {
				return err
			}
		}

		// 更新监控
		return tx.Model(&entity.MonitorTask{}).
			Where(&entity.MonitorTask{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&monitorTask).Error
	})
}

// UpdateAlertStatusById 更新
func (repo *MonitorTaskRepositoryImpl) UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if status == entity.MonitorAlertStatusClose {
			// 更新task_alert
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				AlertStatus: entity.MonitorTaskAlertStatusNormal,
			}).Error; err != nil {
				return err
			}

			// 更新所有task_event
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskEvent{
				CompleteTime: &common.LocalTime{Time: time.Now()},
				DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
			}).Error; err != nil {
				return err
			}
		}

		if status == entity.MonitorAlertStatusOpen {
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				FirstFlagTime: &common.LocalTime{Time: time.Now()},
			}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&entity.MonitorTask{}).
			Where("id = ?", id).
			UpdateColumn("alert_status", status).Error
	})
}

// UpdateTaskStatusById 更新
func (repo *MonitorTaskRepositoryImpl) UpdateTaskStatusById(id int64, status entity.MonitorTaskStatus) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		UpdateColumn("task_status", status).Error
}

// UpdateSampledById 更新
func (repo *MonitorTaskRepositoryImpl) UpdateSampledById(id int64, status entity.MonitorSampledStatus) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		UpdateColumn("sampled", status).Error
}

// Delete 删除
func (repo *MonitorTaskRepositoryImpl) Delete(id int64) error {
	err := repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		Updates(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: 1}}).Error
	return err
}

// GetById 获取对象
func (repo *MonitorTaskRepositoryImpl) GetById(id int64) (*entity.MonitorTask, error) {
	var data entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).First(&data, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// SelectByIdsWithMap 获取对象
func (repo *MonitorTaskRepositoryImpl) SelectByIdsWithMap(ids []int64) (map[int64]entity.MonitorTask, error) {
	data, err := repo.SelectByIds(ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]entity.MonitorTask, 0)
	if data != nil {
		for _, item := range data {
			result[item.Id] = item
		}
	}
	return result, err
}

// SelectByIds 获取对象
func (repo *MonitorTaskRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Where("id in ?", ids).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

func (repo *MonitorTaskRepositoryImpl) SelectByTaskKey(taskKey string) (*entity.MonitorTask, error) {
	var data entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).Where("task_key", taskKey).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// Count 统计总数
func (repo *MonitorTaskRepositoryImpl) Count() (*int64, error) {
	var count int64
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return &count, err
}

// Select 分页查询
func (repo *MonitorTaskRepositoryImpl) Select(req *types.MonitorTaskQueryRequest) (int64, []entity.MonitorTask, error) {
	var count int64 = 0
	whereArg := make([]interface{}, 0)
	whereSql := "1 = 1 "
	if req.TaskName != "" {
		whereSql += "and task_name like ?"
		whereArg = append(whereArg, "%"+req.TaskName+"%")
	}

	if req.TaskKey != "" {
		whereSql += "and task_key like ?"
		whereArg = append(whereArg, "%"+req.TaskKey+"%")
	}

	if req.TaskType != nil {
		whereSql += "and task_type = ?"
		whereArg = append(whereArg, req.TaskType)
	}

	if req.TaskStatus != nil {
		whereSql += "and task_status = ?"
		whereArg = append(whereArg, req.TaskStatus)
	}

	if req.AlertStatus != nil {
		whereSql += "and alert_status = ?"
		whereArg = append(whereArg, req.AlertStatus)
	}

	repo.db.Model(&entity.MonitorTask{}).
		Where(whereSql, whereArg...).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count)

	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where(whereSql, whereArg...).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("id desc").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&data).Error
	return count, data, err
}

// SelectAll 获取对象
func (repo *MonitorTaskRepositoryImpl) SelectAll() ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}
