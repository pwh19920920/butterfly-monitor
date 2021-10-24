package persistence

import (
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/domain/repository"
)

type Repository struct {
	MonitorTaskRepository      repository.MonitorTaskRepository
	MonitorDatabaseRepository  repository.MonitorDatabaseRepository
	MonitorDashboardRepository repository.MonitorDashboardRepository
}

func NewRepository(config config.Config) *Repository {
	return &Repository{
		MonitorTaskRepository:      NewMonitorTaskRepositoryImpl(config.DatabaseForGorm),
		MonitorDatabaseRepository:  NewMonitorDatabaseRepositoryImpl(config.DatabaseForGorm),
		MonitorDashboardRepository: NewMonitorDashboardRepositoryImpl(config.DatabaseForGorm),
	}
}
