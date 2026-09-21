package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

func NewRedisTokenRepository(client *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{client: client}
}

func tokenKey(tokenID string) string {
	return "auth:token:" + tokenID
}

func (s *RedisTokenRepository) Store(ctx context.Context, tokenID string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, tokenKey(tokenID), "1", ttl).Err()
}

func (s *RedisTokenRepository) Exists(ctx context.Context, tokenID string) (bool, error) {
	n, err := s.client.Exists(ctx, tokenKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *RedisTokenRepository) Delete(ctx context.Context, tokenID string) error {
	return s.client.Del(ctx, tokenKey(tokenID)).Err()
}
