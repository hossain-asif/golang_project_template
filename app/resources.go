package app

import (
	"fmt"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/config/resources"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// persistence database
func SetupDB() (*gorm.DB, error) {
	db, err := resources.SetupDB()
	if err != nil {
		appLog.Method("setupDB").WithError(err).Error("Error setting up database.")
		return nil, fmt.Errorf("database setup: %w", err)
	}
	return db, nil
}

// no-sql database for logger
func setupMongoHook() (*resources.MongoDB, error) {
	hook, err := resources.SetupMongoDB()
	if err != nil {
		appLog.Method("setupMongoHook").WithError(err).Error("Failed to connect to MongoDB log hook.")
		return nil, fmt.Errorf("mongo hook setup: %w", err)
	}
	logger.AddHook(hook)
	return hook, nil
}

// non-persistence database
func SetupRedis() (*redis.Client, error) {
	db, err := resources.SetupRedis()
	if err != nil {
		appLog.Method("SetupRedis").WithError(err).Error("Error setting up redis.")
		return nil, fmt.Errorf("redis setup: %w", err)
	}
	return db, nil
}
