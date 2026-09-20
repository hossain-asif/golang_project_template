package di

import (
	"context"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/di/dependency"
	"go_project_structure/di/module"

	"github.com/go-chi/chi/v5"
)

func BuildApplicationModules(ctx context.Context) (*chi.Mux, error) {
	modules := module.NewDomainModules(module.BuildModules())

	dep, err := dependency.LoadDependency()
	if err != nil {
		return nil, err
	}

	rootRouter, allTasks, err := modules.SetupDomainModules(dep)
	if err != nil {
		return nil, err
	}

	go scheduler.TaskAssignment(ctx, allTasks)

	return rootRouter, nil
}
