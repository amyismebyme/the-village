package external

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryPolicyRetriesRetryableError(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		time.Millisecond,
		5*time.Millisecond,
	)

	var attempts int

	err := policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++

			if attempts < 3 {
				return ErrUpstream
			}

			return nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf(
			"expected success, got %v",
			err,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"expected 3 attempts, got %d",
			attempts,
		)
	}
}

func TestRetryPolicyDoesNotRetryPermanentError(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		time.Millisecond,
		5*time.Millisecond,
	)

	var attempts int

	err := policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++
			return ErrUnauthorized
		},
		nil,
	)

	if !errors.Is(
		err,
		ErrUnauthorized,
	) {
		t.Fatalf(
			"expected ErrUnauthorized, got %v",
			err,
		)
	}

	if attempts != 1 {
		t.Fatalf(
			"expected one attempt, got %d",
			attempts,
		)
	}
}

func TestRetryPolicyAlreadyCanceled(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		50*time.Millisecond,
		100*time.Millisecond,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	var attempts int

	err := policy.Do(
		ctx,
		func(context.Context) error {
			attempts++
			return ErrUpstream
		},
		nil,
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if attempts != 0 {
		t.Fatalf(
			"expected zero attempts, got %d",
			attempts,
		)
	}
}

func TestRetryPolicyCancellationAfterFailedAttempt(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		100*time.Millisecond,
		time.Second,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	var attempts int

	err := policy.Do(
		ctx,
		func(context.Context) error {
			attempts++

			cancel()

			return ErrUpstream
		},
		nil,
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if attempts != 1 {
		t.Fatalf(
			"expected one attempt, got %d",
			attempts,
		)
	}
}

func TestRetryPolicyCancellationDuringBackoff(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		500*time.Millisecond,
		time.Second,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	defer cancel()

	var attempts int

	done := make(chan error, 1)

	go func() {
		done <- policy.Do(
			ctx,
			func(context.Context) error {
				attempts++
				return ErrUpstream
			},
			nil,
		)
	}()

	// Allow the first attempt to enter backoff.
	time.Sleep(20 * time.Millisecond)

	cancel()

	select {
	case err := <-done:
		if !errors.Is(
			err,
			context.Canceled,
		) {
			t.Fatalf(
				"expected context.Canceled, got %v",
				err,
			)
		}

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"retry policy did not stop during backoff",
		)
	}

	if attempts != 1 {
		t.Fatalf(
			"expected one attempt, got %d",
			attempts,
		)
	}
}

func TestRetryPolicyEmitsRetryEvent(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		time.Millisecond,
		5*time.Millisecond,
	)

	var (
		attempts int
		events   []RetryEvent
	)

	err := policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++

			if attempts == 1 {
				return ErrUpstream
			}

			return nil
		},
		func(event RetryEvent) {
			events = append(
				events,
				event,
			)
		},
	)

	if err != nil {
		t.Fatalf(
			"expected success, got %v",
			err,
		)
	}

	if len(events) != 1 {
		t.Fatalf(
			"expected one retry event, got %d",
			len(events),
		)
	}

	event := events[0]

	if event.Attempt != 1 {
		t.Fatalf(
			"expected attempt=1, got %d",
			event.Attempt,
		)
	}

	if event.NextAttempt != 2 {
		t.Fatalf(
			"expected next attempt=2, got %d",
			event.NextAttempt,
		)
	}

	if event.ErrorType != "upstream" {
		t.Fatalf(
			"expected error_type=upstream, got %q",
			event.ErrorType,
		)
	}
}

func TestRetryPolicyHonorsRetryAfter(
	t *testing.T,
) {
	backoff, err := NewBackoff(
		time.Millisecond,
		100*time.Millisecond,
		2,
		0,
	)
	if err != nil {
		t.Fatalf(
			"create backoff: %v",
			err,
		)
	}

	policy, err := NewRetryPolicy(
		2,
		backoff,
	)
	if err != nil {
		t.Fatalf(
			"create retry policy: %v",
			err,
		)
	}

	var attempts int

	start := time.Now()

	err = policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++

			if attempts == 1 {
				return &RateLimitError{
					Cause: ErrRateLimited,
					RetryAfter: 25 *
						time.Millisecond,
				}
			}

			return nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf(
			"expected success, got %v",
			err,
		)
	}

	if attempts != 2 {
		t.Fatalf(
			"expected 2 attempts, got %d",
			attempts,
		)
	}

	elapsed := time.Since(start)

	if elapsed < 20*time.Millisecond {
		t.Fatalf(
			"expected Retry-After to be honored, got %s",
			elapsed,
		)
	}
}

func newTestRetryPolicy(
	t *testing.T,
	maxAttempts int,
	initial time.Duration,
	max time.Duration,
) *RetryPolicy {
	t.Helper()

	backoff, err := NewBackoff(
		initial,
		max,
		2,
		0,
	)
	if err != nil {
		t.Fatalf(
			"create backoff: %v",
			err,
		)
	}

	policy, err := NewRetryPolicy(
		maxAttempts,
		backoff,
	)
	if err != nil {
		t.Fatalf(
			"create retry policy: %v",
			err,
		)
	}

	return policy
}

func TestRetryPolicyReportsExhaustion(
	t *testing.T,
) {
	policy := newTestRetryPolicy(
		t,
		3,
		time.Millisecond,
		5*time.Millisecond,
	)

	var attempts int

	err := policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++
			return ErrUpstream
		},
		nil,
	)

	if !IsRetryExhausted(err) {
		t.Fatalf(
			"expected retry exhaustion, got %v",
			err,
		)
	}

	if !errors.Is(
		err,
		ErrUpstream,
	) {
		t.Fatalf(
			"expected underlying upstream error, got %v",
			err,
		)
	}

	var exhausted *RetryExhaustedError

	if !errors.As(
		err,
		&exhausted,
	) {
		t.Fatalf(
			"expected RetryExhaustedError, got %T",
			err,
		)
	}

	if exhausted.Attempts != 3 {
		t.Fatalf(
			"expected 3 exhausted attempts, got %d",
			exhausted.Attempts,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"expected 3 operation attempts, got %d",
			attempts,
		)
	}
}
