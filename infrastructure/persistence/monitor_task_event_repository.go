package persistence

import (
	"butterfly-monitor/domain/entity"
	"butterfly-monitor/types"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
	"time"
)

type MonitorTaskEventRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskEventRepositoryImpl(db *gorm.DB) *MonitorTaskEventRepositoryImpl {
	return &MonitorTaskEventRepositoryImpl{db: db}
}

// FindEventJob 获取需要执行的job
func (repo *MonitorTaskEventRepositoryImpl) FindEventJob() ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("deal_status = ? and now() >= next_alert_time", entity.MonitorTaskEventDealStatusPending).
		Find(&data).Error
	return data, err
}

// Create 创建
func (repo *MonitorTaskEventRepositoryImpl) Create(monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).Create(&monitorTaskEvent).Error
}

// Modify 更新
func (repo *MonitorTaskEventRepositoryImpl) Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTaskEvent).Error
}

// ModifyByEvent 更新
func (repo *MonitorTaskEventRepositoryImpl) ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(whereCase).
		Updates(&monitorTaskEvent).Error
}

func (repo *MonitorTaskEventRepositoryImpl) BatchModifyByEvents(eventIds []int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where("id in (?)", eventIds).
		Updates(&monitorTaskEvent).Error
}

func (repo *MonitorTaskEventRepositoryImpl) SelectByTaskId(taskId int64) ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("task_id = ? and deal_status = ?", taskId, entity.MonitorTaskEventDealStatusPending).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

func (repo *MonitorTaskEventRepositoryImpl) Select(req *types.MonitorTaskEventQueryRequest) (int64, []entity.MonitorTaskEvent, error) {
	var count int64 = 0
	repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(&req.MonitorTaskEvent).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count)

	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where(&req.MonitorTaskEvent).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Order("id desc").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&data).Error
	return count, data, err
}

// DealEvent 事件处理中, 更新event, alert
func (repo *MonitorTaskEventRepositoryImpl) DealEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("task_id = ? and id = ? and deal_status = ?", req.TaskId, eventId, entity.MonitorTaskEventDealStatusPending).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Updates(&entity.MonitorTaskEvent{
				DealStatus: entity.MonitorTaskEventDealStatusProcessing,
				DealTime:   &common.LocalTime{Time: time.Now()},
				DealUser:   req.DealUser,
			}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("task_id = ? and deal_status = ?", req.TaskId, entity.MonitorTaskAlertDealStatusNormal).
			Updates(&entity.MonitorTaskAlert{DealStatus: entity.MonitorTaskAlertDealStatusProcessing}).Error
	})
}

// CompleteEvent 事件完成
func (repo *MonitorTaskEventRepositoryImpl) CompleteEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("task_id = ? and id = ? and deal_status = ?", req.TaskId, eventId, entity.MonitorTaskAlertDealStatusProcessing).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Updates(&entity.MonitorTaskEvent{
				DealStatus:   entity.MonitorTaskEventDealStatusComplete,
				CompleteTime: &common.LocalTime{Time: time.Now()},
			}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("task_id = ? and deal_status = ?", req.TaskId, entity.MonitorTaskAlertDealStatusProcessing).
			Updates(&entity.MonitorTaskAlert{DealStatus: entity.MonitorTaskAlertDealStatusNormal}).Error
	})
}
