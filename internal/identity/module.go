package identity

import (
	"context"
	"go_project_structure/common_pkg/scheduler"

	"go_project_structure/internal/db/repositories/user"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	repository user.Repository
	service    Service
}

func NewModule(db *gorm.DB) *Module {
	repository := user.NewRepository(db)
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
			Name:     "user.sync-all",
			Interval: 24 * time.Hour,
			Fn: func(ctx context.Context) error {
				_, err := um.repository.GetAll(ctx)
				return err
			},
		},
		{
			Name:     "user.auto-export-csv",
			Interval: 50 * time.Minute,
			Fn: func(ctx context.Context) error {
				_, err := um.service.GetAllUsers(ctx)
				return err
			},
		},
	}, nil

}
