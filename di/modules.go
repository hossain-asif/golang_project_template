package di

import (
	"fmt"
	"go_project_structure/common_pkg/scheduler"

	"github.com/go-chi/chi/v5"
)

type Module interface {
	Initialize(r chi.Router) ([]scheduler.Task, error)
}

type Modules struct {
	modules []Module
}

func NewModules(modules []Module) *Modules {
	return &Modules{
		modules: modules,
	}
}

// dependencyInit runs the full Init → RegisterRoutes → RegisterTasks
// lifecycle on every module, then returns the assembled router and task list.
func (m *Modules) SetupModules(dependencies Dependencies) (*chi.Mux, []scheduler.Task, error) {

	if dependencies.DB == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil db")
	}

	if dependencies.RedisClient == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil redisCLient")
	}

	router := chi.NewRouter()
	var tasks []scheduler.Task

	for _, module := range m.modules {

		moduleTasks, err := module.Initialize(router)
		if err != nil {
			return nil, nil, fmt.Errorf("module initialize (%T): %w", module, err)
		}

		tasks = append(tasks, moduleTasks...)
	}

	return router, tasks, nil
}
