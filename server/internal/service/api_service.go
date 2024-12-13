package service

import (
	"net/http"

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

func (s *ApiService) Login(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, "Logged in")
}

func (s *ApiService) Signup(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, "Signed up")
}
