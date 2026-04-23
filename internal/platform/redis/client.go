package redisx

import (
	"context"

	"portal-system/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewClient(cfg *config.RedisConfig) *redis.Client {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	return redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
	})
}

func Ping(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	return rdb.Ping(ctx).Err()
}
