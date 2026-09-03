package external

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNilRequest      = errors.New("external request is required")
	ErrNilResponseBody = errors.New("external response body is required")

	ErrUnauthorized   = errors.New("external unauthorized")
	ErrForbidden      = errors.New("external forbidden")
	ErrNotFound       = errors.New("external resource not found")
	ErrRateLimited    = errors.New("external rate limited")
	ErrUpstream       = errors.New("external upstream error")
	ErrTimeout        = errors.New("external request timeout")
	ErrInvalidPayload = errors.New("external invalid payload")
	ErrInvalidConfig  = errors.New("external invalid configuration")

	ErrRetryExhausted = errors.New("external retry exhausted")
)

// ErrorClass is the bounded classification used for external
// integration observability and logging.
//
// Keep this set intentionally small and stable because values are
// exposed through Prometheus labels.
type ErrorClass string

const (
	ErrorClassUnknown        ErrorClass = "unknown"
	ErrorClassCanceled       ErrorClass = "canceled"
	ErrorClassUnauthorized   ErrorClass = "unauthorized"
	ErrorClassForbidden      ErrorClass = "forbidden"
	ErrorClassNotFound       ErrorClass = "not_found"
	ErrorClassRateLimited    ErrorClass = "rate_limited"
	ErrorClassUpstream       ErrorClass = "upstream"
	ErrorClassTimeout        ErrorClass = "timeout"
	ErrorClassInvalidPayload ErrorClass = "invalid_payload"
	ErrorClassInvalidConfig  ErrorClass = "invalid_config"
	ErrorClassRetryExhausted ErrorClass = "retry_exhausted"
)

// ClassifyError converts an external error into a bounded ErrorClass.
//
// The function uses errors.Is so wrapped and joined errors retain their
// meaningful classification.
func ClassifyError(err error) ErrorClass {
	switch {
	case err == nil:
		return ErrorClassUnknown

	case errors.Is(err, ErrRetryExhausted):
		return ErrorClassRetryExhausted

	case errors.Is(err, ErrUnauthorized):
		return ErrorClassUnauthorized

	case errors.Is(err, ErrForbidden):
		return ErrorClassForbidden

	case errors.Is(err, ErrNotFound):
		return ErrorClassNotFound

	case errors.Is(err, ErrRateLimited):
		return ErrorClassRateLimited

	case errors.Is(err, ErrUpstream):
		return ErrorClassUpstream

	case errors.Is(err, ErrTimeout):
		return ErrorClassTimeout

	case errors.Is(err, ErrInvalidPayload):
		return ErrorClassInvalidPayload

	case errors.Is(err, ErrInvalidConfig):
		return ErrorClassInvalidConfig

	case errors.Is(err, context.Canceled):
		return ErrorClassCanceled

	default:
		return ErrorClassUnknown
	}
}

type RetryExhaustedError struct {
	Cause    error
	Attempts int
}

func (e *RetryExhaustedError) Error() string {
	if e == nil {
		return ErrRetryExhausted.Error()
	}

	if e.Cause == nil {
		return fmt.Sprintf(
			"%s after %d attempts",
			ErrRetryExhausted,
			e.Attempts,
		)
	}

	return fmt.Sprintf(
		"%s after %d attempts: %v",
		ErrRetryExhausted,
		e.Attempts,
		e.Cause,
	)
}

func (e *RetryExhaustedError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

func (e *RetryExhaustedError) Is(target error) bool {
	return target == ErrRetryExhausted
}

func IsRetryExhausted(err error) bool {
	return errors.Is(err, ErrRetryExhausted)
}

type RateLimitError struct {
	Cause      error
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause == nil {
		return ErrRateLimited.Error()
	}

	return e.Cause.Error()
}

func (e *RateLimitError) Unwrap() error {
	if e == nil {
		return nil
	}

	if e.Cause == nil {
		return ErrRateLimited
	}

	return e.Cause
}

func IsRetryable(err error) bool {
	switch {
	case errors.Is(err, ErrRateLimited):
		return true

	case errors.Is(err, ErrUpstream):
		return true

	case errors.Is(err, ErrTimeout):
		return true

	default:
		return false
	}
}

func IsPermanent(err error) bool {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return true

	case errors.Is(err, ErrForbidden):
		return true

	case errors.Is(err, ErrNotFound):
		return true

	case errors.Is(err, ErrInvalidPayload):
		return true

	case errors.Is(err, ErrInvalidConfig):
		return true

	default:
		return false
	}
}

func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsInvalidPayload(err error) bool {
	return errors.Is(err, ErrInvalidPayload)
}

func IsInvalidConfig(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

func IsUpstream(err error) bool {
	return errors.Is(err, ErrUpstream)
}
