package example

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/internal/db/repositories/example"
)

type Service interface {
	Get(ctx context.Context) error
}

type ServiceImpl struct {
	exampleRepository example.Repository
	Log               *logger.ScopeLogger
}

func NewService(_exampleRepository example.Repository) Service {
	return &ServiceImpl{
		exampleRepository: _exampleRepository,
		Log:               logger.Log.Scope("", "example", "example_service"),
	}
}

func (us *ServiceImpl) Get(ctx context.Context) error {
	log := us.Log.Method("Get").WithContext(ctx)

	err := us.exampleRepository.Get(ctx)
	if err != nil {
		log.Errorf("Error getting Example: %v\n", err)
		return err
	}

	return nil
}
