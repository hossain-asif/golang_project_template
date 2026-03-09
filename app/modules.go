package app

import (
	"fmt"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"
	"go_project_structure/internal/module"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// bootstrapModules runs the full Init → RegisterRoutes → RegisterTasks
// lifecycle on every module, then returns the assembled router and task list.
func BootstrapModules(
	modules []module.Module,
	db *gorm.DB,
	fs *storage.FileStore,
) (*chi.Mux, []scheduler.Task, error) {

	rootRouter := chi.NewRouter()
	var allTasks []scheduler.Task

	for _, m := range modules {
		// Step 1 — wire dependencies
		if err := m.InitDependency(db, fs); err != nil {
			appLog.Method("bootstrapModules").WithError(err).
				Errorf("Module InitUserDependency failed: %T", m)
			return nil, nil, fmt.Errorf("module init (%T): %w", m, err)
		}

		// Step 2 — mount routes 
		m.RegisterRoutes(rootRouter)

		// Step 3 — collect tasks
		allTasks = append(allTasks, m.RegisterTasks()...)
	}

	return rootRouter, allTasks, nil
}
