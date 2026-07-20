package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Middleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rec := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		start := time.Now()

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		RequestsTotal.
			WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rec.status),
			).
			Inc()

		RequestDuration.
			WithLabelValues(
				r.Method,
				r.URL.Path,
			).
			Observe(duration.Seconds())

	})

}