package module

import (
	"go_project_structure/common_pkg/scheduler"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module interface {
    RegisterRoutes(db *gorm.DB, r chi.Router)
    RegisterTasks(db *gorm.DB) []scheduler.Task
}
