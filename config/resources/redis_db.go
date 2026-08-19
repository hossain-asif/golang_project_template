package resources

import (
	"context"
	"fmt"
	env "go_project_structure/config/env"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         env.GetString("REDIS_ADDR", "127.0.0.1:6379"), // "localhost:6379"
		Password:     env.GetString("REDIS_PASSWORD", ""),      // "" when no auth
		DB:           env.GetInt("REDIS_DB", 0),                // 0
		PoolSize:     10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}
