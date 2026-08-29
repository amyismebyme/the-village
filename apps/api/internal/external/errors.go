package external

import (
	"errors"
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
)

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

func IsInvalidPayload(err error) bool {
	return errors.Is(err, ErrInvalidPayload)
}

func IsInvalidConfig(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

func IsUpstream(err error) bool {
	return errors.Is(err, ErrUpstream)
}
