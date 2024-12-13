package config

import (
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var once sync.Once
var appConfig *AppConfig

type AppConfig struct {
	DbClient    *gorm.DB
	RedisClient *redis.Client
}

func NewAppConfig() *AppConfig {
	once.Do(func() {
		appConfig = &AppConfig{
			RedisClient: newRedisClient(),
		}
		db, err := newDBClient()
		if err != nil {
			log.Fatalln(err)
		}
		appConfig.DbClient = db
	})
	return appConfig
}
