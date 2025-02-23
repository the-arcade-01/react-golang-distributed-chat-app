package internal

import (
	"log/slog"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	Log       = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	once      sync.Once
	appConfig *AppConfig
)

type AppConfig struct {
	cache *redis.Client
}

func NewAppConfig() *AppConfig {
	once.Do(func() {
		appConfig = &AppConfig{
			cache: newCacheClient(),
		}
	})
	return appConfig
}

func newCacheClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     Envs.REDIS_ADDR,
		Password: Envs.REDIS_PWD,
		DB:       Envs.REDIS_DB,
	})
}
