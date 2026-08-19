package app

import (
	"fmt"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/internal/pkg/module"

	"github.com/go-chi/chi/v5"
)

// dependencyInit runs the full Init → RegisterRoutes → RegisterTasks
// lifecycle on every module, then returns the assembled router and task list.
func dependencyInit(modules []module.Module, dependency module.Dependency) (*chi.Mux, []scheduler.Task, error) {

	if dependency.DB == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil db")
	}

	if dependency.RedisClient == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil redisCLient")
	}

	rootRouter := chi.NewRouter()
	var scheduledTasks []scheduler.Task

	for _, m := range modules {

		tasks, err := m.Initialize(dependency, rootRouter)
		if err != nil {
			return nil, nil, fmt.Errorf("module initialize (%T): %w", m, err)
		}

		scheduledTasks = append(scheduledTasks, tasks...)
	}

	return rootRouter, scheduledTasks, nil
}
