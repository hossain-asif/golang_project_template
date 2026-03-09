package module

import (
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module interface {
	RegisterRoutes(db *gorm.DB, r chi.Router, fs *storage.FileStore)
	RegisterTasks(db *gorm.DB, fs *storage.FileStore) []scheduler.Task
}
