package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

// NewHTTPServer returns an HTTP server configured with application timeouts.
func NewHTTPServer(
	appLogger *slog.Logger,
	cfg config.Config,
	healthRegistry *health.Registry,
) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      NewRouter(appLogger, healthRegistry),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
