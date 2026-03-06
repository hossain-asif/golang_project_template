package module

import (
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage/file_system"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module interface {
	RegisterRoutes(db *gorm.DB, r chi.Router, fs *file_system.FileStore)
	RegisterTasks(db *gorm.DB, fs *file_system.FileStore) []scheduler.Task
}
