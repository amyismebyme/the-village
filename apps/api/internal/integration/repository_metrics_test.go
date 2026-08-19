//go:build integration

package integration

import (
	"context"
	"testing"

	appmetrics "github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestCommunityRepositoryDBQueryDuration(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	registry := prometheus.NewRegistry()

	registry.MustRegister(
		appmetrics.DatabaseQueryDuration,
	)

	before := repositoryQueryDurationCount(
		t,
		registry,
		"find_by_id",
	)

	// The test database does not need to contain ID 1.
	// FindByID still executes a database query and therefore must
	// produce a duration observation regardless of whether the row exists.
	_, err := app.repo.FindByID(
		context.Background(),
		1,
	)
	_ = err

	after := repositoryQueryDurationCount(
		t,
		registry,
		"find_by_id",
	)

	if after != before+1 {
		t.Fatalf(
			"expected find_by_id duration sample to increase by 1; before=%d after=%d",
			before,
			after,
		)
	}
}

func repositoryQueryDurationCount(
	t *testing.T,
	registry *prometheus.Registry,
	operation string,
) uint64 {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() != "village_db_query_duration_seconds" {
			continue
		}

		for _, metric := range family.GetMetric() {
			if metric.Histogram == nil {
				continue
			}

			for _, label := range metric.GetLabel() {
				if label.GetName() != "operation" {
					continue
				}

				if label.GetValue() != operation {
					continue
				}

				return metric.Histogram.GetSampleCount()
			}
		}
	}

	return 0
}