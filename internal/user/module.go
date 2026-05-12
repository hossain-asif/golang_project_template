package user

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"
	repositories "go_project_structure/internal/database/repositories/user"
	"go_project_structure/internal/module"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserModule struct {
	repo repositories.UserRepository
	svc  UserService
	fs   *storage.FileStore
	once sync.Once
}

// Init wires all user dependencies.
// Safe to call multiple times — only executes once due to sync.Once.
// Returns an error if called with a nil db.
func (m *UserModule) InitDependency(dependency module.Dependency) error {
	var initErr error

	m.once.Do(func() {
		if dependency.DB == nil {
			initErr = fmt.Errorf("user module: Init called with nil db")
			return
		}
		m.fs = dependency.FS
		m.repo = repositories.NewUserRepository(dependency.DB)
		m.svc = NewUserService(m.repo)
	})
	return initErr
}

func (m *UserModule) RegisterRoutes(r chi.Router) {
	if m.svc == nil || m.repo == nil {
		panic("user module: RegisterRoutes called before Init")
	}
	uc := NewUserController(m.svc, m.fs)
	NewUserRouter(uc).Register(r)
}

func (m *UserModule) RegisterTasks() []scheduler.Task {
	if m.svc == nil || m.repo == nil {
		panic("user module: RegisterTasks called before InitUserDependency")
	}

	return []scheduler.Task{
		{
			Name:     "user.sync-all",
			Interval: 24 * time.Hour,
			Fn: func(ctx context.Context) error {
				_, err := m.repo.GetAll(ctx)
				return err
			},
		},
		{
			Name:     "user.auto-export-csv",
			Interval: 50 * time.Minute,
			Fn: func(ctx context.Context) error {
				_, err := m.svc.ExportUsersAsCSV(ctx)
				return err
			},
		},
		{
			Name:     "user.file-rebuild",
			Interval: 59 * time.Minute,
			Fn: func(ctx context.Context) error {
				return m.fs.RebuildIfChecksumChanged()
			},
		},
	}
}
