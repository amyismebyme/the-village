package middleware

import (
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"log/slog"
	"net/http"
	"time"
)

func Logging(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		rec := httputil.NewResponseRecorder(w)

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		logger.Info(
			"http request completed",
			"request_id", GetRequestID(r.Context()),
			"method", r.Method,
			"route", routeLabel(r),
			"status", rec.Status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}
