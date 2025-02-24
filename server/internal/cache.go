package internal

import (
	"context"
	"strings"

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

func (c *cacheRepo) publish(ctx context.Context, channel string, msg string) error {
	return c.redis.Publish(ctx, channel, msg).Err()
}

func (c *cacheRepo) initStream(ctx context.Context, streamKey string) {
	if err := c.redis.XGroupCreateMkStream(ctx, streamKey, Envs.STREAM_CONSUMER_GROUP, "$").Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			Log.ErrorContext(ctx, "error creating stream group", "error", err)
		}
	}
}

func (c *cacheRepo) writeToStream(ctx context.Context, streamKey, msg string) error {
	return c.redis.XAdd(ctx, &redis.XAddArgs{
		Stream:     streamKey,
		NoMkStream: true,
		MaxLen:     int64(Envs.MAX_CHAT_LEN),
		ID:         "*",
		Values: map[string]interface{}{
			"message": msg,
		},
	}).Err()
}

func (c *cacheRepo) getMessagesFromStream(ctx context.Context, streamKey string) ([]string, error) {
	msgs, err := c.redis.XRevRangeN(ctx, streamKey, "+", "-", int64(Envs.MAX_CHAT_LEN)).Result()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, msg := range msgs {
		result = append(result, msg.Values["message"].(string))
	}
	return result, nil
}
