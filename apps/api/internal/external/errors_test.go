package external

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorClass
	}{
		{
			name: "canceled",
			err:  context.Canceled,
			want: ErrorClassCanceled,
		},
		{
			name: "unauthorized",
			err:  ErrUnauthorized,
			want: ErrorClassUnauthorized,
		},
		{
			name: "forbidden",
			err:  ErrForbidden,
			want: ErrorClassForbidden,
		},
		{
			name: "not found",
			err:  ErrNotFound,
			want: ErrorClassNotFound,
		},
		{
			name: "rate limited",
			err:  ErrRateLimited,
			want: ErrorClassRateLimited,
		},
		{
			name: "upstream",
			err:  ErrUpstream,
			want: ErrorClassUpstream,
		},
		{
			name: "timeout",
			err:  ErrTimeout,
			want: ErrorClassTimeout,
		},
		{
			name: "invalid payload",
			err:  ErrInvalidPayload,
			want: ErrorClassInvalidPayload,
		},
		{
			name: "invalid config",
			err:  ErrInvalidConfig,
			want: ErrorClassInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyError(tt.err); got != tt.want {
				t.Fatalf(
					"expected %q, got %q",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestClassifyErrorPreservesWrappedErrors(
	t *testing.T,
) {
	err := errors.Join(
		errors.New("operation failed"),
		ErrRateLimited,
	)

	if got := ClassifyError(err); got !=
		ErrorClassRateLimited {
		t.Fatalf(
			"expected rate_limited, got %q",
			got,
		)
	}
}

func TestIsPermanentIncludesNotFound(
	t *testing.T,
) {
	if !IsPermanent(ErrNotFound) {
		t.Fatal(
			"expected ErrNotFound to be permanent",
		)
	}
}

func TestCancellationIsNotRetryable(
	t *testing.T,
) {
	if IsRetryable(context.Canceled) {
		t.Fatal(
			"context.Canceled must not be retryable",
		)
	}
}

func TestClassifyRetryExhaustion(t *testing.T) {
	err := &RetryExhaustedError{
		Cause:    ErrUpstream,
		Attempts: 3,
	}

	if got := ClassifyError(err); got !=
		ErrorClassRetryExhausted {
		t.Fatalf(
			"expected retry_exhausted, got %q",
			got,
		)
	}

	if !IsRetryExhausted(err) {
		t.Fatal(
			"expected retry exhaustion to be detected",
		)
	}

	if !errors.Is(err, ErrUpstream) {
		t.Fatal(
			"expected original upstream error to be preserved",
		)
	}
}
