package persistence

import (
	"butterfly-monitor/src/app/config"
	"butterfly-monitor/src/app/domain/repository"
)

type Repository struct {
	JobTaskRepository     repository.JobTaskRepository
	JobDatabaseRepository repository.JobDatabaseRepository
}

func NewRepository(config config.Config) *Repository {
	return &Repository{
		JobTaskRepository:     NewJobTaskRepositoryImpl(config.DatabaseForGorm),
		JobDatabaseRepository: NewJobDatabaseRepositoryImpl(config.DatabaseForGorm),
	}
}
