package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerSystemRoutes(
	mux *http.ServeMux,
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
) {
	healthHandler := handlers.NewHealthHandler(
		appLogger,
		healthRegistry,
	)

	mux.Handle(
		"GET /health",
		healthHandler,
	)

	mux.HandleFunc(
		"GET /ready",
		handlers.ReadyHandler,
	)

	mux.HandleFunc(
		"GET /version",
		handlers.VersionHandler,
	)

	mux.HandleFunc(
		"GET /status",
		handlers.StatusHandler,
	)

	mux.Handle(
		"GET /metrics",
		promhttp.Handler(),
	)
}
