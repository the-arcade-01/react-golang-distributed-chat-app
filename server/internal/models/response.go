package models

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func ResponseWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("[ResponseWithJSON] error on sending response %v, err: %v\n", payload, err)
		return
	}
}
