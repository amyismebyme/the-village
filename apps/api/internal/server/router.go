package server

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func NewRouter(
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
	handler *handlers.Handler,
) http.Handler {

	mux := http.NewServeMux()

	registerRoutes(
		mux,
		appLogger,
		healthRegistry,
		handler,
	)

	return middleware.Recovery(
		appLogger,
		middleware.RequestID(
			middleware.Logging(
				appLogger,
				metrics.Middleware(
					mux,
				),
			),
		),
	)
}
