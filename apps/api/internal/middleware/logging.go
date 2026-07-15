package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {

	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: w,
			status: http.StatusOK,
		}

		next.ServeHTTP(recorder, r)
		duration := time.Since(start)

		log.Printf(
			"%s %s %d %v request_id=%s",
			r.Method,
			r.URL.Path,
			recorder.status,
			duration,
			GetRequestID(r.Context()),
		)
	})
}