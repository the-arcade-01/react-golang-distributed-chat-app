package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/websocket"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"github.com/the-arcade-01/go-chat-app/server/internal/repository"
)

type Service struct {
	upgrader websocket.Upgrader
	repo     *repository.Repository
}

func NewService() *Service {
	return &Service{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		repo: repository.NewRepository(),
	}
}

func (s *Service) Greet(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, "Hello, World")
}

func (s *Service) SignUp(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.Response{
			Msg:  "please provide correct request body",
			Data: nil,
		})
		return
	}
	defer r.Body.Close()
	user, status, err := s.repo.CreateUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, models.Response{
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	models.ResponseWithJSON(w, status, models.Response{
		Msg:  "user created successfully",
		Data: user,
	})
}

func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.Response{
			Msg:  "please provide correct request body",
			Data: nil,
		})
		return
	}
	defer r.Body.Close()
	user, status, err := s.repo.LoginUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, models.Response{
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	models.ResponseWithJSON(w, status, models.Response{
		Msg:  "user logged in successfully",
		Data: user,
	})
}

func (s *Service) CreateRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}
	var body *models.Room
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.Response{
			Msg:  "please provide correct request body",
			Data: nil,
		})
	}
	room, status, err := s.repo.CreateRoom(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, models.Response{
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	models.ResponseWithJSON(w, status, models.Response{
		Msg:  "room created successfully",
		Data: room,
	})
}

func (s *Service) GetRooms(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}

	rooms, status, err := s.repo.GetRooms(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, status, models.Response{
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	models.ResponseWithJSON(w, status, models.Response{
		Msg:  "list of all rooms",
		Data: rooms,
	})
}

func (s *Service) HandleWs(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.Response{
			Msg:  "invalid auth credentials",
			Data: nil,
		})
		return
	}

	roomId := r.URL.Query().Get("room_id")
	if roomId == "" {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.Response{
			Msg:  "please provide room_id",
			Data: nil,
		})
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[HandleWs] Failed to upgrade to WebSocket: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	metadataKey := "room:metadata:" + strings.TrimPrefix(roomId, "room:")
	s.repo.IncDecActiveUsers(ctx, metadataKey, 1)

	defer func() {
		s.repo.IncDecActiveUsers(ctx, metadataKey, -1)
	}()

	joinMsg := fmt.Sprintf("%s has joined the room.", username)
	err = s.repo.PublishMessageToChatRoom(ctx, roomId, joinMsg)
	if err != nil {
		log.Printf("[HandleWs] Failed to publish join message: %v", err)
		return
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Printf("[HandleWs] Error reading message: %v", err)
					close(done)
					return
				}
				err = s.repo.PublishMessageToChatRoom(ctx, roomId, string(msg))
				if err != nil {
					log.Printf("[HandleWs] Failed to publish message: %v", err)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.repo.SubscribeToChatRoom(ctx, conn, roomId, done)
	}()

	<-done

	conn.Close()
	wg.Wait()

	leaveMsg := fmt.Sprintf("%s has left the room.", username)
	err = s.repo.PublishMessageToChatRoom(ctx, roomId, leaveMsg)
	if err != nil {
		log.Printf("[HandleWs] Failed to publish leave message: %v", err)
	}
}
