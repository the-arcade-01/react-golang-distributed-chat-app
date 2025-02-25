package internal

import (
	"net/http"
)

func Run() {
	handlers := newHandlers()
	handlersStreams := newHandlersStreams()

	http.HandleFunc("/chat/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if Envs.WS_TYPE == "pubsub" {
		Log.Info("pubsub ws established")
		http.HandleFunc("/chat/ws", func(w http.ResponseWriter, r *http.Request) {
			handlers.handleWS(w, r)
		})
	} else {
		Log.Info("streams ws established")
		http.HandleFunc("/chat/ws", func(w http.ResponseWriter, r *http.Request) {
			handlersStreams.handleWS(w, r)
		})
	}

	Log.Info("server running on port:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		Log.Error("error on starting server", "error", err)
	}
}
