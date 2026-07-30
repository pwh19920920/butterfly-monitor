package persistence

import (
	"errors"
	"fmt"
	"time"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type MonitorTaskAlertRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskAlertRepositoryImpl(db *gorm.DB) *MonitorTaskAlertRepositoryImpl {
	return &MonitorTaskAlertRepositoryImpl{db: db}
}

// FindCheckJob 分片查询待检测的告警规则
func (repo *MonitorTaskAlertRepositoryImpl) FindCheckJob(shardIndex, shardTotal int64) ([]entity.MonitorTaskAlert, error) {
	if shardTotal <= 0 {
		return nil, fmt.Errorf("shardTotal 必须为正数，当前为 %d（请检查 XXL-JOB 广播路由配置）", shardTotal)
	}
	var data []entity.MonitorTaskAlert
	err := repo.db.
		Model(&entity.MonitorTaskAlert{}).
		Where("mod(id, ?) = ? "+
			"and deal_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_check_time", shardTotal, shardIndex, entity.MonitorTaskAlertDealStatusNormal).
		Find(&data).Error
	return data, err
}

// BatchGetByIds 批量按 id 查询（仅处理中状态为 Normal 的规则）
func (repo *MonitorTaskAlertRepositoryImpl) BatchGetByIds(ids []int64) ([]entity.MonitorTaskAlert, error) {
	if ids == nil || len(ids) == 0 {
		return make([]entity.MonitorTaskAlert, 0), nil
	}
	var data []entity.MonitorTaskAlert
	err := repo.db.
		Model(&entity.MonitorTaskAlert{}).
		Where("deal_status = ? and id in (?)", entity.MonitorTaskAlertDealStatusNormal, ids).
		Find(&data).Error
	return data, err
}

// BatchGetByTaskIds 按任务 id 批量查询（排除软删除）
func (repo *MonitorTaskAlertRepositoryImpl) BatchGetByTaskIds(taskIds []int64) ([]entity.MonitorTaskAlert, error) {
	if taskIds == nil || len(taskIds) == 0 {
		return make([]entity.MonitorTaskAlert, 0), nil
	}
	var data []entity.MonitorTaskAlert
	err := repo.db.
		Model(&entity.MonitorTaskAlert{}).
		Where("task_id in (?)", taskIds).
		Not(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Modify 按主键更新
func (repo *MonitorTaskAlertRepositoryImpl) Modify(id int64, monitorTaskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Model(&entity.MonitorTaskAlert{}).
		Where(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTaskAlert).Error
}

// ModifyByAlert 按条件更新
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByAlert(whereCase *entity.MonitorTaskAlert, monitorTaskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Model(&entity.MonitorTaskAlert{}).
		Where(whereCase).
		Updates(&monitorTaskAlert).Error
}

// ModifyByPending 标记为待触发（Pending）
// 从 Normal 进入异常时写入 FirstFlagTime 作为异常起点；
// 已在 Pending/Firing 时只刷新 PreCheckTime，不重置 FirstFlagTime。
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByPending(id int64, currentTime time.Time) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 首次异常：Normal → Pending，记录异常起点
		res := tx.Model(&entity.MonitorTaskAlert{}).
			Where("id = ? and deal_status = ? and alert_status = ?",
				id, entity.MonitorTaskAlertDealStatusNormal, entity.MonitorTaskAlertStatusNormal).
			Updates(map[string]interface{}{
				"first_flag_time": currentTime,
				"pre_check_time":  currentTime,
				"alert_status":    entity.MonitorTaskAlertStatusPending,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return nil
		}
		// 已在异常中：只推进检查时间与状态，保留 FirstFlagTime
		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("id = ? and deal_status = ?", id, entity.MonitorTaskAlertDealStatusNormal).
			Updates(map[string]interface{}{
				"pre_check_time": currentTime,
				"alert_status":   entity.MonitorTaskAlertStatusPending,
			}).Error
	})
}

// ModifyForNormal 恢复正常，并忽略关联的 Pending 事件
func (repo *MonitorTaskAlertRepositoryImpl) ModifyForNormal(id int64, currentTime time.Time) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&entity.MonitorTaskAlert{
			BaseEntity: common.BaseEntity{Id: id},
			DealStatus: entity.MonitorTaskAlertDealStatusNormal,
		}).Updates(&entity.MonitorTaskAlert{
			FirstFlagTime: &common.LocalTime{Time: currentTime},
			PreCheckTime:  &common.LocalTime{Time: currentTime},
			AlertStatus:   entity.MonitorTaskAlertStatusNormal,
		}).Error; err != nil {
			return err
		}
		return tx.Where(&entity.MonitorTaskEvent{
			DealStatus: entity.MonitorTaskEventDealStatusPending,
			AlertId:    id,
		}).Updates(&entity.MonitorTaskEvent{
			CompleteTime: &common.LocalTime{Time: currentTime},
			DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
		}).Error
	})
}

// ModifyByFiring 标记为触发：存在 pending/processing 事件则更新消息与等级，否则创建事件
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByFiring(id int64, currentTime time.Time, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&entity.MonitorTaskAlert{
			BaseEntity: common.BaseEntity{Id: id},
			DealStatus: entity.MonitorTaskAlertDealStatusNormal,
		}).Updates(&entity.MonitorTaskAlert{
			PreCheckTime: &common.LocalTime{Time: currentTime},
			AlertStatus:  entity.MonitorTaskAlertStatusFiring,
		}).Error; err != nil {
			return err
		}

		var count int64 = 0
		tx.Model(&entity.MonitorTaskEvent{}).Where("alert_id = ? "+
			"and deal_status in (?, ?)", id, entity.MonitorTaskEventDealStatusPending, entity.MonitorTaskEventDealStatusProcessing).Count(&count)

		if count != 0 {
			return tx.Model(&entity.MonitorTaskEvent{}).
				Where("alert_id = ? and deal_status in (?, ?)", id, entity.MonitorTaskEventDealStatusPending, entity.MonitorTaskEventDealStatusProcessing).
				Updates(&entity.MonitorTaskEvent{
					AlertMsg:   monitorTaskEvent.AlertMsg,
					EventLevel: monitorTaskEvent.EventLevel,
				}).Error
		}
		return tx.Create(&monitorTaskEvent).Error
	})
}

// GetByTaskId 按任务 id 查询告警规则
func (repo *MonitorTaskAlertRepositoryImpl) GetByTaskId(taskId int64) (*entity.MonitorTaskAlert, error) {
	var data entity.MonitorTaskAlert
	err := repo.db.Model(&entity.MonitorTaskAlert{}).
		Where("task_id = ?", taskId).
		Not(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}
