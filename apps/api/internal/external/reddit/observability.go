package reddit

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
)

func observeOperation(
	logger *slog.Logger,
	operation string,
	externalID string,
	status string,
	requestAttempted bool,
	start time.Time,
	err error,
) {
	external.ObserveOperation(
		logger,
		external.SourceReddit,
		operation,
		externalID,
		status,
		requestAttempted,
		start,
		err,
	)

	// Retry exhaustion is a logical operation event. Count it here,
	// exactly once, when the operation ultimately returns the exhausted
	// error.
	if external.IsRetryExhausted(err) {
		observeRetryExhausted(
			logger,
			operation,
		)
	}
}

func observeRequestAttempt(
	operation string,
	status string,
	start time.Time,
	err error,
) {
	external.ObserveRequestAttempt(
		external.SourceReddit,
		operation,
		status,
		start,
		err,
	)
}

func observeRetry(
	logger *slog.Logger,
	operation string,
	event external.RetryEvent,
) {
	// NextAttempt == 0 represents terminal retry exhaustion.
	//
	// Do not increment ExternalRetryExhaustedTotal here. The final
	// logical operation error is observed by observeOperation(), which
	// owns the single exhaustion metric increment.
	if event.NextAttempt == 0 {
		return
	}

	// This event represents an actual scheduled retry.
	metrics.ExternalRetriesTotal.
		WithLabelValues(
			string(external.SourceReddit),
			operation,
			event.ErrorType,
		).
		Inc()

	metrics.ExternalRetryDelay.
		WithLabelValues(
			string(external.SourceReddit),
			operation,
		).
		Observe(
			event.Delay.Seconds(),
		)

	if logger == nil {
		return
	}

	logger.Warn(
		"external request retry scheduled",
		"source",
		external.SourceReddit,
		"operation",
		operation,
		"attempt",
		event.Attempt,
		"next_attempt",
		event.NextAttempt,
		"delay_ms",
		event.Delay.Milliseconds(),
		"error_type",
		event.ErrorType,
	)
}

func observeRetryExhausted(
	logger *slog.Logger,
	operation string,
) {
	metrics.ExternalRetryExhaustedTotal.
		WithLabelValues(
			string(external.SourceReddit),
			operation,
		).
		Inc()

	if logger == nil {
		return
	}

	logger.Error(
		"external request retry budget exhausted",
		"source",
		external.SourceReddit,
		"operation",
		operation,
	)
}

func redditStatusFromError(err error) string {
	switch {
	case external.IsRetryExhausted(err):
		return "retry_exhausted"

	case errors.Is(
		err,
		external.ErrUnauthorized,
	):
		return "401"

	case errors.Is(
		err,
		external.ErrForbidden,
	):
		return "403"

	case errors.Is(
		err,
		external.ErrNotFound,
	):
		return "404"

	case errors.Is(
		err,
		external.ErrRateLimited,
	):
		return "429"

	case errors.Is(
		err,
		external.ErrTimeout,
	):
		return "timeout"

	case errors.Is(
		err,
		external.ErrUpstream,
	):
		return "5xx"

	case errors.Is(
		err,
		external.ErrInvalidPayload,
	):
		return "invalid_payload"

	case errors.Is(
		err,
		external.ErrInvalidConfig,
	):
		return "invalid_config"

	case errors.Is(
		err,
		context.Canceled,
	):
		return "canceled"

	default:
		return "error"
	}
}
