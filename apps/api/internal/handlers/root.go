package handlers

import (
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
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

	httputil.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}
