package external

import (
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
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
	start time.Time,
	err error,
) {
	duration := time.Since(start)

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
		Observe(duration.Seconds())

	if err != nil {
		errorType := classifyErrorType(err)

		metrics.ExternalErrorsTotal.
			WithLabelValues(
				string(source),
				operation,
				errorType,
			).
			Inc()
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
	case IsRetryable(err):
		switch {
		case IsRateLimited(err):
			return "rate_limited"

		case IsTimeout(err):
			return "timeout"

		default:
			return "upstream"
		}

	case IsPermanent(err):
		switch {
		case IsUnauthorized(err):
			return "unauthorized"

		case IsForbidden(err):
			return "forbidden"

		case IsInvalidPayload(err):
			return "invalid_payload"

		case IsInvalidConfig(err):
			return "invalid_config"

		default:
			return "permanent"
		}

	default:
		return "unknown"
	}
}
