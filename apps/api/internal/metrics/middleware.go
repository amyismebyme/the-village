package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	// net/http ignores subsequent WriteHeader calls after the first one.
	if r.wroteHeader {
		return
	}

	r.status = code
	r.wroteHeader = true

	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	// net/http implicitly sends a 200 response when Write is called
	// before WriteHeader.
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(body)
}

// Unwrap allows http.ResponseController and other stdlib functionality
// to reach the original ResponseWriter capabilities.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		RequestsInFlight.Inc()
		defer RequestsInFlight.Dec()

		rec := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

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
				strconv.Itoa(rec.status),
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
