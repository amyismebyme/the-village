package worker

import (
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"testing"
)

func TestWorkerFailureClassificationMetric(t *testing.T) {
	metrics.WorkerFailureTypesTotal.Reset()

	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics.WorkerFailureTypesTotal)

	metrics.WorkerFailureTypesTotal.
		WithLabelValues("reddit_ingestion", "unknown").
		Inc()

	if got := testutil.ToFloat64(
		metrics.WorkerFailureTypesTotal.WithLabelValues(
			"reddit_ingestion",
			"unknown",
		),
	); got != 1 {
		t.Fatalf("expected one classified worker failure, got %v", got)
	}

	_ = errors.New("ensure errors import remains intentional")
}

func TestWorkerMetricsRegistrationIncludesFailureTypes(t *testing.T) {
	registry := prometheus.NewRegistry()

	if err := metrics.Register(registry, nil); err != nil {
		t.Fatalf("register metrics: %v", err)
	}

	metrics.WorkerFailureTypesTotal.
		WithLabelValues("reddit_ingestion", "upstream").
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	var found bool
	for _, family := range families {
		if family.GetName() == "village_worker_failures_by_type_total" {
			found = true
		}
	}

	if !found {
		t.Fatal("worker failure classification metric not found")
	}
}

var _ = prometheus.Labels{}
