package handlers

import (
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

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}
