package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.Handle(
		"/health",
		handlers.NewHealthHandler(appLogger, healthRegistry),
	)

	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)
	mux.HandleFunc("/status", handlers.StatusHandler)
	mux.Handle("/metrics", promhttp.Handler())

	return middleware.Recovery(
		appLogger,
		middleware.RequestID(
			middleware.Logging(appLogger, mux),
		),
	)
}
