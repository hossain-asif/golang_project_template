package database

import (
	"context"
	env "go_project_structure/config/env"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Collection *mongo.Collection
}

// Connect to MongoDB and return hook
func SetupMongoDB() (*MongoDB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := env.GetString("MONGO_URI", "127.0.0.1")
	dbName := env.GetString("MONGO_DB_NAME", "logdb")
	collectionName := env.GetString("MONGO_COLLECTION_NAME", "logs")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	col := client.Database(dbName).Collection(collectionName)
	return &MongoDB{Collection: col}, nil
}

// Fire is called on every log entry
func (m *MongoDB) Fire(entry *logrus.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := map[string]interface{}{
		"time":    entry.Time,
		"level":   entry.Level.String(),
		"message": entry.Message,
		"fields":  entry.Data,
	}

	_, err := m.Collection.InsertOne(ctx, doc)
	return err
}

// Levels specifies which log levels to send to MongoDB
func (m *MongoDB) Levels() []logrus.Level {
	return logrus.AllLevels
}
