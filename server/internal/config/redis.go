package config

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PWD"),
		DB:       0,
	})
}
