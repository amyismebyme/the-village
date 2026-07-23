package handlers

//Root or Deafult handler for the app

import (
	"encoding/json"
	"log"
	"net/http"
)

type RootResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	response := RootResponse{
		Service: "village-api",
		Status:  "running",
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("failed to encode Root response: %v", err)
		http.Error(
			w,
			"failed to encode Root response",
			http.StatusInternalServerError,
		)
		return
	}
}
