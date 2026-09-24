package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitCache(cfg *Config) *redis.Client {
	var (
		addr     = cfg.Get("REDIS_ADDR", "localhost:6379")
		password = cfg.Get("REDIS_PASSWORD", "")
		db       = cfg.GetInt("REDIS_DB", 0)
		poolSize = cfg.GetInt("REDIS_POOL_SIZE", 10)
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	return client
}
