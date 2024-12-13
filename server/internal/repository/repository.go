package repository

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/the-arcade-01/go-chat-app/server/internal/config"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"github.com/the-arcade-01/go-chat-app/server/internal/utils"
	"golang.org/x/crypto/bcrypt"
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

func (repo *Repository) RegisterUser(user *models.User) (string, int, error) {
	var existingUser models.User
	if err := repo.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return "", http.StatusConflict, fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to hash password")
	}
	user.Password = string(hashedPassword)

	if err := repo.db.Create(&user).Error; err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to create user")
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error generating token")
	}

	return token, http.StatusCreated, nil
}

func (repo *Repository) LoginUser(user *models.User) (string, int, error) {
	var existingUser models.User
	if err := repo.db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
		return "", http.StatusUnauthorized, fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
		return "", http.StatusUnauthorized, fmt.Errorf("invalid username or password")
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error generating token")
	}

	return token, http.StatusOK, nil
}

/* Below functions need to commented out */

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
