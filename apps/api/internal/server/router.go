package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(appLogger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)
	mux.HandleFunc("/status", handlers.StatusHandler)
	mux.Handle("/metrics", promhttp.Handler())

	handler := middleware.Recovery(
		appLogger,
		middleware.RequestID(
			middleware.Logging(appLogger, mux),
		),
	)

	return handler
}
