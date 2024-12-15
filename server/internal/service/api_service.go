package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/websocket"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"github.com/the-arcade-01/go-chat-app/server/internal/repository"
	"github.com/the-arcade-01/go-chat-app/server/internal/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ApiService struct {
	repo *repository.Repository
}

func NewApiService() *ApiService {
	return &ApiService{
		repo: repository.NewRepository(),
	}
}

func (s *ApiService) AuthGreet(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, fmt.Sprintf("Hello, %s!", username))
}

func (s *ApiService) Signup(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide valid body"}, nil))
		return
	}
	defer r.Body.Close()

	token, status, err := s.repo.RegisterUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, http.StatusCreated, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "user created successfully"}, models.UserLoginResponse{Token: token}))
}

func (s *ApiService) Login(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide valid body"}, nil))
		return
	}
	defer r.Body.Close()

	token, status, err := s.repo.LoginUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "user logged in successfully"}, models.UserLoginResponse{Token: token}))
}

func (s *ApiService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	users, err := s.repo.GetAllUsers(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusInternalServerError, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "error on the server"}, nil))
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "list of all users"}, users))
}

func (s *ApiService) CreateRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	var body *models.CreateRoomReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct body"}, nil))
		return
	}

	room, status, err := s.repo.CreateRoom(r.Context(), body, username)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "Room created successfully"}, room))
}

func (s *ApiService) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	id := chi.URLParam(r, "roomId")
	roomId, err := strconv.Atoi(id)
	if err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct roomId"}, nil))
		return
	}

	status, err := s.repo.DeleteRoom(r.Context(), roomId, username)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "room deleted successfully"}, nil))
}

func (s *ApiService) GetUsersRooms(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	rooms, status, err := s.repo.GetUsersRooms(r.Context(), username)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "list of all rooms"}, rooms))
}

func (s *ApiService) AddUserToRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	var body *models.AddUsersToRoomReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct body"}, nil))
		return
	}

	status, err := s.repo.AddUsersToRoom(r.Context(), body, username)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "users added successfully"}, nil))
}

func (s *ApiService) RemoveUserFromRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	roomIdStr := r.URL.Query().Get("roomId")
	roomId, err := strconv.Atoi(roomIdStr)
	if err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct roomId"}, nil))
		return
	}

	removeUser := r.URL.Query().Get("username")
	if removeUser == "" {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct username to remove"}, nil))
		return
	}

	status, err := s.repo.RemoveUserFromRoom(r.Context(), roomId, removeUser, username)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "user removed successfully"}, nil))
}

func (s *ApiService) ListUsersInRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	roomIdStr := r.URL.Query().Get("roomId")
	roomId, err := strconv.Atoi(roomIdStr)
	if err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "please provide correct roomId"}, nil))
		return
	}

	users, status, err := s.repo.ListUsersInRoom(r.Context(), roomId)
	if err != nil {
		models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: err.Error()}, nil))
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "fetched users successfully"}, users))
}

// JoinChatRoom
/* Subscribes the redis message channel, and listens for the incoming messages
 */
func (s *ApiService) JoinChatRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid token"}, nil))
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewResponse(http.StatusUnauthorized, models.MetaResponse{Msg: "invalid or missing username in token claims"}, nil))
		return
	}

	roomId := r.URL.Query().Get("roomId")
	roomName := r.URL.Query().Get("roomName")

	if roomId == "" || roomName == "" {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewResponse(http.StatusBadRequest, models.MetaResponse{Msg: "missing required parameters"}, nil))
		return
	}

	channel := utils.GetRedisKey("channel", "_", roomId, roomName)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[JoinChatRoom] Failed to upgrade to websocket: %v\n", err)
		return
	}
	defer conn.Close()

	go s.repo.SubscribeToChatRoom(r.Context(), conn, channel)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[JoinChatRoom] error reading msg: %v", err)
			break
		}

		err = s.repo.PublishMessageToChatRoom(r.Context(), channel, string(msg))
		if err != nil {
			log.Printf("[JoinChatRoom] Failed to publish message to Redis: %v", err)
		}
	}
}
