package di

import (
	"go_project_structure/config/resources"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB          *gorm.DB
	RedisClient *redis.Client
}

func NewDependencies(db *gorm.DB, redisClient *redis.Client) Dependencies {
	return Dependencies{
		DB:          db,
		RedisClient: redisClient,
	}
}

func LoadDependencies() (Dependencies, error) {
	// db setup
	db, err := resources.SetupDB()
	if err != nil {
		return Dependencies{}, err
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
		return Dependencies{}, err
	}

	dep := Dependencies{
		DB:          db,
		RedisClient: redisClient,
	}
	return dep, nil
}
