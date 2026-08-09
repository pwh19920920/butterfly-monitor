package repository

import (
	"butterfly-monitor/internal/domain/entity"
)

type SysTokenRepository interface {
	// Save 保存
	Save(token entity.SysToken) error

	// ModifyById 更新
	ModifyById(token entity.SysToken, id int64) error

	// Delete 删除
	Delete(subject string) error

	// GetBySubject 查询
	GetBySubject(subject string) (*entity.SysToken, error)
}
