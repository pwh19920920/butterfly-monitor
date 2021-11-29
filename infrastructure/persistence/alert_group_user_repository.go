package persistence

import (
	"butterfly-monitor/domain/entity"
	"gorm.io/gorm"
)

type AlertGroupUserRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertGroupUserRepositoryImpl(db *gorm.DB) *AlertGroupUserRepositoryImpl {
	return &AlertGroupUserRepositoryImpl{db: db}
}

// SelectByGroupId 查询全部
func (repo *AlertGroupUserRepositoryImpl) SelectByGroupId(groupId int64) ([]int64, error) {
	var data []int64
	err := repo.db.Model(&entity.AlertGroupUser{}).
		Select("user_id").
		Where("group_id = ?", groupId).
		Find(&data).Error
	return data, err
}
