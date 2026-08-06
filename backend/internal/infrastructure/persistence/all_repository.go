package persistence

import (
	"dragonfly-monitor/internal/config"
	"dragonfly-monitor/internal/domain/repository"
)

type Repository struct {
	SysUserRepository              repository.SysUserRepository
	SysTokenRepository             repository.SysTokenRepository
	SysMenuRepository              repository.SysMenuRepository
	SysRoleRepository              repository.SysRoleRepository
	SysPermissionRepository        repository.SysPermissionRepository
	SysMenuOptionRepository        repository.SysMenuOptionRepository
	MonitorTaskRepository          repository.MonitorTaskRepository
	MonitorDatabaseRepository      repository.MonitorDatabaseRepository
	MonitorDashboardRepository     repository.MonitorDashboardRepository
	MonitorDashboardTaskRepository repository.MonitorDashboardTaskRepository
	MonitorGroupRepository         repository.MonitorGroupRepository
	MonitorTaskAlertRepository     repository.MonitorTaskAlertRepository
	MonitorTaskEventRepository     repository.MonitorTaskEventRepository
	MonitorVolatilityDayRepository repository.MonitorVolatilityDayRepository
	AlertConfRepository            repository.AlertConfRepository
	AlertGroupRepository           repository.AlertGroupRepository
	AlertGroupUserRepository       repository.AlertGroupUserRepository
	AlertChannelRepository         repository.AlertChannelRepository
}

func NewRepository(config config.Config) *Repository {
	return &Repository{
		SysMenuOptionRepository:        NewSysMenuOptionRepositoryImpl(config.DatabaseForGorm),
		SysPermissionRepository:        NewSysPermissionRepositoryImpl(config.DatabaseForGorm),
		SysUserRepository:              NewSysUserRepositoryImpl(config.DatabaseForGorm),
		SysTokenRepository:             NewSysTokenRepositoryImpl(config.DatabaseForGorm),
		SysMenuRepository:              NewSysMenuRepositoryImpl(config.DatabaseForGorm),
		SysRoleRepository:              NewSysRoleRepositoryImpl(config.DatabaseForGorm),
		MonitorTaskRepository:          NewMonitorTaskRepositoryImpl(config.DatabaseForGorm),
		MonitorDatabaseRepository:      NewMonitorDatabaseRepositoryImpl(config.DatabaseForGorm),
		MonitorDashboardRepository:     NewMonitorDashboardRepositoryImpl(config.DatabaseForGorm),
		MonitorDashboardTaskRepository: NewMonitorDashboardTaskRepositoryImpl(config.DatabaseForGorm),
		MonitorGroupRepository:         NewMonitorGroupRepositoryImpl(config.DatabaseForGorm),
		MonitorTaskAlertRepository:     NewMonitorTaskAlertRepositoryImpl(config.DatabaseForGorm),
		MonitorTaskEventRepository:     NewMonitorTaskEventRepositoryImpl(config.DatabaseForGorm),
		MonitorVolatilityDayRepository: NewMonitorVolatilityDayRepositoryImpl(config.DatabaseForGorm),
		AlertConfRepository:            NewAlertConfRepositoryImpl(config.DatabaseForGorm),
		AlertGroupRepository:           NewAlertGroupRepositoryImpl(config.DatabaseForGorm),
		AlertGroupUserRepository:       NewAlertGroupUserRepositoryImpl(config.DatabaseForGorm),
		AlertChannelRepository:         NewAlertChannelRepositoryImpl(config.DatabaseForGorm),
	}
}
