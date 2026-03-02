package database

import (
	"fmt"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// global decalaration

// Initialize the scoped logger once. Used "Scoped Logging" to reduce boilerplate.
// The "Define Once" Pattern (Package-Scoped)
var pglog = logger.Log.Scope("config", "database", "postgres_database")

func SetupDB() (*gorm.DB, error) {

	host := env.GetString("DB_HOST", "127.0.0.1")
	port := env.GetString("DB_PORT", "5432")
	user := env.GetString("DB_USER", "minhaz_hossain")
	password := env.GetString("DB_PASSWORD", "12345")
	dbname := env.GetString("DB_NAME", "auth_dev")
	sslmode := env.GetString("DB_SSLMODE", "disable")
	timezone := env.GetString("DB_TIMEZONE", "UTC")

	// Example of using these values to construct a connection string
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, password, dbname, port, sslmode, timezone,
	)

	// fmt.Println(dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// logger.Log.WithFields(map[string]interface{}{
		// 	"layer":     "config",
		// 	"module":    "database",
		// 	"component": "postgres_database",
		// 	"method":    "SetupDB",
		// 	"error":     err,
		// }).Error("Failed to connect to database")
		pglog.Method("SetupDB").WithError(err).Error("Failed to connect to database")

		return nil, err
	}

	pgsqlDB, err := db.DB()
	if err != nil {
		// logger.Log.WithFields(map[string]interface{}{
		// 	"layer":     "config",
		// 	"module":    "database",
		// 	"component": "postgres_database",
		// 	"method":    "SetupDB",
		// 	"error":     err,
		// }).Error("Failed to get database connection")
		pglog.Method("SetupDB").WithError(err).Error("Failed to get database connection")
		return nil, err
	}
	err = pgsqlDB.Ping()
	if err != nil {
		// logger.Log.WithFields(map[string]interface{}{
		// 	"layer":     "config",
		// 	"module":    "database",
		// 	"component": "postgres_database",
		// 	"method":    "SetupDB",
		// 	"error":     err,
		// }).Error("Failed to ping database")
		pglog.Method("SetupDB").WithError(err).Error("Failed to ping database")
		return nil, err
	}

	// logger.Log.WithFields(map[string]interface{}{
	// 	"layer":     "config",
	// 	"module":    "database",
	// 	"component": "postgres_database",
	// 	"method":    "SetupDB",
	// }).Info("Successfully connected to database")
	pglog.Method("SetupDB").Info("Successfully connected to database")

	var dbName string
	db.Raw("SELECT current_database()").Scan(&dbName)

	// logger.Log.WithFields(map[string]interface{}{
	// 	"layer":     "config",
	// 	"module":    "database",
	// 	"component": "postgres_database",
	// 	"method":    "SetupDB",
	// 	"database":  dbName,
	// }).Info("Connected to database")

	// log.Method("SetupDB").WithField("database", dbName).Info("Connected to database")

	pglog.Method("SetupDB").Infof("Connected to database: %s", dbName)


	return db, nil

}
