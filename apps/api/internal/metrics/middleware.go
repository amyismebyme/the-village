package metrics

import (
	"net/http"
	"strconv"
	"strings"
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

		// ServeMux populates Request.Pattern after route matching.
		//
		// With method-aware ServeMux patterns, Pattern can be:
		// "GET /api/v1/communities/{id}".
		// Metrics use the path pattern only.
		route := r.Pattern

		if route != "" {
			if i := strings.IndexByte(route, ' '); i >= 0 {
				route = route[i+1:]
			}
		}

		// If there is no ServeMux route pattern, use the request path.
		if route == "" {
			route = r.URL.Path
		}

		// Keep the normalized pattern available to callers/tests.
		r.Pattern = route

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
	})
}
