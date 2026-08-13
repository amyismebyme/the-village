package handlers

import (
	"net/http"
)

type ReadyResponse struct {
	Status string `json:"status"`
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	response := ReadyResponse{
		Status: "ready",
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}
