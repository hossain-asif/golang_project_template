package module

import (
	"fmt"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/di/dependency"

	"github.com/go-chi/chi/v5"
)

type Module interface {
	Initialize(dependencies dependency.Dependency, r chi.Router) ([]scheduler.Task, error)
}

type DomainModules struct {
	modules []Module
}

func NewDomainModules(modules []Module) *DomainModules {
	return &DomainModules{
		modules: modules,
	}
}

// dependencyInit runs the full Init → RegisterRoutes → RegisterTasks
// lifecycle on every module, then returns the assembled router and task list.
func (domain *DomainModules) SetupDomainModules(dependency dependency.Dependency) (*chi.Mux, []scheduler.Task, error) {

	if dependency.DB == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil db")
	}

	if dependency.RedisClient == nil {
		return nil, nil, fmt.Errorf("dependencyInit: nil redisCLient")
	}

	rootRouter := chi.NewRouter()
	var scheduledTasks []scheduler.Task

	for _, module := range domain.modules {

		tasks, err := module.Initialize(dependency, rootRouter)
		if err != nil {
			return nil, nil, fmt.Errorf("module initialize (%T): %w", module, err)
		}

		scheduledTasks = append(scheduledTasks, tasks...)
	}

	return rootRouter, scheduledTasks, nil
}
