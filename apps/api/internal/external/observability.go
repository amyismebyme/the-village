package external

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"log/slog"
	"time"
)

// ObserveOperation records the bounded Prometheus metrics and the
// corresponding structured log for one external operation.
//
// Metrics intentionally do not contain external_id or request_id.
// Those belong in logs, where they are useful for correlation but
// are not used as high-cardinality Prometheus dimensions.
func ObserveOperation(
	logger *slog.Logger,
	source Source,
	operation string,
	externalID string,
	status string,
	requestAttempted bool,
	start time.Time,
	err error,
) {
	duration := time.Since(start)

	if requestAttempted {
		metrics.ExternalRequestsTotal.
			WithLabelValues(
				string(source),
				operation,
				status,
			).
			Inc()

		metrics.ExternalRequestDuration.
			WithLabelValues(
				string(source),
				operation,
			).
			Observe(
				duration.Seconds(),
			)

		if err != nil {
			metrics.ExternalErrorsTotal.
				WithLabelValues(
					string(source),
					operation,
					classifyErrorType(err),
				).
				Inc()
		}
	}

	if logger == nil {
		return
	}

	args := []any{
		"source",
		source,
		"operation",
		operation,
		"status",
		status,
		"duration_ms",
		duration.Milliseconds(),
	}

	if externalID != "" {
		args = append(
			args,
			"external_id",
			externalID,
		)
	}

	if err != nil {
		args = append(
			args,
			"error_type",
			classifyErrorType(err),
		)
	}

	logger.Info(
		"external integration operation completed",
		args...,
	)
}

func classifyErrorType(err error) string {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return "unauthorized"

	case errors.Is(err, ErrForbidden):
		return "forbidden"

	case errors.Is(err, ErrNotFound):
		return "not_found"

	case errors.Is(err, ErrRateLimited):
		return "rate_limited"

	case errors.Is(err, ErrTimeout):
		return "timeout"

	case errors.Is(err, ErrInvalidPayload):
		return "invalid_payload"

	case errors.Is(err, context.Canceled):
		return "canceled"

	case errors.Is(err, ErrUpstream):
		return "upstream"

	case errors.Is(err, ErrInvalidConfig):
		return "invalid_config"

	default:
		return "unknown"
	}
}
