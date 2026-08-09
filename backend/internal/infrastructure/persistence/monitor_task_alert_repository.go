package persistence

import (
	"errors"
	"fmt"
	"time"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"

	"gorm.io/gorm"
)

type MonitorTaskAlertRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskAlertRepositoryImpl(db *gorm.DB) *MonitorTaskAlertRepositoryImpl {
	return &MonitorTaskAlertRepositoryImpl{db: db}
}

// FindCheckJob 分片查询待检测的告警规则（单表）。
// 任务关停/告警关闭的跳过在 execCheck 内 advancePreCheckTime 处理，不在此 join。
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
		Not(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
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
		// 已在异常中：只推进检查时间，保留 FirstFlagTime。
		// 仅限 Pending 刷新；Firing 是"已触发"终态，不因暂时缺数据被降级回 Pending
		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("id = ? and deal_status = ? and alert_status = ?",
				id, entity.MonitorTaskAlertDealStatusNormal, entity.MonitorTaskAlertStatusPending).
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

// ModifyByFiring 标记为触发：存在 pending/processing 事件则更新消息与等级，否则创建事件。
// 不加显式行锁，用状态机条件更新串行化「首次进入 Firing」：
//  1. Normal/Pending → Firing：本事务赢得首次触发权，无未完成事件则 Create
//  2. 已是 Firing：只刷新 pre_check_time + 更新已有事件消息，不再 Create
//  3. DealStatus=Processing：整单跳过
//
// InnoDB 同行 UPDATE 会排队；第二事务重评 WHERE 时 status 已是 Firing，RowsAffected=0，不会双建。
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByFiring(id int64, currentTime time.Time, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 首次进入 Firing（Normal/Pending → Firing）
		first := tx.Model(&entity.MonitorTaskAlert{}).
			Where("id = ? and deal_status = ? and alert_status in (?, ?)",
				id, entity.MonitorTaskAlertDealStatusNormal,
				entity.MonitorTaskAlertStatusNormal, entity.MonitorTaskAlertStatusPending).
			Updates(map[string]interface{}{
				"pre_check_time": currentTime,
				"alert_status":   entity.MonitorTaskAlertStatusFiring,
			})
		if first.Error != nil {
			return first.Error
		}

		if first.RowsAffected > 0 {
			// 赢得首次触发：有未完成事件则只更新消息，否则 Create
			return repo.upsertFiringEvent(tx, id, monitorTaskEvent, true)
		}

		// 已是 Firing：只推进检查时间并刷新事件文案，不允许再 Create 新事件
		again := tx.Model(&entity.MonitorTaskAlert{}).
			Where("id = ? and deal_status = ? and alert_status = ?",
				id, entity.MonitorTaskAlertDealStatusNormal, entity.MonitorTaskAlertStatusFiring).
			Updates(map[string]interface{}{
				"pre_check_time": currentTime,
			})
		if again.Error != nil {
			return again.Error
		}
		if again.RowsAffected == 0 {
			// 处理中或不存在
			return nil
		}
		return repo.upsertFiringEvent(tx, id, monitorTaskEvent, false)
	})
}

// upsertFiringEvent 更新未完成事件的消息/等级；allowCreate 时若无未完成事件则新建。
func (repo *MonitorTaskAlertRepositoryImpl) upsertFiringEvent(tx *gorm.DB, alertId int64, event *entity.MonitorTaskEvent, allowCreate bool) error {
	evtRes := tx.Model(&entity.MonitorTaskEvent{}).
		Where("alert_id = ? and deal_status in (?, ?)",
			alertId, entity.MonitorTaskEventDealStatusPending, entity.MonitorTaskEventDealStatusProcessing).
		Updates(map[string]interface{}{
			"alert_msg":   event.AlertMsg,
			"event_level": event.EventLevel,
		})
	if evtRes.Error != nil {
		return evtRes.Error
	}
	if evtRes.RowsAffected > 0 || !allowCreate {
		return nil
	}
	return tx.Create(&event).Error
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

// SoftDeleteAlert 软删除告警规则并忽略关联 Pending 事件。
// 用于任务已删除后清理孤儿规则，避免 FindCheckJob 每轮空捞。
func (repo *MonitorTaskAlertRepositoryImpl) SoftDeleteAlert(id int64) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 软删除告警规则
		if err := tx.Model(&entity.MonitorTaskAlert{}).
			Where(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Id: id}}).
			Updates(&entity.MonitorTaskAlert{
				BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue},
			}).Error; err != nil {
			return err
		}
		// 忽略关联的 Pending 事件
		now := time.Now()
		return tx.Where(&entity.MonitorTaskEvent{
			DealStatus: entity.MonitorTaskEventDealStatusPending,
			AlertId:    id,
		}).Updates(&entity.MonitorTaskEvent{
			CompleteTime: &common.LocalTime{Time: now},
			DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
		}).Error
	})
}
