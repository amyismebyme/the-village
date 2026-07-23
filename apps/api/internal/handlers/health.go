package handlers

//Health handler for the app
import (
	"encoding/json"
	"net/http"
	"log"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	response := HealthResponse{
		Status: "healthy",
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("failed to encode Root response: %v", err)
		http.Error(
			w,
			"failed to encode health response",
			http.StatusInternalServerError,
		)
		return
	}
}
