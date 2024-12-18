package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/the-arcade-01/go-chat-app/server/internal/config"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"github.com/the-arcade-01/go-chat-app/server/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db    *sql.DB
	cache *redis.Client
}

func NewRepository() *Repository {
	appConfig := config.NewAppConfig()
	return &Repository{
		db:    appConfig.Db,
		cache: appConfig.Cache,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) (*models.UserLoginResponse, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[CreateUser] error on starting transaction: %v\n", err)
		return nil, http.StatusInternalServerError, errors.New("error on creating user, please try again later")
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
		}
	}(tx)

	var existingUserID int
	err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", user.Username).Scan(&existingUserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[CreateUser] error on checking existing user for %v, err: %v\n", user.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on checking existing user")
	}

	if existingUserID != 0 {
		return nil, http.StatusBadRequest, errors.New("please use different username")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to hash password")
	}

	result, err := tx.ExecContext(ctx, "INSERT INTO users (username, password) VALUES (?, ?)", user.Username, string(hashedPassword))
	if err != nil {
		log.Printf("[CreateUser] error on creating user %v, err: %v\n", user.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on creating user, please try again later")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		log.Printf("[CreateUser] error on getting userId for user %v, err: %v\n", user.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on creating user, please try again later")
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		log.Printf("[CreateUser] error on generating token for user: %v, err: %v\n", user.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on creating user, please try again later")
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[CreateUser] error on committing transaction: %v\n", err)
		return nil, http.StatusInternalServerError, errors.New("error on creating user, please try again later")
	}

	userLogin := &models.UserLoginResponse{
		UserDetails: models.UserDetails{
			Username: user.Username,
			UserId:   int(userId),
		},
		Token: token,
	}

	return userLogin, http.StatusCreated, nil
}

func (r *Repository) LoginUser(ctx context.Context, user *models.User) (*models.UserLoginResponse, int, error) {
	var existingUser models.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, password FROM users WHERE username = ?", user.Username).Scan(&existingUser.UserId, &existingUser.Username, &existingUser.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.StatusUnauthorized, errors.New("invalid username or password")
		}
		log.Printf("[LoginUser] error on retrieving user %v, err: %v\n", user.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on logging in, please try again later")
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid username or password")
	}

	token, err := utils.GenerateJWT(existingUser.Username)
	if err != nil {
		log.Printf("[LoginUser] error on generating token for user: %v, err: %v\n", existingUser.Username, err)
		return nil, http.StatusInternalServerError, errors.New("error on logging in, please try again later")
	}

	userLogin := &models.UserLoginResponse{
		UserDetails: models.UserDetails{
			UserId:   existingUser.UserId,
			Username: existingUser.Username,
		},
		Token: token,
	}

	return userLogin, http.StatusOK, nil
}

// CreateRoom TODO: this is breaking, check on transaction
func (r *Repository) CreateRoom(ctx context.Context, room *models.Room) (*models.Room, int, error) {
	roomId := "room:" + room.RoomName
	exists, err := r.cache.Exists(ctx, roomId).Result()
	if err != nil {
		log.Printf("[CreateRoom] error creating room %v, err: %v\n", room.RoomName, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating room, please try again later")
	}
	if exists == 1 {
		return nil, http.StatusConflict, fmt.Errorf("please use a different room name")
	}

	pipe := r.cache.TxPipeline()
	pipe.HSet(ctx, "room:metadata:"+room.RoomName, "active_users", 0, "created_at", time.Now().Unix())
	pipe.Expire(ctx, "room:metadata:"+room.RoomName, 24*time.Hour)
	pipe.SAdd(ctx, "room:users:"+room.RoomName)
	pipe.Expire(ctx, "room:users:"+room.RoomName, 24*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("[CreateRoom] error setting room %v, err: %v\n", room.RoomName, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating room, please try again later")
	}

	room.RoomId = roomId
	return room, http.StatusCreated, nil
}

func (r *Repository) GetRooms(ctx context.Context) ([]*models.Room, int, error) {
	keys, err := r.cache.Keys(ctx, "room:metadata:*").Result()
	if err != nil {
		log.Printf("[GetRooms] error fetching rooms, err: %v\n", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error fetching rooms")
	}

	rooms := make([]*models.Room, 0, len(keys))
	for _, key := range keys {
		metadata, err := r.cache.HGetAll(ctx, key).Result()
		if err != nil {
			log.Printf("[GetRooms] error fetching metadata for %v, err: %v\n", key, err)
			continue
		}

		roomName := strings.TrimPrefix(key, "room:metadata:")
		activeUsers, _ := strconv.Atoi(metadata["active_users"])
		rooms = append(rooms, &models.Room{
			RoomId:      "room:" + roomName,
			RoomName:    roomName,
			ActiveUsers: activeUsers,
		})
	}

	return rooms, http.StatusOK, nil
}

func (r *Repository) GetRoomDetails(ctx context.Context, roomId string) (*models.Room, int, error) {
	metadataKey := "room:metadata:" + strings.TrimPrefix(roomId, "room:")
	usersKey := "room:users:" + strings.TrimPrefix(roomId, "room:")

	pipe := r.cache.TxPipeline()
	metadataCmd := pipe.HGetAll(ctx, metadataKey)
	usersCmd := pipe.SMembers(ctx, usersKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[GetRoomDetails] error fetching room details for room %v, err: %v\n", roomId, err)
		return nil, http.StatusInternalServerError, fmt.Errorf("error fetching room details, please try again later")
	}

	metadata := metadataCmd.Val()
	activeUsers, _ := strconv.Atoi(metadata["active_users"])
	usernames := usersCmd.Val()

	return &models.Room{
		RoomId:      roomId,
		RoomName:    strings.TrimPrefix(roomId, "room:"),
		ActiveUsers: activeUsers,
		Users:       usernames,
	}, http.StatusOK, nil
}

func (r *Repository) IncDecActiveUsers(ctx context.Context, roomId string, val int64, username string) {
	usersKey := "room:users:" + strings.TrimPrefix(roomId, "room:")

	pipe := r.cache.TxPipeline()
	pipe.HIncrBy(ctx, roomId, "active_users", val)

	if val > 0 {
		pipe.SAdd(ctx, usersKey, username)
	} else if val < 0 {
		pipe.SRem(ctx, usersKey, username)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[IncDecActiveUsers] error updating active_users or usernames for room %v, err: %v\n", roomId, err)
	}
}

func (r *Repository) PublishMessageToChatRoom(ctx context.Context, roomId, user, content string, msgType models.MessageType) error {
	msg := models.Message{
		User:    user,
		Type:    msgType,
		Content: content,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal chat message: %w", err)
	}

	return r.cache.Publish(ctx, roomId, string(payload)).Err()
}

func (r *Repository) SubscribeToChatRoom(ctx context.Context, conn *websocket.Conn, roomId string, done chan struct{}) {
	pubSub := r.cache.Subscribe(ctx, roomId)
	defer pubSub.Close()

	ch := pubSub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case msg := <-ch:
			var chatMsg models.Message
			if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
				log.Printf("[SubscribeToChatRoom] Failed to unmarshal chat message: %v", err)
				continue
			}

			wsMessage, err := json.Marshal(chatMsg)
			if err != nil {
				log.Printf("[SubscribeToChatRoom] Failed to marshal chat message: %v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, wsMessage); err != nil {
				log.Printf("[SubscribeToChatRoom] Failed to send msg to websocket for channel:%v : %v", roomId, err)
				return
			}
		}
	}
}
