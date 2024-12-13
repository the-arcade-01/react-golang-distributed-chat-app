package service

import (
	"net/http"

	"github.com/the-arcade-01/go-chat-app/server/internal/models"
)

type ApiService struct {
}

func NewApiService() *ApiService {
	return &ApiService{}
}

func (s *ApiService) Greet(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, "Hello, World!!")
}
