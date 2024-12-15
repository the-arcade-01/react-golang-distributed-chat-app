package repository

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

func (repo *Repository) RegisterUser(ctx context.Context, user *models.User) (string, int, error) {
	var existingUser models.User
	if err := repo.db.WithContext(ctx).Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return "", http.StatusConflict, fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to hash password")
	}
	user.Password = string(hashedPassword)

	if err := repo.db.WithContext(ctx).Create(&user).Error; err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to create user")
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error generating token")
	}

	return token, http.StatusCreated, nil
}

func (repo *Repository) LoginUser(ctx context.Context, user *models.User) (string, int, error) {
	var existingUser models.User
	if err := repo.db.WithContext(ctx).Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
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

func (repo *Repository) GetAllUsers(ctx context.Context) ([]string, error) {
	var usernames []string
	if err := repo.db.WithContext(ctx).Model(&models.User{}).Pluck("username", &usernames).Error; err != nil {
		log.Printf("[GetAllUsers] error %v\n", err)
		return nil, err
	}
	return usernames, nil
}

func (repo *Repository) CreateRoom(ctx context.Context, body *models.CreateRoomReqBody, username string) (*models.Room, int, error) {
	room := &models.Room{
		RoomName: body.RoomName,
		Admin:    username,
	}

	if err := repo.db.WithContext(ctx).Create(&room).Error; err != nil {
		log.Printf("[CreateRoom] error creating room: %v\n", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating room")
	}

	return room, http.StatusCreated, nil
}

func (repo *Repository) DeleteRoom(ctx context.Context, roomId int, username string) (int, error) {
	var room models.Room
	if err := repo.db.WithContext(ctx).First(&room, roomId).Error; err != nil {
		return http.StatusNotFound, fmt.Errorf("room not found")
	}

	if room.Admin != username {
		return http.StatusForbidden, fmt.Errorf("only the admin can delete the room")
	}

	if err := repo.db.WithContext(ctx).Delete(&room).Error; err != nil {
		log.Printf("[DeleteRoom] error deleting room: %v\n", err)
		return http.StatusInternalServerError, fmt.Errorf("error deleting room")
	}

	return http.StatusOK, nil
}

func (repo *Repository) GetUsersRooms(ctx context.Context, username string) ([]*models.Room, int, error) {
	var rooms []*models.Room
	err := repo.db.WithContext(ctx).Where("admin = ?", username).Find(&rooms).Error
	if err != nil {
		log.Printf("[GetUsersRooms] error finding rooms for user %s: %v\n", username, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error finding rooms for user")
	}
	return rooms, http.StatusOK, nil
}

func (repo *Repository) AddUsersToRoom(ctx context.Context, body *models.AddUsersToRoomReqBody, username string) (int, error) {
	var room models.Room
	if err := repo.db.WithContext(ctx).First(&room, body.RoomId).Error; err != nil {
		return http.StatusNotFound, fmt.Errorf("room not found")
	}

	if room.Admin != username {
		return http.StatusUnauthorized, fmt.Errorf("only the admin can add users to room")
	}

	var userRooms []models.UserRooms
	for _, user := range body.Users {
		userRooms = append(userRooms, models.UserRooms{
			Username: user,
			RoomId:   body.RoomId,
		})
	}

	if err := repo.db.WithContext(ctx).Create(&userRooms).Error; err != nil {
		log.Printf("[AddUsersToRoom] error adding users to room %d: %v\n", body.RoomId, err)
		return http.StatusInternalServerError, fmt.Errorf("error adding users to room")
	}

	return http.StatusOK, nil
}

func (repo *Repository) RemoveUserFromRoom(ctx context.Context, roomId int, removeUser, username string) (int, error) {
	var room models.Room
	if err := repo.db.WithContext(ctx).First(&room, roomId).Error; err != nil {
		return http.StatusNotFound, fmt.Errorf("room not found")
	}

	if room.Admin != username {
		return http.StatusUnauthorized, fmt.Errorf("only the admin can remove users from the room")
	}

	if err := repo.db.WithContext(ctx).Where("room_id = ? AND username = ?", roomId, removeUser).Delete(&models.UserRooms{}).Error; err != nil {
		log.Printf("[RemoveUserForRoom] error removing user %s from room %d: %v\n", removeUser, roomId, err)
		return http.StatusInternalServerError, fmt.Errorf("error removing user from room")
	}

	return http.StatusOK, nil
}

func (repo *Repository) ListUsersInRoom(ctx context.Context, roomId int) ([]string, int, error) {
	var usernames []string
	if err := repo.db.WithContext(ctx).Where("room_id = ?", roomId).Pluck("username", &usernames).Error; err != nil {
		log.Printf("[ListUsersInRoom] error retrieving users for room %d: %v\n", roomId, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error retrieving users for room")
	}
	return usernames, http.StatusOK, nil
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
