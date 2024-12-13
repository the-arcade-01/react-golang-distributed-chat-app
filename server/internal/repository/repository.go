package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/the-arcade-01/go-chat-app/server/internal/config"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository() *Repository {
	appConfig := config.NewAppConfig()
	return &Repository{
		db:    appConfig.DbClient,
		redis: appConfig.RedisClient,
	}
}

func (repo *Repository) SetValue(ctx context.Context, key, val string) error {
	status := repo.redis.Set(ctx, key, val, 0)
	cmd, err := status.Result()
	if err != nil {
		log.Printf("[SetValue] error on key: %v, val: %v, cmd: %v, err: %v\n", key, val, cmd, err)
		return fmt.Errorf("error on key: %v, val: %v, please try again with proper values", key, val)
	}
	return nil
}

func (repo *Repository) GetValue(ctx context.Context, key string) (string, error) {
	status := repo.redis.Get(ctx, key)
	cmd, err := status.Result()
	if err != nil {
		log.Printf("[GetValue] error on key: %v, cmd: %v, err: %v\n", key, cmd, err)
		return "", fmt.Errorf("error on key: %v, please try again with proper values", key)
	}
	return cmd, nil
}

func (repo *Repository) GetCount(ctx context.Context) (int64, error) {
	var count int64
	err := repo.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	if err != nil {
		log.Printf("[GetCount] error %v\n", err)
		return 0, fmt.Errorf("error occurred, please try again later")
	}
	return count, nil
}
