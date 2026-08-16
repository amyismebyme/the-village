package handlers

import (
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"net/http"
)

type ReadyResponse struct {
	Status string `json:"status"`
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	response := ReadyResponse{
		Status: "ready",
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}
