package dependency

import (
	"go_project_structure/config/resources"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependency struct {
	DB          *gorm.DB
	RedisClient *redis.Client

	// add new infra here only
	// Redis *redis.Client
}

func NewDependency(db *gorm.DB, redisClient *redis.Client) Dependency {
	return Dependency{
		DB:          db,
		RedisClient: redisClient,
	}
}

func LoadDependency() (Dependency, error) {
	// db setup
	db, err := resources.SetupDB()
	if err != nil {
		return Dependency{}, err
	}

	// Connect MongoDB as a logrus hook (uncomment when needed)
	// mongoHook, err := setupMongoHook()
	// if err != nil {
	// 	return err
	// }
	// defer mongoHook.Disconnect()

	// redis setup
	redisClient, err := resources.SetupRedis()
	if err != nil {
		return Dependency{}, err
	}

	dep := Dependency{
		DB:          db,
		RedisClient: redisClient,
	}
	return dep, nil
}
