package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

// registerRoutes registers all application routes.
//
// Keep domain-specific route registration in separate files.
// This function acts as the single place where the router
// assembles the application's HTTP surface.
func registerRoutes(
	mux *http.ServeMux,
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
	handler *handlers.Handler,
) {
	registerSystemRoutes(
		mux,
		appLogger,
		healthRegistry,
	)

	registerAPIV1Routes(
		mux,
		handler,
	)
}
