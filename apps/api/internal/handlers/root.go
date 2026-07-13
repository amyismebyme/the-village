package handlers

import (
	"encoding/json"
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
		http.Error(
			w,
			"failed to encode Root response",
			http.StatusInternalServerError,
		)
		return
	}
}
