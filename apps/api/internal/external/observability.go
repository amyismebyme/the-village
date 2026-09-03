package external

import (
	"context"
	"errors"
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
	requestAttempted bool,
	start time.Time,
	err error,
) {
	duration := time.Since(start)
	errorClass := ClassifyError(err)

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
					string(errorClass),
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
