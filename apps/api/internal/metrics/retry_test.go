package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRetryMetricsExposeExpectedFamilies(t *testing.T) {
	ExternalRetriesTotal.Reset()
	ExternalRetryDelay.Reset()
	ExternalRetryExhaustedTotal.Reset()

	registry := prometheus.NewRegistry()

	for _, collector := range []prometheus.Collector{
		ExternalRetriesTotal,
		ExternalRetryDelay,
		ExternalRetryExhaustedTotal,
	} {
		if err := registry.Register(collector); err != nil {
			t.Fatalf("register retry collector: %v", err)
		}
	}

	ExternalRetriesTotal.
		WithLabelValues("reddit", "fetch", "upstream").
		Inc()

	ExternalRetryDelay.
		WithLabelValues("reddit", "fetch").
		Observe(0.025)

	ExternalRetryExhaustedTotal.
		WithLabelValues("reddit", "fetch").
		Inc()

	if got := testutil.ToFloat64(
		ExternalRetriesTotal.WithLabelValues(
			"reddit",
			"fetch",
			"upstream",
		),
	); got != 1 {
		t.Fatalf("expected one retry, got %v", got)
	}

	if got := testutil.ToFloat64(
		ExternalRetryExhaustedTotal.WithLabelValues(
			"reddit",
			"fetch",
		),
	); got != 1 {
		t.Fatalf("expected one exhaustion, got %v", got)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather retry metrics: %v", err)
	}

	expected := map[string]bool{
		"village_external_retries_total":         false,
		"village_external_retry_delay_seconds":   false,
		"village_external_retry_exhausted_total": false,
	}

	for _, family := range families {
		if _, ok := expected[family.GetName()]; ok {
			expected[family.GetName()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf("expected metric family %q", name)
		}
	}
}

func TestRetryMetricsUseBoundedLabels(t *testing.T) {
	ExternalRetriesTotal.Reset()

	registry := prometheus.NewRegistry()
	if err := registry.Register(ExternalRetriesTotal); err != nil {
		t.Fatalf("register retry metric: %v", err)
	}

	ExternalRetriesTotal.
		WithLabelValues("reddit", "fetch", "timeout").
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather retry metric: %v", err)
	}

	for _, family := range families {
		if family.GetName() != "village_external_retries_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "source", "operation", "error_type":
				default:
					t.Fatalf("unexpected retry label %q", label.GetName())
				}
			}
		}
	}
}
