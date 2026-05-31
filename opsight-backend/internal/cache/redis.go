package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"opsight-backend/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() *redis.Client {
	host := envOrDefault("REDIS_HOST", "redis")
	port := envOrDefault("REDIS_PORT", "6379")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn().Err(err).Str("addr", rdb.Options().Addr).Msg("Redis unavailable, running without cache")
		return nil
	}

	RDB = rdb
	logger.Info().Str("addr", rdb.Options().Addr).Msg("Connected to Redis")
	return rdb
}

func Get(ctx context.Context, key string) (string, error) {
	if RDB == nil {
		return "", redis.Nil
	}
	return RDB.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if RDB == nil {
		return nil
	}
	return RDB.Set(ctx, key, value, expiration).Err()
}

func Del(ctx context.Context, keys ...string) error {
	if RDB == nil {
		return nil
	}
	return RDB.Del(ctx, keys...).Err()
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}