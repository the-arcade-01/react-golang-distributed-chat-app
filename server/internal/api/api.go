package api

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/the-arcade-01/go-chat-app/server/internal/service"
)

type Server struct {
	Router *chi.Mux
}

func (s *Server) mountMiddlewares() {
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.CleanPath)
	s.Router.Use(middleware.Heartbeat("/ping"))
	s.Router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
}

func (s *Server) mountHandlers() {
	tokenAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET_KEY")), nil)
	apiService := service.NewApiService()
	s.Router.Get("/greet", apiService.Greet)
	s.Router.Post("/auth/login", apiService.Login)
	s.Router.Post("/auth/signup", apiService.Signup)

	s.Router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Get("/auth/greet", apiService.AuthGreet)
	})

	// s.Router.Get("/db/count", apiService.GetUsersTotalCount)
	// s.Router.Get("/redis", apiService.GetRedisValue)
	// s.Router.Post("/redis", apiService.SetRedisValue)
}

func CreateNewServer() *Server {
	server := &Server{
		Router: chi.NewRouter(),
	}
	server.mountMiddlewares()
	server.mountHandlers()
	return server
}
