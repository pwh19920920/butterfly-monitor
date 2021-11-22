package persistence

import (
	"butterfly-monitor/config"
	"butterfly-monitor/domain/repository"
)

type Repository struct {
	MonitorTaskRepository          repository.MonitorTaskRepository
	MonitorDatabaseRepository      repository.MonitorDatabaseRepository
	MonitorDashboardRepository     repository.MonitorDashboardRepository
	MonitorDashboardTaskRepository repository.MonitorDashboardTaskRepository
	AlertConfRepository            repository.AlertConfRepository
}

func NewRepository(config config.Config) *Repository {
	return &Repository{
		MonitorTaskRepository:          NewMonitorTaskRepositoryImpl(config.DatabaseForGorm),
		MonitorDatabaseRepository:      NewMonitorDatabaseRepositoryImpl(config.DatabaseForGorm),
		MonitorDashboardRepository:     NewMonitorDashboardRepositoryImpl(config.DatabaseForGorm),
		MonitorDashboardTaskRepository: NewMonitorDashboardTaskRepositoryImpl(config.DatabaseForGorm),
		AlertConfRepository:            NewAlertConfRepositoryImpl(config.DatabaseForGorm),
	}
}
