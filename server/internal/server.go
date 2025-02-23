package internal

import (
	"net/http"
)

/*
for security think about, rate limit, ws should have origin check which doesn't allow it to connect with other app
*/
func Run() {
	handlers := newHandlers()

	http.HandleFunc("/chat/ping", func(w http.ResponseWriter, r *http.Request) {
		ResponseWithJSON(w, http.StatusOK, nil)
	})

	http.HandleFunc("/chat/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.handleWS(w, r)
	})

	if Envs.ENV == "development" {
		http.HandleFunc("/chat/greet", handlers.greet)
		http.HandleFunc("/chat/exists", handlers.checkExists)
	}

	Log.Info("server running on port:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		Log.Error("error on starting server", "error", err)
	}
}
