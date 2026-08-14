package postgres

import (
	"time"

	appmetrics "github.com/amyismebyme/the-village/apps/api/internal/metrics"
)

func observeQuery(
	operation string,
	start time.Time,
	err error,
) {
	status := "success"

	if err != nil {
		status = "failure"
	}

	appmetrics.DatabaseQueriesTotal.
		WithLabelValues(
			operation,
			status,
		).
		Inc()

	appmetrics.DatabaseQueryDuration.
		WithLabelValues(
			operation,
		).
		Observe(
			time.Since(start).Seconds(),
		)
}
