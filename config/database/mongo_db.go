package database

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Package-scoped logger — defined once, reused across all methods.
var mongoLog = logger.Log.Scope("config", "database", "mongo_database")

// mongoConfig holds all values needed to connect to MongoDB.
type mongoConfig struct {
	URI            string
	DBName         string
	CollectionName string
}

// loadMongoConfig reads all MongoDB connection values from the environment.
func loadMongoConfig() mongoConfig {
	return mongoConfig{
		URI:            env.GetString("MONGO_URI", "mongodb://127.0.0.1:27017"),
		DBName:         env.GetString("MONGO_DB_NAME", "logdb"),
		CollectionName: env.GetString("MONGO_COLLECTION_NAME", "logs"),
	}
}

// MongoDB wraps a mongo.Collection and implements logrus.Hook,
// allowing log entries to be persisted directly to MongoDB.
type MongoDB struct {
	client     *mongo.Client // retained so the caller can Disconnect on shutdown
	Collection *mongo.Collection
}

// SetupMongoDB connects to MongoDB, pings the server to verify connectivity,
// and returns a *MongoDB ready to be used as a logrus hook.
//
// The caller is responsible for calling client.Disconnect when done.
func SetupMongoDB() (*MongoDB, error) {
	log := mongoLog.Method("SetupMongoDB")

	cfg := loadMongoConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		log.WithError(err).Error("Failed to connect to MongoDB.")
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	// Ping to confirm the server is reachable before returning.
	if err = client.Ping(ctx, nil); err != nil {
		log.WithError(err).Error("MongoDB ping failed — server may be unreachable.")
		// Disconnect the dangling client before returning the error.
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	col := client.Database(cfg.DBName).Collection(cfg.CollectionName)

	log.Infof("Successfully connected to MongoDB. db: %s, collection: %s", cfg.DBName, cfg.CollectionName)

	return &MongoDB{
		client:     client,
		Collection: col,
	}, nil
}

// Disconnect cleanly closes the MongoDB connection.
// Should be called via defer in the caller (e.g. application.go).
func (m *MongoDB) Disconnect() {
	if m.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.client.Disconnect(ctx); err != nil {
		mongoLog.Method("Disconnect").WithError(err).Warn("MongoDB disconnect encountered an error.")
	}
}

// Fire implements logrus.Hook — called on every log entry.
// Inserts the log entry as a document into the configured MongoDB collection.
func (m *MongoDB) Fire(entry *logrus.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := map[string]interface{}{
		"time":    entry.Time,
		"level":   entry.Level.String(),
		"message": entry.Message,
		"fields":  entry.Data,
	}

	if _, err := m.Collection.InsertOne(ctx, doc); err != nil {
		// Log the failure but do NOT return the error —
		// a logging hook must never crash the application.
		mongoLog.Method("Fire").WithError(err).Warn("Failed to persist log entry to MongoDB.")
	}

	return nil
}

// Levels implements logrus.Hook — specifies which log levels trigger Fire.
func (m *MongoDB) Levels() []logrus.Level {
	return logrus.AllLevels
}
