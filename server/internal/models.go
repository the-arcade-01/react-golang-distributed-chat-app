package internal

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Timestamp int64  `json:"timestamp"`
	Username  string `json:"username"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func ResponseWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
