package server

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"log/slog"
)

func registerSystemRoutes(
	mux *http.ServeMux,
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
) {

	mux.HandleFunc("/", handlers.RootHandler)

	mux.Handle(
		"/health",
		handlers.NewHealthHandler(
			appLogger,
			healthRegistry,
		),
	)

	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)
	mux.HandleFunc("/status", handlers.StatusHandler)

	mux.Handle(
		"/metrics",
		promhttp.Handler(),
	)
}
