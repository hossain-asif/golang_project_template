package database

import (
	"fmt"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"
	"os"
	"time"
	goLog "log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Package-scoped logger — defined once, reused across all methods.
var pglog = logger.Log.Scope("config", "database", "postgres_database")

// postgresConfig holds all values needed to build a DSN.
// Keeping them in a struct makes the setup testable and self-documenting.
type postgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// loadPostgresConfig reads all Postgres connection values from the environment.
func loadPostgresConfig() postgresConfig {
	return postgresConfig{
		Host:     env.GetString("DB_HOST", "127.0.0.1"),
		Port:     env.GetString("DB_PORT", "5432"),
		User:     env.GetString("DB_USER", "user"),
		Password: env.GetString("DB_PASSWORD", "12345"),
		DBName:   env.GetString("DB_NAME", "mydb"),
		SSLMode:  env.GetString("DB_SSLMODE", "disable"),
		TimeZone: env.GetString("DB_TIMEZONE", "UTC"),
	}
}

// dsn builds the PostgreSQL connection string from config.
func (c postgresConfig) dsn() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode, c.TimeZone,
	)
}

// SetupDB connects to PostgreSQL, verifies the connection with a ping,
// and returns a ready-to-use *gorm.DB instance.
func SetupDB() (*gorm.DB, error) {
	log := pglog.Method("SetupDB")

	// GORM logger
	// gormLog := logger.Log.Scope("repository", "gorm", "query")
	newLogger := gormlogger.New(
		// logger.GormLogWriter{Logger: gormLog},
		goLog.New(os.Stdout, "\r\n", goLog.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	cfg := loadPostgresConfig()

	db, err := gorm.Open(postgres.Open(cfg.dsn()), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.WithError(err).Error("Failed to open database connection.")
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Retrieve the underlying *sql.DB to verify connectivity.
	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get underlying sql.DB.")
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.WithError(err).Error("Database ping failed — server may be unreachable.")
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Confirm which database we actually connected to.
	var connectedDB string
	db.Raw("SELECT current_database()").Scan(&connectedDB)

	log.Infof("Successfully connected to database: %s", connectedDB)

	return db, nil
}
