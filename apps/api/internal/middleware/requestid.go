package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Propagate the request ID back to the caller.
		w.Header().Set("X-Request-ID", requestID)

		// Store it in request context for logging/recovery/etc.
		ctx := WithRequestID(
			r.Context(),
			requestID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
