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
}

func observeRetry(
	logger *slog.Logger,
	operation string,
	event external.RetryEvent,
) {
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

func redditStatusFromError(
	err error,
) string {
	switch {
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
