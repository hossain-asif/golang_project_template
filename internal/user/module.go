package user

import (
	"context"
	"go_project_structure/common_pkg/scheduler"
	userrepo "go_project_structure/internal/db/repositories/user"
	"go_project_structure/internal/pkg/module"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserModule struct {
	repository userrepo.UserRepository
	service    UserService
}

func (um *UserModule) Initialize(dependency module.Dependency, r chi.Router) ([]scheduler.Task, error) {
	um.repository = userrepo.NewUserRepository(dependency.DB)
	um.service = NewUserService(um.repository)
	handler := NewUserHandler(um.service)
	router := NewUserRouter(handler)
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
				_, err := um.service.ExportUsersAsCSV(ctx)
				return err
			},
		},
	}, nil

}
