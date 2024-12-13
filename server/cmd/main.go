package main

import (
	"log"
	"net/http"

	"github.com/the-arcade-01/go-chat-app/server/internal/api"
)

func main() {
	server := api.CreateNewServer()
	log.Println("[main] server running on port:8080")
	err := http.ListenAndServe(":8080", server.Router)
	if err != nil {
		log.Println("[main] error launching server", err)
		return
	}
}
