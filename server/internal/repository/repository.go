package repository

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

func (repo *Repository) GetAllUsers() ([]string, error) {
	var usernames []string
	if err := repo.db.Model(&models.User{}).Pluck("username", &usernames).Error; err != nil {
		log.Printf("[GetAllUsers] error %v\n", err)
		return nil, err
	}
	return usernames, nil
}

func (repo *Repository) CreateChatRoom(ctx context.Context, body *models.ChatRoomReqBody, username string) (*models.ChatRoom, int, error) {
	roomId := uuid.New().String()
	redisKey := fmt.Sprintf("%v:%v", roomId, body.RoomName)

	status := repo.redis.RPush(ctx, redisKey, username)
	if err := status.Err(); err != nil {
		log.Printf("[CreateChatRoom] error on key: %v, username: %v, err: %v\n", redisKey, username, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating chat room")
	}

	chatRoom := &models.ChatRoom{
		RoomId:   roomId,
		RoomName: body.RoomName,
	}

	return chatRoom, http.StatusCreated, nil
}

func (repo *Repository) AddUserToChatRoom(ctx context.Context, body *models.ChatRoomAddUserReqBody) (int, error) {
	redisKey := fmt.Sprintf("%v:%v", body.RoomId, body.RoomName)
	exists, err := repo.redis.Exists(ctx, redisKey).Result()

	if err != nil {
		log.Printf("[AddUserToChatRoom] error checking key existence: %v, err: %v\n", redisKey, err)
		return http.StatusInternalServerError, fmt.Errorf("error checking key existence")
	}
	if exists == 0 {
		log.Printf("[AddUserToChatRoom] key does not exist: %v\n", redisKey)
		return http.StatusBadRequest, fmt.Errorf("chat room does not exist")
	}

	status := repo.redis.RPush(ctx, redisKey, convertToInterfaceSlice(body.Users))
	if err := status.Err(); err != nil {
		log.Printf("[AddUserToChatRoom] error on key: %v, err: %v\n", redisKey, err)
		return http.StatusInternalServerError, fmt.Errorf("error adding user to chat room")
	}

	return http.StatusCreated, nil
}

func (repo *Repository) RemoveUserFromChatRoom(ctx context.Context, roomId, roomName, username string) (int, error) {
	redisKey := fmt.Sprintf("%v:%v", roomId, roomName)
	exists, err := repo.redis.Exists(ctx, redisKey).Result()
	if err != nil {
		log.Printf("[RemoveUserFromChatRoom] error checking key existence: %v, err: %v\n", redisKey, err)
		return http.StatusInternalServerError, fmt.Errorf("error checking key existence")
	}
	if exists == 0 {
		log.Printf("[RemoveUserFromChatRoom] key does not exist: %v\n", redisKey)
		return http.StatusBadRequest, fmt.Errorf("chat room does not exist")
	}
	status := repo.redis.LRem(ctx, redisKey, 0, username)
	if err := status.Err(); err != nil {
		log.Printf("[RemoveUserFromChatRoom] error on key: %v, username: %v, err: %v\n", redisKey, username, err)
		return http.StatusInternalServerError, fmt.Errorf("error removing user from chat room")
	}
	return http.StatusAccepted, nil
}

func (repo *Repository) ListUsersInChatRoom(ctx context.Context, roomId, roomName string) ([]string, int, error) {
	redisKey := fmt.Sprintf("%v:%v", roomId, roomName)
	exists, err := repo.redis.Exists(ctx, redisKey).Result()
	if err != nil {
		log.Printf("[ListUsersInChatRoom] error checking key existence: %v, err: %v\n", redisKey, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error checking key existence")
	}
	if exists == 0 {
		log.Printf("[ListUsersInChatRoom] key does not exist: %v\n", redisKey)
		return nil, http.StatusBadRequest, fmt.Errorf("chat room does not exist")
	}
	users, err := repo.redis.LRange(ctx, redisKey, 0, -1).Result()
	if err != nil {
		log.Printf("[ListUsersInChatRoom] error on key: %v, err: %v\n", redisKey, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error retrieving users from chat room")
	}
	return users, http.StatusOK, nil
}

func (repo *Repository) PublishMessageToChatRoom(ctx context.Context, channel, msg string) error {
	return repo.redis.Publish(ctx, channel, msg).Err()
}

func (repo *Repository) SubscribeToChatRoom(ctx context.Context, conn *websocket.Conn, channel string) {
	pubSub := repo.redis.Subscribe(ctx, channel)
	defer pubSub.Close()

	ch := pubSub.Channel()
	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Printf("[SubscribeToChatRoom] Failed to send msg to websocket for channel:%v : %v", channel, err)
			return
		}
	}
}

func convertToInterfaceSlice(s []string) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
