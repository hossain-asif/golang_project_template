package module

import (
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Dependency struct {
	DB *gorm.DB
	FS *storage.FileStore

	// add new infra here only
	// Redis *redis.Client
}

type Module interface {
	InitDependency(dependency Dependency) error
	RegisterRoutes(r chi.Router)
	RegisterTasks() []scheduler.Task
}
