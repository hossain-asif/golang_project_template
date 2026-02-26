package user

import (
	"context"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/internal/db/repositories"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type UserModule struct{
    repo       repositories.UserRepository  // data layer
    svc        UserService                  // business layer
}

func (m *UserModule) initUserDependency(db *gorm.DB) {
    m.repo       = repositories.NewUserRepository(db)
    m.svc        = NewUserService(m.repo)
}

func (m *UserModule) RegisterRoutes(db *gorm.DB, r chi.Router) {
    m.initUserDependency(db)
	uc := NewUserController(m.svc) // local variable, not stored
    NewUserRouter(uc).Register(r)
}
func (m *UserModule) RegisterTasks(db *gorm.DB) []scheduler.Task {
	return []scheduler.Task{
		{
			Name: "get-all-users", 
			Interval: 24 * time.Hour, 
			Fn: func(ctx context.Context) error {
				_, err := m.repo.GetAll()
				return err
			},
		},
		{
			Name: "auto-export-csv",
			Interval: 50 * time.Minute,
			Fn: func(ctx context.Context) error {
				_, err := m.svc.ExportUsersAsCSV()
				return err
			},	
		},
	}
}
