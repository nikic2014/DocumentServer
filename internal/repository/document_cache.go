package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type DocumentCache struct {
	client *redis.Client
}

func NewDocumentCache(client *redis.Client) *DocumentCache {
	return &DocumentCache{client: client}
}

func documentCacheKey(id int64) string {
	return fmt.Sprintf("docs:doc:%d", id)
}

func listCacheRedisKey(ownerLogin, key string) string {
	return fmt.Sprintf("docs:list:%s:%s", ownerLogin, key)
}

func ownerIndexKey(ownerLogin string) string {
	return fmt.Sprintf("docs:owner-index:%s", ownerLogin)
}

func (c *DocumentCache) GetDocument(ctx context.Context, id int64) ([]byte, bool, error) {
	return c.get(ctx, documentCacheKey(id))
}

func (c *DocumentCache) SetDocument(ctx context.Context, id int64, ownerLogin string, data []byte, ttl time.Duration) error {
	return c.set(ctx, ownerLogin, documentCacheKey(id), data, ttl)
}

func (c *DocumentCache) GetList(ctx context.Context, ownerLogin, key string) ([]byte, bool, error) {
	return c.get(ctx, listCacheRedisKey(ownerLogin, key))
}

func (c *DocumentCache) SetList(ctx context.Context, ownerLogin, key string, data []byte, ttl time.Duration) error {
	return c.set(ctx, ownerLogin, listCacheRedisKey(ownerLogin, key), data, ttl)
}

func (c *DocumentCache) get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return val, true, nil
}

func (c *DocumentCache) set(ctx context.Context, ownerLogin, key string, data []byte, ttl time.Duration) error {
	pipe := c.client.TxPipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, ownerIndexKey(ownerLogin), key)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *DocumentCache) InvalidateOwner(ctx context.Context, ownerLogin string) error {
	indexKey := ownerIndexKey(ownerLogin)

	keys, err := c.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	return c.client.Del(ctx, indexKey).Err()
}
