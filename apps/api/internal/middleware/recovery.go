package middleware
// How to recover and save the application in case of any issues.and
// Acts as a buffer
import (
	"log/slog"
	"net/http"
)

func Recovery(appLogger *slog.Logger, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				// First try to get the request ID from the response header.
				requestID := w.Header().Get("X-Request-ID")

				// Fallback to the request context.
				if requestID == "" {
					requestID = GetRequestID(r.Context())
				}

				// Final fallback.
				if requestID == "unknown" || requestID == "" {
					requestID = "unavailable"
				}

				appLogger.Error(
                	"panic recovered",
                	"request_id", requestID,
                	"method", r.Method,
                	"path", r.URL.Path,
                	"error", err,
                )

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

		}()

		next.ServeHTTP(w, r)
	})
}