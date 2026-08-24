package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		RequestsInFlight.Inc()
		defer RequestsInFlight.Dec()

		rec := httputil.NewResponseRecorder(w)

		start := time.Now()

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		route := httputil.RouteLabel(r)

		RequestsTotal.
			WithLabelValues(
				r.Method,
				route,
				strconv.Itoa(rec.Status),
			).
			Inc()

		RequestDuration.
			WithLabelValues(
				r.Method,
				route,
			).
			Observe(duration.Seconds())

		if rec.Status >= http.StatusBadRequest {
			errorType := "client"
			if rec.Status >= http.StatusInternalServerError {
				errorType = "server"
			}

			ErrorsTotal.WithLabelValues(errorType).Inc()
		}
	})
}
