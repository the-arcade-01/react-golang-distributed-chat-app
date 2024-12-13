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

/* Below functions need to commented out */

func (s *ApiService) SetRedisValue(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("val")
	if len(key) == 0 || len(val) == 0 {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide correct values in key and val")
		return
	}
	err := s.repo.SetValue(r.Context(), key, val)
	if err != nil {
		models.ResponseWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, "key and value set successfully")
}

func (s *ApiService) GetRedisValue(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if len(key) == 0 {
		models.ResponseWithJSON(w, http.StatusBadRequest, "please provide correct values in key")
		return
	}
	val, err := s.repo.GetValue(r.Context(), key)
	if err != nil {
		models.ResponseWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, val)
}

func (s *ApiService) GetUsersTotalCount(w http.ResponseWriter, r *http.Request) {
	count, err := s.repo.GetCount(r.Context())
	if err != nil {
		models.ResponseWithJSON(w, http.StatusInternalServerError, err)
		return
	}
	models.ResponseWithJSON(w, http.StatusOK, count)
}
