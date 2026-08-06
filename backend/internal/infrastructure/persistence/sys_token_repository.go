package persistence

import (
	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SysTokenRepositoryImpl struct {
	db *gorm.DB
}

func NewSysTokenRepositoryImpl(db *gorm.DB) *SysTokenRepositoryImpl {
	return &SysTokenRepositoryImpl{db: db}
}

func (tokenRepository *SysTokenRepositoryImpl) Save(token entity.SysToken) error {
	return tokenRepository.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "subject"}},
		DoUpdates: clause.AssignmentColumns([]string{"secret", "user_id", "expire_at"}),
	}).Create(&token).Error
}

func (tokenRepository *SysTokenRepositoryImpl) ModifyById(token entity.SysToken, id int64) error {
	return tokenRepository.db.Model(&entity.SysToken{}).Where(&entity.SysToken{BaseEntity: common.BaseEntity{Id: id}}).Updates(&token).Error
}

func (tokenRepository *SysTokenRepositoryImpl) Delete(subject string) error {
	return tokenRepository.db.Model(&entity.SysToken{}).
		Where(&entity.SysToken{Subject: subject}).
		Updates(&entity.SysToken{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).Error
}

func (tokenRepository *SysTokenRepositoryImpl) GetBySubject(subject string) (*entity.SysToken, error) {
	var token entity.SysToken
	err := tokenRepository.db.Model(&entity.SysToken{}).
		Where(&entity.SysToken{Subject: subject}).
		Not(&entity.SysToken{BaseEntity: common.BaseEntity{Deleted: common.DeletedTrue}}).
		Last(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}
