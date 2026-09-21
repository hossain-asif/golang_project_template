package example

import (
	"context"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/internal/repository/example"

	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	repository example.Repository
	service    Service
}

func NewModule(db *gorm.DB) *Module {
	repository := example.NewRepository(db)
	service := NewService(repository)

	return &Module{
		repository: repository,
		service:    service,
	}
}

func (um *Module) Initialize(r chi.Router) ([]scheduler.Task, error) {
	handler := NewHandler(um.service)
	router := NewRouter(handler)
	router.Register(r)

	return []scheduler.Task{
		{
			Name:     "example.",
			Interval: 24 * time.Hour,
			Fn: func(ctx context.Context) error {
				err := um.repository.Get(ctx)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:     "example.auto-export-csv",
			Interval: 50 * time.Minute,
			Fn: func(ctx context.Context) error {
				err := um.service.Get(ctx)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}, nil

}
