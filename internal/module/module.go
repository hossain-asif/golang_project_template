package module

import (
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module interface {
	InitDependency(db *gorm.DB, fs *storage.FileStore) error
	RegisterRoutes(r chi.Router)
	RegisterTasks() []scheduler.Task
}
