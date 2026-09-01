package persistence

import (
	"errors"
	"fmt"
	"time"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorTaskRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskRepositoryImpl(db *gorm.DB) *MonitorTaskRepositoryImpl {
	return &MonitorTaskRepositoryImpl{db: db}
}

// FindJobBySharding 分片查询待采集任务（不分页）。
// 条件 date_add(now(), -time_span) >= pre_execute_time 只决定「是否到期可采」；
// 真正的采集窗口始终锚定 now（见 data_collect_job.renderCommandTemplate），
// 落后多个周期也不会按 pre_execute_time 回溯补历史空窗。
func (repo *MonitorTaskRepositoryImpl) FindJobBySharding(shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	if shardTotal <= 0 {
		return nil, fmt.Errorf("shardTotal 必须为正数，当前为 %d（请检查 XXL-JOB 广播路由配置）", shardTotal)
	}
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("mod(id, ?) = ? "+
			"and task_status = ? "+
			"and date_add(now(), interval - time_span second) >= pre_execute_time", shardTotal, shardIndex, entity.MonitorTaskStatusOpen).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// FindSamplingJobBySharding 分片查询待采样任务。
// 排除聚合任务（DataType=Aggregate）：聚合只收集多维点，不生成样本基线。
//
// 到期条件对齐 sampleOne 的 cutoff=now+1d：只要 pre_sample_time 还没追到
// 「now+1 天 - time_span」就继续捞，允许提前生成未来一天的 *_sample 基线。
// 若仍按 now 判断，pre_sample_time 一超过当前时刻就不再入选，+1d 预生成形同虚设。
func (repo *MonitorTaskRepositoryImpl) FindSamplingJobBySharding(pageSize, lastId, shardIndex, shardTotal int64) ([]entity.MonitorTask, error) {
	if shardTotal <= 0 {
		return nil, fmt.Errorf("shardTotal 必须为正数，当前为 %d（请检查 XXL-JOB 广播路由配置）", shardTotal)
	}
	var data []entity.MonitorTask
	err := repo.db.
		Model(&entity.MonitorTask{}).
		Where("id > ? "+
			"and mod(id, ?) = ? "+
			"and task_status = ? "+
			"and data_type <> ? "+
			"and date_add(date_add(now(), interval 1 day), interval -time_span second) >= pre_sample_time",
			lastId, shardTotal, shardIndex, entity.MonitorTaskStatusOpen, entity.DataTypeAggregate).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("id asc").
		Limit(int(pageSize)).
		Find(&data).Error
	return data, err
}

// Save 保存任务、面板关联与告警规则（告警规则为可选，alert 为 nil 时跳过）
func (repo *MonitorTaskRepositoryImpl) Save(monitorTask *entity.MonitorTask, dashboardTasks []entity.MonitorDashboardTask, taskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if len(dashboardTasks) > 0 {
			if err := tx.Model(&entity.MonitorDashboardTask{}).Create(&dashboardTasks).Error; err != nil {
				return err
			}
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
		// 软删除旧关联
		if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorDashboardTask{
			BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
		}).Error; err != nil {
			return err
		}

		if len(dashboardTasks) > 0 {
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
			if err := tx.Model(&entity.MonitorTaskAlert{}).
				Where("task_id = ?", taskAlert.TaskId).
				First(&exist).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err := tx.Model(&entity.MonitorTaskAlert{}).Create(&taskAlert).Error; err != nil {
					return err
				}
			} else {
				taskAlert.Id = exist.Id
				taskAlert.Deleted = common.DeletedFalse
				// 只更新业务字段，保留运行态字段（PreCheckTime、FirstFlagTime）
				if err := tx.Model(&entity.MonitorTaskAlert{}).
					Where("id = ?", exist.Id).
					Select("task_id", "alert_channels", "alert_groups", "time_span", "duration", "params", "alert_status", "deal_status", "deleted").
					Updates(taskAlert).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Model(&entity.MonitorTask{}).
			Where(&entity.MonitorTask{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&monitorTask).Error; err != nil {
			return err
		}
		return nil
	})
}

// UpdateAlertStatusById 更新告警开关
func (repo *MonitorTaskRepositoryImpl) UpdateAlertStatusById(id int64, status entity.MonitorAlertStatus) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if status == entity.MonitorAlertStatusClose {
			// 关闭告警：恢复规则状态，仅忽略未完成事件（Pending/Processing），不污染历史 Complete
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				AlertStatus: entity.MonitorTaskAlertStatusNormal,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("task_id = ? and deal_status in (?, ?)",
				id, entity.MonitorTaskEventDealStatusPending, entity.MonitorTaskEventDealStatusProcessing).
				Updates(&entity.MonitorTaskEvent{
					CompleteTime: &common.LocalTime{Time: time.Now()},
					DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
				}).Error; err != nil {
				return err
			}
		}

		if status == entity.MonitorAlertStatusOpen {
			// 开启告警：刷新首次异常时间与上次检查时间，避免冷启动
			now := time.Now()
			if err := tx.Where("task_id = ?", id).Updates(&entity.MonitorTaskAlert{
				FirstFlagTime: &common.LocalTime{Time: now},
				PreCheckTime:  &common.LocalTime{Time: now},
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
	return repo.updateColumnById(id, "task_status", status)
}

// UpdateSampledById 更新采样展示开关
func (repo *MonitorTaskRepositoryImpl) UpdateSampledById(id int64, status entity.MonitorSampledStatus) error {
	return repo.updateColumnById(id, "sampled", status)
}

// updateColumnById 按 id 更新单列
func (repo *MonitorTaskRepositoryImpl) updateColumnById(id int64, column string, value interface{}) error {
	return repo.db.Model(&entity.MonitorTask{}).
		Where("id = ?", id).
		UpdateColumn(column, value).Error
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
	if len(ids) == 0 {
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

// CountDrilldownBySourceTaskId 统计依赖指定聚合任务的下钻任务数。
// onlyOpen=true 时仅统计处于开启状态的下钻；onlyOpen=false 时不区分开关。
func (repo *MonitorTaskRepositoryImpl) CountDrilldownBySourceTaskId(sourceTaskId int64, onlyOpen bool) (int64, error) {
	var count int64
	idStr := fmt.Sprintf("%d", sourceTaskId)
	db := repo.db.Model(&entity.MonitorTask{}).
		Where("task_type = ? "+
			"and (JSON_UNQUOTE(JSON_EXTRACT(exec_params, '$.sourceTaskId')) = ? "+
			"or JSON_EXTRACT(exec_params, '$.sourceTaskId') = ?)",
			entity.TaskTypeDrilldown, idStr, sourceTaskId).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}})
	if onlyOpen {
		db = db.Where("task_status = ?", entity.MonitorTaskStatusOpen)
	}
	err := db.Count(&count).Error
	return count, err
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
	whereArg := make([]interface{}, 0)
	whereSql := "1 = 1 "
	if req.TaskName != "" {
		whereSql += " and task_name like ?"
		whereArg = append(whereArg, "%"+req.TaskName+"%")
	}
	if req.TaskKey != "" {
		whereSql += " and task_key like ?"
		whereArg = append(whereArg, "%"+req.TaskKey+"%")
	}
	if req.TaskType != nil {
		whereSql += " and task_type = ?"
		whereArg = append(whereArg, req.TaskType)
	}
	if req.TaskStatus != nil {
		whereSql += " and task_status = ?"
		whereArg = append(whereArg, req.TaskStatus)
	}
	if req.AlertStatus != nil {
		whereSql += " and alert_status = ?"
		whereArg = append(whereArg, req.AlertStatus)
	}
	if req.DataType != nil {
		whereSql += " and data_type = ?"
		whereArg = append(whereArg, req.DataType)
	}

	notCase := &entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	return paginate[entity.MonitorTask](repo.db, &entity.MonitorTask{}, whereSql, whereArg, notCase, req.PageSize, req.Offset(), "id desc")
}

// SelectAll 查询全部
func (repo *MonitorTaskRepositoryImpl) SelectAll() ([]entity.MonitorTask, error) {
	var data []entity.MonitorTask
	err := repo.db.Model(&entity.MonitorTask{}).
		Not(&entity.MonitorTask{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}
