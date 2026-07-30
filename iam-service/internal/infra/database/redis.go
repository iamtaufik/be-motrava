package database

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr, password string, db int, logger *slog.Logger) (*redis.Client, error) {
	logger.Info("connecting to redis", "module", "redis", "addr", addr, "db", db)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error("redis connection failed", "module", "redis", "error", err)
		return nil, err
	}

	logger.Info("redis connected", "module", "redis")
	return client, nil
}
