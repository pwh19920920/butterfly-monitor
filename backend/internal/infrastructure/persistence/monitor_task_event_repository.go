package persistence

import (
	"errors"
	"time"

	"butterfly-monitor/internal/common"
	"butterfly-monitor/internal/domain/entity"
	"butterfly-monitor/internal/types"

	"gorm.io/gorm"
)

type MonitorTaskEventRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorTaskEventRepositoryImpl(db *gorm.DB) *MonitorTaskEventRepositoryImpl {
	return &MonitorTaskEventRepositoryImpl{db: db}
}

// FindEventJob 查询到达下次告警时间的 Pending 事件
func (repo *MonitorTaskEventRepositoryImpl) FindEventJob() ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("deal_status = ? and now() >= next_alert_time", entity.MonitorTaskEventDealStatusPending).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// FindPendingEventAll 查询全部 Pending 事件
func (repo *MonitorTaskEventRepositoryImpl) FindPendingEventAll() ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("deal_status = ?", entity.MonitorTaskEventDealStatusPending).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Create 创建事件
func (repo *MonitorTaskEventRepositoryImpl) Create(monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).Create(&monitorTaskEvent).Error
}

// Modify 按主键更新
func (repo *MonitorTaskEventRepositoryImpl) Modify(id int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&monitorTaskEvent).Error
}

// ModifyByEvent 按条件更新
func (repo *MonitorTaskEventRepositoryImpl) ModifyByEvent(whereCase *entity.MonitorTaskEvent, monitorTaskEvent *entity.MonitorTaskEvent) error {
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(whereCase).
		Updates(&monitorTaskEvent).Error
}

// BatchModifyByEvents 批量按 id 更新
func (repo *MonitorTaskEventRepositoryImpl) BatchModifyByEvents(eventIds []int64, monitorTaskEvent *entity.MonitorTaskEvent) error {
	if eventIds == nil || len(eventIds) == 0 {
		return nil
	}
	return repo.db.Model(&entity.MonitorTaskEvent{}).
		Where("id in (?)", eventIds).
		Updates(&monitorTaskEvent).Error
}

// SelectByTaskId 查询任务下 Pending 事件
func (repo *MonitorTaskEventRepositoryImpl) SelectByTaskId(taskId int64) ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Where("task_id = ? and deal_status = ?", taskId, entity.MonitorTaskEventDealStatusPending).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Find(&data).Error
	return data, err
}

// Select 分页查询，支持 TaskId、DealStatus 过滤
func (repo *MonitorTaskEventRepositoryImpl) Select(req *types.MonitorTaskEventQueryRequest) (int64, []entity.MonitorTaskEvent, error) {
	whereCase := "1 = 1"
	whereValue := make([]interface{}, 0)
	if req.DealStatus != nil {
		whereCase = whereCase + " and deal_status = ?"
		whereValue = append(whereValue, req.DealStatus)
	}
	if req.TaskId != nil {
		whereCase = whereCase + " and task_id = ?"
		whereValue = append(whereValue, req.TaskId)
	}
	if req.EventLevel != nil {
		whereCase = whereCase + " and event_level = ?"
		whereValue = append(whereValue, *req.EventLevel)
	}
	// 创建时间区间：解析失败则忽略该条件，保证容错
	if req.StartTime != nil && *req.StartTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", *req.StartTime, time.Local); err == nil {
			whereCase = whereCase + " and created_at >= ?"
			whereValue = append(whereValue, t)
		}
	}
	if req.EndTime != nil && *req.EndTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", *req.EndTime, time.Local); err == nil {
			whereCase = whereCase + " and created_at < ?"
			whereValue = append(whereValue, t)
		}
	}

	notCase := &entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	return paginate[entity.MonitorTaskEvent](repo.db, &entity.MonitorTaskEvent{}, whereCase, whereValue, notCase, req.PageSize, req.Offset(), "id desc")
}

// GetById 按主键查询
func (repo *MonitorTaskEventRepositoryImpl) GetById(id int64) (*entity.MonitorTaskEvent, error) {
	var data entity.MonitorTaskEvent
	err := repo.db.Model(&entity.MonitorTaskEvent{}).
		Where(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Id: id}}).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &data, err
}

// DealEvent 事件处理中：
// 1. 查出该 task 下全部 Pending 事件；
// 2. 若多于 1 条，除当前 eventId 外其余批量置 Ignore 并设 CompleteTime；
// 3. 当前事件置 Processing（DealTime/DealUser）；
// 4. 对应 alert 置 Processing。
func (repo *MonitorTaskEventRepositoryImpl) DealEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 查询该 task 下全部 Pending 事件
		var pendingEvents []entity.MonitorTaskEvent
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("task_id = ? and deal_status = ?", req.TaskId, entity.MonitorTaskEventDealStatusPending).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Find(&pendingEvents).Error; err != nil {
			return err
		}

		// 多余 Pending 事件置为 Ignore
		if len(pendingEvents) > 1 {
			ignoreIds := make([]int64, 0)
			for _, item := range pendingEvents {
				if item.Id != eventId {
					ignoreIds = append(ignoreIds, item.Id)
				}
			}
			if len(ignoreIds) > 0 {
				if err := tx.Model(&entity.MonitorTaskEvent{}).
					Where("id in (?)", ignoreIds).
					Updates(&entity.MonitorTaskEvent{
						DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
						CompleteTime: &common.LocalTime{Time: time.Now()},
					}).Error; err != nil {
					return err
				}
			}
		}

		// 当前事件置 Processing
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

		// 对应 alert 置 Processing
		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("task_id = ? and deal_status = ?", req.TaskId, entity.MonitorTaskAlertDealStatusNormal).
			Updates(&entity.MonitorTaskAlert{DealStatus: entity.MonitorTaskAlertDealStatusProcessing}).Error
	})
}

// CompleteEvent 事件完成：
// 更新 Content/CompleteTime/DealStatus=Complete；
// 同时把 task 对应 alert 的 DealStatus 恢复 Normal、AlertStatus Normal、刷新 FirstFlagTime。
func (repo *MonitorTaskEventRepositoryImpl) CompleteEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("task_id = ? and id = ? and deal_status = ?", req.TaskId, eventId, entity.MonitorTaskEventDealStatusProcessing).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Updates(&entity.MonitorTaskEvent{
				Content:      req.Content,
				DealStatus:   entity.MonitorTaskEventDealStatusComplete,
				CompleteTime: &common.LocalTime{Time: time.Now()},
			}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.MonitorTaskAlert{}).
			Where("task_id = ? and deal_status = ?", req.TaskId, entity.MonitorTaskAlertDealStatusProcessing).
			Updates(&entity.MonitorTaskAlert{
				DealStatus:    entity.MonitorTaskAlertDealStatusNormal,
				AlertStatus:   entity.MonitorTaskAlertStatusNormal,
				FirstFlagTime: &common.LocalTime{Time: time.Now()},
			}).Error
	})
}

// IgnoreEvent 忽略事件：
//  1. 仅当前事件置 Ignore（设 CompleteTime/DealUser），不动同任务其他 Pending 事件；
//  2. 若该 alert 下再无 Pending 事件，恢复 alert 为 Normal（AlertStatus/FirstFlagTime/PreCheckTime），
//     使下一轮检测从 Pending 重新计时，避免刚忽略的告警被立刻重新创建。
func (repo *MonitorTaskEventRepositoryImpl) IgnoreEvent(eventId int64, req *types.MonitorTaskEventProcessRequest) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		// 当前事件置 Ignore（仅待处理状态可忽略）
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("task_id = ? and id = ? and deal_status = ?", req.TaskId, eventId, entity.MonitorTaskEventDealStatusPending).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Updates(&entity.MonitorTaskEvent{
				DealStatus:   entity.MonitorTaskEventDealStatusIgnore,
				CompleteTime: &common.LocalTime{Time: time.Now()},
				DealUser:     req.DealUser,
			}).Error; err != nil {
			return err
		}

		// 该 alert 下是否还有剩余 Pending 事件（排除当前已置 Ignore 这条）
		var pendingCount int64
		if err := tx.Model(&entity.MonitorTaskEvent{}).
			Where("alert_id = ? and id <> ? and deal_status = ?", req.AlertId, eventId, entity.MonitorTaskEventDealStatusPending).
			Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
			Count(&pendingCount).Error; err != nil {
			return err
		}
		// 仍有待处理事件：仅忽略本条，alert 维持现状
		if pendingCount > 0 {
			return nil
		}
		// 已无待处理事件：恢复 alert 正常，下一轮从 Pending 重新计时
		now := time.Now()
		return tx.Model(&entity.MonitorTaskAlert{}).
			Where(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Id: req.AlertId}}).
			Updates(&entity.MonitorTaskAlert{
				AlertStatus:   entity.MonitorTaskAlertStatusNormal,
				FirstFlagTime: &common.LocalTime{Time: now},
				PreCheckTime:  &common.LocalTime{Time: now},
			}).Error
	})
}

// Count 统计总数
func (repo *MonitorTaskEventRepositoryImpl) Count() (*int64, error) {
	var count int64
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Count(&count).Error
	return &count, err
}

// CountByStatus 按处理状态统计事件数
func (repo *MonitorTaskEventRepositoryImpl) CountByStatus() (map[entity.MonitorTaskEventDealStatus]int64, error) {
	type statusCount struct {
		DealStatus entity.MonitorTaskEventDealStatus
		Count      int64
	}
	var results []statusCount
	if err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Not(&entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Select("deal_status, COUNT(*) as count").
		Group("deal_status").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	m := make(map[entity.MonitorTaskEventDealStatus]int64)
	for _, r := range results {
		m[r.DealStatus] = r.Count
	}
	return m, nil
}

// CountByLevel 按告警级别统计事件数
func (repo *MonitorTaskEventRepositoryImpl) CountByLevel() (map[int32]int64, error) {
	type levelCount struct {
		EventLevel int32
		Count      int64
	}
	var results []levelCount
	notCase := &entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	if err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Not(notCase).
		Select("event_level, COUNT(*) as count").
		Group("event_level").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	m := make(map[int32]int64)
	for _, r := range results {
		m[r.EventLevel] = r.Count
	}
	return m, nil
}

// SelectRecent 查询最近 N 条事件（按创建时间倒序）
func (repo *MonitorTaskEventRepositoryImpl) SelectRecent(limit int) ([]entity.MonitorTaskEvent, error) {
	var data []entity.MonitorTaskEvent
	notCase := &entity.MonitorTaskEvent{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}
	err := repo.db.
		Model(&entity.MonitorTaskEvent{}).
		Not(notCase).
		Order("created_at desc").
		Limit(limit).
		Find(&data).Error
	return data, err
}
