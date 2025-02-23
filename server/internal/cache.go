package internal

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type cacheRepo struct {
	redis *redis.Client
}

func newCacheRepo() *cacheRepo {
	return &cacheRepo{
		redis: NewAppConfig().cache,
	}
}

func (c *cacheRepo) exists(ctx context.Context, key string) bool {
	return c.redis.Exists(ctx, key).Val() > 0
}

func (c *cacheRepo) publish(ctx context.Context, channel string, msg string) error {
	return c.redis.Publish(ctx, channel, msg).Err()
}
