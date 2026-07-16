package middleware
//Logging of requests and all the details
import (
	"log/slog"
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

func Logging(appLogger *slog.Logger, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: w,
			status: http.StatusOK,
		}

		next.ServeHTTP(recorder, r)
		duration := time.Since(start)

		appLogger.Info(
        	"request completed",
        	"request_id", GetRequestID(r.Context()),
        	"method", r.Method,
        	"path", r.URL.Path,
        	"status", recorder.status,
        	"duration_ms", duration.Milliseconds(),
        )
	})
}