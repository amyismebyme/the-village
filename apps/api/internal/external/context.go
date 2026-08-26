package external

import (
	"context"
	"time"
)

// WithRequestTimeout derives a child context for an external operation.
//
// The caller's context remains authoritative:
//
//   - cancellation propagates immediately
//   - an existing earlier deadline wins
//   - timeout cannot extend the caller's deadline
func WithRequestTimeout(
	parent context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	if timeout <= 0 {
		return context.WithCancel(parent)
	}

	if deadline, ok := parent.Deadline(); ok {
		remaining := time.Until(deadline)

		if remaining <= 0 {
			return context.WithCancel(parent)
		}

		if remaining < timeout {
			return context.WithTimeout(parent, remaining)
		}
	}

	return context.WithTimeout(parent, timeout)
}
