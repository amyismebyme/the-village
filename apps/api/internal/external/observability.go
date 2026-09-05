package external

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
)

// ObserveOperation records the logical external operation and its
// corresponding structured log.
//
// Physical HTTP request metrics are recorded by ObserveRequestAttempt.
// Keeping those responsibilities separate prevents retry attempts from
// being double-counted.
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
	errorClass := ClassifyError(err)

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
		"request_attempted",
		requestAttempted,
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
			errorClass,
		)
	}

	if errors.Is(err, context.Canceled) {
		logger.Info(
			"external integration operation canceled",
			args...,
		)

		return
	}

	logger.Info(
		"external integration operation completed",
		args...,
	)
}

// ObserveRequestAttempt records metrics for one physical external
// request attempt.
//
// This function is intentionally the sole owner of:
//
//   - ExternalRequestsTotal
//   - ExternalRequestDuration
//   - ExternalErrorsTotal
//
// Retry attempts therefore produce one set of physical-request metrics
// each, while ObserveOperation remains a logical-operation observer.
func ObserveRequestAttempt(
	source Source,
	operation string,
	status string,
	start time.Time,
	err error,
) {
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
			time.Since(start).Seconds(),
		)

	if err == nil {
		return
	}

	metrics.ExternalErrorsTotal.
		WithLabelValues(
			string(source),
			operation,
			string(ClassifyError(err)),
		).
		Inc()
}
