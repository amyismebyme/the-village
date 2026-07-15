package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// RequestID assigns a unique ID to every incoming request.
func RequestID(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := uuid.New().String()

		ctx := context.WithValue(
			r.Context(),
			RequestIDKey,
			requestID,
		)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

// GetRequestID returns the request ID from the request context.
func GetRequestID(ctx context.Context) string {

	requestID, ok := ctx.Value(RequestIDKey).(string)

	if !ok {
		return "unknown"
	}

	return requestID
}