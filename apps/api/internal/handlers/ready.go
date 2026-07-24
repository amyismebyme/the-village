package handlers

//Ready handler for the app

import (
	"encoding/json"
	"log"
	"net/http"
)

type ReadyResponse struct {
	Status string `json:"status"`
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {

	response := ReadyResponse{
		Status: "ready",
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("failed to encode Root response: %v", err)
		http.Error(
			w,
			"failed to encode Ready response",
			http.StatusInternalServerError,
		)
		return
	}
}
