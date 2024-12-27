package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
	"github.com/the-arcade-01/go-chat-app/server/internal/service"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[init] error loading env, err: %v", err)
	}
}

type server struct {
	router *chi.Mux
}

func (s *server) mountMiddlewares() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.CleanPath)
	s.router.Use(middleware.Heartbeat("/ping"))
	s.router.Use(cors.Handler(cors.Options{
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

func (s *server) mountHandlers() {
	tokenAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET_KEY")), nil)

	svc := service.NewService()
	s.router.Get("/", svc.Greet)
	s.router.Post("/auth/signup", svc.SignUp)
	s.router.Post("/auth/login", svc.Login)
	s.router.Get("/ws", svc.HandleWs)

	s.router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Post("/rooms", svc.CreateRoom)
		r.Get("/rooms", svc.GetRooms)
		r.Get("/rooms/{room_id}", svc.GetRoomDetails)
		r.Delete("/rooms/{room_id}", svc.DeleteRoom)
	})
}

func newServer() *server {
	s := &server{
		router: chi.NewRouter(),
	}
	s.mountMiddlewares()
	s.mountHandlers()
	return s
}

func main() {
	server := newServer()
	log.Println("[main] server running on port:8080")
	err := http.ListenAndServe(":8080", server.router)
	if err != nil {
		log.Printf("[main] error on running server, err: %v\n", err)
		return
	}
}
