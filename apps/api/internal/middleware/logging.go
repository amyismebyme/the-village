package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
)

func Logging(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := httputil.NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		logger.InfoContext(
			r.Context(),
			"http request completed",
			"request_id", GetRequestID(r.Context()),
			"method", r.Method,
			"route", httputil.RouteLabel(r),
			"status", rec.Status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
