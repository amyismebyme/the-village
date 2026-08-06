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
		"POST /api/v1/communities",
		h.CreateCommunity,
	)

	mux.HandleFunc(
		"GET /api/v1/communities",
		h.ListCommunities,
	)

	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		h.GetCommunity,
	)

	mux.HandleFunc(
		"PUT /api/v1/communities/{id}",
		h.UpdateCommunity,
	)

	mux.HandleFunc(
		"DELETE /api/v1/communities/{id}",
		h.DeleteCommunity,
	)
}
