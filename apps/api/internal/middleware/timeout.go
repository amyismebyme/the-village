package middleware

import (
	"context"
	"net/http"
	"time"
)

// RequestTimeout bounds the lifetime of an HTTP request context.
// A non-positive timeout leaves the incoming request context unchanged.
func RequestTimeout(
	timeout time.Duration,
	next http.Handler,
) http.Handler {
	if timeout <= 0 {
		return next
	}

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ctx, cancel := context.WithTimeout(
			r.Context(),
			timeout,
		)
		defer cancel()

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
