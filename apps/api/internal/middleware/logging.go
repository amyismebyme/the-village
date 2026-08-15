package middleware

import (
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

		rec := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		logger.Info(
			"http request completed",
			"request_id", GetRequestID(r.Context()),
			"method", r.Method,
			"route", routeLabel(r),
			"status", rec.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}
