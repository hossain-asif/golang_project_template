package example

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/db/models"

	"gorm.io/gorm"
)

type Repository interface {
	Get(ctx context.Context) error
}

type RepositoryImpl struct {
	db  *gorm.DB
	Log *logger.GormLogWriter
}

func NewRepository(_db *gorm.DB) Repository {
	return &RepositoryImpl{
		db:  _db,
		Log: &logger.GormLogWriter{Logger: logger.Log.Scope("repository", "gorm", "example_repository")},
	}
}

func (u *RepositoryImpl) Get(ctx context.Context) error {
	log := u.Log.Method("Get").WithContext(ctx)

	result := u.db.First(&models.Example{})

	if result.Error != nil {
		log.Errorf("Error getting Example: %v\n", result.Error)
		return result.Error
	}

	return nil
}
