package external

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type RetryEvent struct {
	Attempt     int
	NextAttempt int
	Delay       time.Duration
	ErrorType   string
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     Backoff
}

func NewRetryPolicy(
	maxAttempts int,
	backoff Backoff,
) (*RetryPolicy, error) {
	if maxAttempts < 1 {
		return nil, errors.New(
			"retry policy: max attempts must be at least 1",
		)
	}

	return &RetryPolicy{
		MaxAttempts: maxAttempts,
		Backoff:     backoff,
	}, nil
}

// Do executes operation and retries only retryable errors.
//
// The observer is called immediately before a retry delay begins.
// Cancellation and deadline expiration are terminal and never cause
// another attempt after the context has become done.
func (p *RetryPolicy) Do(
	ctx context.Context,
	operation func(context.Context) error,
	observer func(RetryEvent),
) error {
	if ctx == nil {
		return errors.New(
			"retry policy: context is required",
		)
	}

	if operation == nil {
		return errors.New(
			"retry policy: operation is required",
		)
	}

	var lastErr error

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := operation(ctx)

		if err == nil {
			return nil
		}

		lastErr = err

		// Once the caller has cancelled or the deadline has expired,
		// retrying would violate the caller's context contract.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		// Permanent error or attempt budget exhausted.
		if !IsRetryable(err) ||
			attempt == p.MaxAttempts {
			return err
		}

		delay := p.Backoff.Delay(attempt)

		var rateLimitErr *RateLimitError

		if errors.As(
			err,
			&rateLimitErr,
		) && rateLimitErr.RetryAfter > 0 {
			delay = rateLimitErr.RetryAfter

			if delay > p.Backoff.Max {
				delay = p.Backoff.Max
			}
		}

		if observer != nil {
			observer(
				RetryEvent{
					Attempt:     attempt,
					NextAttempt: attempt + 1,
					Delay:       delay,
					ErrorType:   classifyErrorType(err),
				},
			)
		}

		if err := waitForRetry(
			ctx,
			delay,
		); err != nil {
			return err
		}
	}

	return fmt.Errorf(
		"retry policy: exhausted attempts: %w",
		lastErr,
	)
}

func waitForRetry(
	ctx context.Context,
	delay time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		return ctx.Err()

	case <-timer.C:
		return nil
	}
}
