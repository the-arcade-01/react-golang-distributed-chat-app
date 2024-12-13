package models

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Status int          `json:"status"`
	Meta   MetaResponse `json:"meta"`
	Data   interface{}  `json:"data"`
}

type MetaResponse struct {
	Msg string `json:"message"`
}

func NewResponse(status int, meta MetaResponse, data interface{}) *Response {
	return &Response{
		Status: status,
		Meta:   meta,
		Data:   data,
	}
}

func ResponseWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("[ResponseWithJSON] error parsing payload: %v\n", payload)
	}
}
