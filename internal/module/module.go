package module

import (
	"go_project_structure/common_pkg/scheduler"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Dependency struct {
	DB *gorm.DB

	// add new infra here only
	// Redis *redis.Client
}

type Module interface {
	InitDependency(dependency Dependency) error
	RegisterRoutes(r chi.Router)
	RegisterTasks() []scheduler.Task
}
