package external

import "context"

// RateLimiter controls the start of outbound requests.
//
// Wait blocks until the caller may begin the next request. Waiting
// must always honor the supplied context.
type RateLimiter interface {
	Wait(ctx context.Context) error
}
