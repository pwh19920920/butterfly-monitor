package persistence

import (
	"butterfly-monitor/domain/entity"
	"github.com/pwh19920920/butterfly-admin/common"
	"gorm.io/gorm"
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
func (repo *MonitorTaskAlertRepositoryImpl) Modify(id int64, MonitorTaskAlert *entity.MonitorTaskAlert) error {
	return repo.db.Model(&entity.MonitorTaskAlert{}).
		Where(&entity.MonitorTaskAlert{BaseEntity: common.BaseEntity{Id: id}}).
		Updates(&MonitorTaskAlert).Error
}
