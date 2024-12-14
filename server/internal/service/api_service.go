package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/the-arcade-01/go-chat-app/server/internal/models"
	"github.com/the-arcade-01/go-chat-app/server/internal/repository"
)

type ApiService struct {
	repo *repository.Repository
}

func NewApiService() *ApiService {
	return &ApiService{
		repo: repository.NewRepository(),
	}
}

func (s *ApiService) Greet(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, "Hello, World!!")
}

func (s *ApiService) AuthGreet(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	username, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, fmt.Sprintf("Hello, %s!", username))
}

func (s *ApiService) Signup(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide valid body")
		return
	}
	defer r.Body.Close()
	token, status, err := s.repo.RegisterUser(body)
	if err != nil {
		models.ResponseWithJSON(w, status, err)
		return
	}
	models.ResponseWithJSON(w, http.StatusCreated, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "user created successfully"}, models.UserLoginResponse{Token: token}))
}

func (s *ApiService) Login(w http.ResponseWriter, r *http.Request) {
	var body *models.User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide valid body")
		return
	}
	defer r.Body.Close()
	token, status, err := s.repo.LoginUser(body)
	if err != nil {
		models.ResponseWithJSON(w, status, err)
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "user logged in successfully"}, models.UserLoginResponse{Token: token}))
}

func (s *ApiService) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	_, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}
	users, err := s.repo.GetAllUsers()
	if err != nil {
		models.ResponseWithJSON(w, http.StatusInternalServerError, "error on the server")
		return
	}

	models.ResponseWithJSON(w, http.StatusOK, models.NewResponse(http.StatusCreated, models.MetaResponse{Msg: "list of all users"}, users))
}

func (s *ApiService) CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	username, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	var body *models.ChatRoomReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide valid body")
		return
	}
	defer r.Body.Close()

	roomInfo, status, err := s.repo.CreateChatRoom(r.Context(), body, username)
	if err != nil {
		models.ResponseWithJSON(w, status, "error on server")
		return
	}

	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "created room successfully"}, roomInfo))
}

func (s *ApiService) ListUsersInChatRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	_, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	roomId := r.URL.Query().Get("roomId")
	roomName := r.URL.Query().Get("roomName")

	if roomId == "" || roomName == "" {
		models.ResponseWithJSON(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	users, status, err := s.repo.ListUsersInChatRoom(r.Context(), roomId, roomName)
	if err != nil {
		models.ResponseWithJSON(w, status, err)
		return
	}

	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "data fetched successfully"}, users))
}

func (s *ApiService) AddUsersToChatRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	_, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	var body *models.ChatRoomAddUserReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide valid body")
		return
	}
	defer r.Body.Close()

	status, err := s.repo.AddUserToChatRoom(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, status, err)
		return
	}
	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "users added successfully"}, nil))
}

func (s *ApiService) RemoveUserFromChatRoom(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token")
		return
	}
	_, ok := claims["username"].(string)
	if !ok {
		models.ResponseWithJSON(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	roomId := r.URL.Query().Get("roomId")
	roomName := r.URL.Query().Get("roomName")
	username := r.URL.Query().Get("username")

	if roomId == "" || roomName == "" || username == "" {
		models.ResponseWithJSON(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	status, err := s.repo.RemoveUserFromChatRoom(r.Context(), roomId, roomName, username)
	if err != nil {
		models.ResponseWithJSON(w, status, err)
		return
	}

	models.ResponseWithJSON(w, status, models.NewResponse(status, models.MetaResponse{Msg: "user removed successfully"}, nil))
}
