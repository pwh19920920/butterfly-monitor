package persistence

import (
	"errors"
	"time"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskRepositoryImpl(db *gorm.DB) *MonitorTaskRepositoryImpl {
	return &MonitorTaskRepositoryImpl{db: db}
}

// FindJobBySharding 分片查询待采集任务（不分页）
func (repo *MonitorTaskRepositoryImpl) FindJobBySharding(shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("mod(id, ?) = ? "+
			"and task_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_execute_time", shardTotal, shardIndex, entity.MonitorTaskStatusOpen).
		Find(&data).Error
	return data, err
}

// FindSamplingJobBySharding 分片查询待采样任务
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

// Save 保存任务、面板关联与告警规则（告警规则为可选，alert 为 nil 时跳过）
func (repo *MonitorTaskRepositoryImpl) Save(monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MonitorDashboardTask{}).Create(&dashboardTasks).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.MonitorTask{}).Create(&monitorTask).Error; err != nil {
			return err
		}
		if taskAlert != nil {
			return tx.Model(&entity.MonitorTaskAlert{}).Create(&taskAlert).Error
		}
		return nil
	})
}

// UpdateById 按主键更新任务
func (repo *MonitorTaskRepositoryImpl) UpdateById(id int64, monitorTask *entity.MonitorTask) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where(&entity.MonitorTask{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTask).Error
}

// UpdateTaskAndDashboardTaskAndAlertById 同时更新任务、面板关联与告警规则
// taskAlert 语义：nil=清空（软删除）该任务原有告警；非 nil=存在（含已软删除）则恢复并更新，不存在则新增
func (repo *MonitorTaskRepositoryImpl) UpdateTaskAndDashboardTaskAndAlertById(id int64, monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if dashboardTasks != nil {
			// 软删除旧关联
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorDashboardTask{
				BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&entity.MonitorDashboardTask{}).Create(&dashboardTasks).Error; err != nil {
				return err
			}
		}

		if taskAlert == nil {
			// 未提交告警配置：软删除该任务原有告警规则
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
			}).Error; err != nil {
				return err
			}
		} else {
			// 提交了告警配置：按 task_id 查询（含已软删除的记录）
			// 存在则恢复软删除并更新业务字段，不存在则新增
			var exist entity.MonitorTaskAlert
			err := tx.Model(&entity.MonitorTaskAlert{}).
				Where("task_id = ?", taskAlert.TaskId).
				First(&exist).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err := tx.Model(&entity.MonitorTaskAlert{}).Create(&taskAlert).Error; err != nil {
					return err
				}
			} else {
				taskAlert.Id = exist.Id
				// Updates 忽略零值字段，需强制覆盖可能为空的通道/分组，并恢复 deleted
				if err := tx.Model(&entity.MonitorTaskAlert{}).
					Where("id = ?", exist.Id).
					Updates(taskAlert).Error; err != nil {
					return err
				}
				// 空字符串零值不会被 Updates 写入，需强制覆盖（如纯 Webhook 清空分组）
				if err := tx.Model(&entity.MonitorTaskAlert{}).
					Where("id = ?", exist.Id).
					UpdateColumns(map[string]interface{}{
						"alert_channels": taskAlert.AlertChannels,
						"alert_groups":   taskAlert.AlertGroups,
						"deleted":        common.DeletedFalse,
					}).Error; err != nil {
					return err
				}
			}
		}

		return tx.Model(&entity.MonitorTask{}).
			Where(&entity.MonitorTask{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&monitorTask).Error
	})
}

// UpdateAlertStatusById 更新告警开关
func (repo *MonitorTaskRepositoryImpl) UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if status == entity.MonitorAlertStatusClose {
			// 关闭告警：恢复规则状态并忽略未完成事件
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				AlertStatus: entity.MonitorTaskAlertStatusNormal,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskEvent{
				CompleteTime: &common.LocalTime{Time: time.Now()},
				DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
			}).Error; err != nil {
				return err
			}
		}

		if status == entity.MonitorAlertStatusOpen {
			// 开启告警：刷新首次异常时间
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

// UpdateTaskStatusById 更新任务开关
func (repo *MonitorTaskRepositoryImpl) UpdateTaskStatusById(id int64, status entity.MonitorTaskStatus) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		UpdateColumn("task_status", status).Error
}

// UpdateSampledById 更新采样展示开关
func (repo *MonitorTaskRepositoryImpl) UpdateSampledById(id int64, status entity.MonitorSampledStatus) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		UpdateColumn("sampled", status).Error
}

// Delete 软删除
func (repo *MonitorTaskRepositoryImpl) Delete(id int64) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		Updates(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
}

// GetById 按主键查询
func (repo *MonitorTaskRepositoryImpl) GetById(id int64) (*entity.MonitorTask, error) {
	var data entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// SelectByIdsWithMap 批量查询并以 id 为 key 返回 map
func (repo *MonitorTaskRepositoryImpl) SelectByIdsWithMap(ids []int64) (map[int64]entity.MonitorTask, error) {
	data, err := repo.SelectByIds(ids)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]entity.MonitorTask)
	for _, item := range data {
		result[item.Id] = item
	}
	return result, nil
}

// SelectByIds 批量查询
func (repo *MonitorTaskRepositoryImpl) SelectByIds(ids []int64) ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	if ids == nil || len(ids) == 0 {
		return data, nil
	}
	err := repo.db.Model(&entity.MonitorTask{}).
		Where("id in ?", ids).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// SelectByTaskKey 按任务标识查询
func (repo *MonitorTaskRepositoryImpl) SelectByTaskKey(taskKey string) (*entity.MonitorTask, error) {
	var data entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Where("task_key = ?", taskKey).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
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

	notCase := &entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	repo.db.Model(&entity.MonitorTask{}).
		Where(whereSql, whereArg...).
		Not(notCase).
		Count(&count)

	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where(whereSql, whereArg...).
		Not(notCase).
		Order("id desc").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&data).Error
	return count, data, err
}

// SelectAll 查询全部
func (repo *MonitorTaskRepositoryImpl) SelectAll() ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}
