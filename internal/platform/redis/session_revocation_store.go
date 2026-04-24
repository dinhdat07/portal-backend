package redisx

import (
	"context"
	"fmt"
	"portal-system/internal/services"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisSessionRevocationStore struct {
	rdb *redis.Client
}

func NewRedisSessionRevocationStore(rdb *redis.Client) services.SessionRevocationStore {
	return &RedisSessionRevocationStore{rdb: rdb}
}

func (rs *RedisSessionRevocationStore) MarkRevoked(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	if sessionID == uuid.Nil {
		return nil
	}

	key := fmt.Sprintf("revoked_session:%s", sessionID.String())
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	return rs.rdb.Set(ctx, key, 1, ttl).Err()
}

func (rs *RedisSessionRevocationStore) IsRevoked(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	if sessionID == uuid.Nil {
		return false, nil
	}

	key := fmt.Sprintf("revoked_session:%s", sessionID.String())
	n, err := rs.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
