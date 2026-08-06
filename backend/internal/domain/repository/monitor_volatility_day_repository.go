package repository

import "dragonfly-monitor/internal/domain/entity"

type MonitorVolatilityDayRepository interface {
	SelectAll() ([]entity.MonitorVolatilityDay, error)
	Save(day *entity.MonitorVolatilityDay) error
	BatchSave(days []entity.MonitorVolatilityDay) error
	Modify(id int64, day *entity.MonitorVolatilityDay) error
	Delete(id int64) error
}
