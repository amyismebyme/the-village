package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

func NewHTTPServer(
	appLogger *slog.Logger,
	cfg config.Config,
	healthRegistry *health.Registry,
	handler *handlers.Handler,
) *http.Server {

	return &http.Server{
		Addr: ":" + cfg.Port,

		Handler: NewRouter(
			appLogger,
			healthRegistry,
			handler,
		),

		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
