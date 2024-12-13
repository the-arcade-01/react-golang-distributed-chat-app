package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/the-arcade-01/go-chat-app/server/internal/service"
)

type Server struct {
	Router *chi.Mux
}

func (s *Server) mountMiddlewares() {
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Heartbeat("/ping"))
}

func (s *Server) mountHandlers() {
	apiService := service.NewApiService()
	s.Router.Get("/greet", apiService.Greet)
	s.Router.Post("/auth/login", apiService.Login)
	s.Router.Post("/auth/signup", apiService.Signup)

	s.Router.Get("/db/count", apiService.GetUsersTotalCount)
	s.Router.Get("/redis", apiService.GetRedisValue)
	s.Router.Post("/redis", apiService.SetRedisValue)
}

func CreateNewServer() *Server {
	server := &Server{
		Router: chi.NewRouter(),
	}
	server.mountMiddlewares()
	server.mountHandlers()
	return server
}
