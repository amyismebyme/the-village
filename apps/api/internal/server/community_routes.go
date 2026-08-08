package server

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
)

func registerCommunityRoutes(
	mux *http.ServeMux,
	h *handlers.Handler,
) {
	mux.HandleFunc(
		"/communities",
		h.CreateCommunity,
	)
}
