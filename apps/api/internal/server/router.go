package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func NewRouter(
	appLogger *slog.Logger,
	healthRegistry *health.Registry,
	handler *handlers.Handler,
	requestTimeout ...time.Duration,
) http.Handler {

	mux := http.NewServeMux()

	registerRoutes(
		mux,
		appLogger,
		healthRegistry,
		handler,
	)

	timeout := time.Duration(0)
	if len(requestTimeout) > 0 {
		timeout = requestTimeout[0]
	}

	return middleware.RequestID(
		middleware.Logging(
			appLogger,
			metrics.Middleware(
				middleware.Recovery(
					appLogger,
					middleware.RequestTimeout(
						timeout,
						mux,
					),
				),
			),
		),
	)
}
