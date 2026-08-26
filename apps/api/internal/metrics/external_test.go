package metrics

//Test village_external_requests_total
//village_external_request_duration_seconds
//village_external_errors_total
//
import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestExternalMetricsExposeAfterObservation(t *testing.T) {
	registry := prometheus.NewRegistry()

	collectors := []prometheus.Collector{
		ExternalRequestsTotal,
		ExternalRequestDuration,
		ExternalErrorsTotal,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			t.Fatalf(
				"register external collector: %v",
				err,
			)
		}
	}

	ExternalRequestsTotal.
		WithLabelValues(
			"reddit",
			"fetch",
			"200",
		).
		Inc()

	ExternalRequestDuration.
		WithLabelValues(
			"reddit",
			"fetch",
		).
		Observe(0.125)

	ExternalErrorsTotal.
		WithLabelValues(
			"reddit",
			"fetch",
			"timeout",
		).
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather external metrics: %v",
			err,
		)
	}

	expected := map[string]bool{
		"village_external_requests_total":           false,
		"village_external_request_duration_seconds": false,
		"village_external_errors_total":             false,
	}

	for _, family := range families {
		if _, ok := expected[family.GetName()]; ok {
			expected[family.GetName()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf(
				"expected metric family %q after observation",
				name,
			)
		}
	}
}

func TestExternalRequestMetricUsesBoundedLabels(t *testing.T) {
	registry := prometheus.NewRegistry()

	if err := registry.Register(
		ExternalRequestsTotal,
	); err != nil {
		t.Fatalf(
			"register external request metric: %v",
			err,
		)
	}

	ExternalRequestsTotal.
		WithLabelValues(
			"reddit",
			"fetch",
			"200",
		).
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather external request metric: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() != "village_external_requests_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "source", "operation", "status":
					// expected bounded labels
				default:
					t.Fatalf(
						"unexpected label %q on external request metric",
						label.GetName(),
					)
				}
			}
		}
	}
}

func TestExternalErrorMetricUsesBoundedLabels(t *testing.T) {
	registry := prometheus.NewRegistry()

	if err := registry.Register(
		ExternalErrorsTotal,
	); err != nil {
		t.Fatalf(
			"register external error metric: %v",
			err,
		)
	}

	ExternalErrorsTotal.
		WithLabelValues(
			"reddit",
			"fetch",
			"timeout",
		).
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather external error metric: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() != "village_external_errors_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "source", "operation", "type":
					// expected bounded labels
				default:
					t.Fatalf(
						"unexpected label %q on external error metric",
						label.GetName(),
					)
				}
			}
		}
	}
}
