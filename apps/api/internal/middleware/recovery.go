package middleware

import (
	"log/slog"
	"net/http"
)

func Recovery(appLogger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := w.Header().Get("X-Request-ID")

				if requestID == "" {
					requestID = GetRequestID(r.Context())
				}

				if requestID == "unknown" || requestID == "" {
					requestID = "unavailable"
				}

				if appLogger != nil {
					appLogger.Error(
						"panic recovered",
						"request_id", requestID,
						"method", r.Method,
						"path", r.URL.Path,
						"error", err,
					)
				}

				// Keep the API error contract consistent.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				_, _ = w.Write([]byte(`{
  "error": {
    "code": "internal_error",
    "message": "internal server error"
  }
}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
