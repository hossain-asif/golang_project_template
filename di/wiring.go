package di

import (
	"context"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/internal/example"
	"go_project_structure/internal/identity"

	"github.com/go-chi/chi/v5"
)

func BuildModules(deps Dependencies) []Module {
	return []Module{
		identity.NewModule(deps.DB),
		example.NewModule(deps.DB),
	}
}

func BuildApplication(ctx context.Context) (*chi.Mux, error) {

	dependencies, err := LoadDependencies()
	if err != nil {
		return nil, err
	}

	mods := BuildModules(dependencies)

	modules := NewModules(mods)

	rootRouter, allTasks, err := modules.SetupModules(dependencies)
	if err != nil {
		return nil, err
	}

	go scheduler.TaskAssignment(ctx, allTasks)

	return rootRouter, nil
}
