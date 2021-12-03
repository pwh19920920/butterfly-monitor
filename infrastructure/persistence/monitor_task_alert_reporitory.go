package persistence

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
	"time"
)

type MonitorTaskAlertRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskAlertRepositoryImpl(db *gorm.DB) *MonitorTaskAlertRepositoryImpl {
	return &MonitorTaskAlertRepositoryImpl{db: db}
}

// FindCheckJob 获取需要执行的job
func (repo *MonitorTaskAlertRepositoryImpl) FindCheckJob(shardIndex, shardTotal int64) ([]entity.MonitorTaskAlert, error) {
	var data []entity.MonitorTaskAlert
	err := repo.db.
		Model(&entity.MonitorTaskAlert{}).
		Where("mod(id, ?) = ? "+
			"and deal_status = ? "+
			"and date_add(now(), interval -time_span second) >= pre_check_time", shardTotal, shardIndex, entity.MonitorTaskAlertDealStatusNormal).
		Find(&data).Error
	return data, err
}

// Modify 更新
func (repo *MonitorTaskAlertRepositoryImpl) Modify(id int64, monitorTaskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Model(&entity.MonitorTaskAlert{}).
		Where(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTaskAlert).Error
}

// ModifyByAlert 更新
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByAlert(whereCase *entity.MonitorTaskAlert, monitorTaskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Model(&entity.MonitorTaskAlert{}).
		Where(whereCase).
		Updates(&monitorTaskAlert).Error
}

// ModifyByPending 更新
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByPending(id int64, currentTime time.Time) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where(&entity.MonitorTaskAlert{
			BaseEntity: common.BaseEntity{Id: id},
			DealStatus: entity.MonitorTaskAlertDealStatusNormal,
		}).Updates(&entity.MonitorTaskAlert{
			PreCheckTime: &common.LocalTime{Time: currentTime},
			AlertStatus:  entity.MonitorTaskAlertStatusPending,
		}).Error
	})
}

// ModifyForNormal 更新
func (repo *MonitorTaskAlertRepositoryImpl) ModifyForNormal(id int64, currentTime time.Time) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 改状态恢复正常
		err := tx.Where(&entity.MonitorTaskAlert{
			BaseEntity: common.BaseEntity{Id: id},
			DealStatus: entity.MonitorTaskAlertDealStatusNormal,
		}).Updates(&entity.MonitorTaskAlert{
			FirstFlagTime: &common.LocalTime{Time: currentTime},
			PreCheckTime:  &common.LocalTime{Time: currentTime},
			AlertStatus:   entity.MonitorTaskAlertStatusNormal,
		}).Error

		if err != nil {
			return err
		}

		// 修改event为恢复
		return tx.Where(&entity.MonitorTaskEvent{
			DealStatus: entity.MonitorTaskEventDealStatusPending,
			AlertId:    id,
		}).Updates(&entity.MonitorTaskEvent{
			CompleteTime: &common.LocalTime{Time: currentTime},
			DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
		}).Error
	})
}

// ModifyByFiring 更新
func (repo *MonitorTaskAlertRepositoryImpl) ModifyByFiring(id int64, currentTime time.Time, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where(&entity.MonitorTaskAlert{
			BaseEntity: common.BaseEntity{Id: id},
			DealStatus: entity.MonitorTaskAlertDealStatusNormal,
		}).Updates(&entity.MonitorTaskAlert{
			PreCheckTime: &common.LocalTime{Time: currentTime},
			AlertStatus:  entity.MonitorTaskAlertStatusFiring,
		}).Error

		if err != nil {
			return err
		}

		var count int64 = 0
		tx.Model(&entity.MonitorTaskEvent{}).Where("alert_id = ? "+
			"and deal_status in (?, ?)", id, entity.MonitorTaskEventDealStatusPending, entity.MonitorTaskEventDealStatusProcessing).Count(&count)

		if count != 0 {
			return nil
		}

		return tx.Create(&monitorTaskEvent).Error
	})
}
