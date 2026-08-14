package postgres

import (
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"testing"
	"time"
)

func TestObserveQuerySuccess(t *testing.T) {
	before := testutil.ToFloat64(
		metrics.DatabaseQueriesTotal.WithLabelValues(
			"create",
			"success",
		),
	)

	observeQuery(
		"create",
		time.Now().Add(-10*time.Millisecond),
		nil,
	)

	after := testutil.ToFloat64(
		metrics.DatabaseQueriesTotal.WithLabelValues(
			"create",
			"success",
		),
	)

	if after-before != 1 {
		t.Fatalf(
			"expected query counter to increase by 1, got %v",
			after-before,
		)
	}
}

func TestObserveQueryFailure(t *testing.T) {
	before := testutil.ToFloat64(
		metrics.DatabaseQueriesTotal.WithLabelValues(
			"find_by_id",
			"failure",
		),
	)

	observeQuery(
		"find_by_id",
		time.Now().Add(-10*time.Millisecond),
		errors.New("database unavailable"),
	)

	after := testutil.ToFloat64(
		metrics.DatabaseQueriesTotal.WithLabelValues(
			"find_by_id",
			"failure",
		),
	)

	if after-before != 1 {
		t.Fatalf(
			"expected failed query counter to increase by 1, got %v",
			after-before,
		)
	}
}
