package module

import (
	"go_project_structure/common_pkg/scheduler"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependency struct {
	DB          *gorm.DB
	RedisClient *redis.Client

	// add new infra here only
	// Redis *redis.Client
}

type Module interface {
	Initialize(dependencies Dependency, r chi.Router) ([]scheduler.Task, error)
}
