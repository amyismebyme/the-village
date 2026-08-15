package middleware

import "context"

type requestIDContextKey struct{}

func WithRequestID(
	ctx context.Context,
	requestID string,
) context.Context {
	return context.WithValue(
		ctx,
		requestIDContextKey{},
		requestID,
	)
}

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)

	if !ok || requestID == "" {
		return "unknown"
	}

	return requestID
}

// withRequestID is retained for package-level tests
// that use the unexported helper.
func withRequestID(
	ctx context.Context,
	requestID string,
) context.Context {
	return WithRequestID(ctx, requestID)
}
