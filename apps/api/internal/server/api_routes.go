package server

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
)

// registerAPIV1Routes registers all version 1 API domain routes.
//
// Keep individual domains isolated in their own route files.
// This function provides the grouping boundary for /api/v1.
func registerAPIV1Routes(
	mux *http.ServeMux,
	handler *handlers.Handler,
) {
	registerCommunityRoutes(
		mux,
		handler,
	)
}
