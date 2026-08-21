package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()

	Register(registry, nil)

	// Gather should succeed without registration errors.
	if _, err := registry.Gather(); err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
}

func TestBuildInfoMetric(t *testing.T) {
	registry := prometheus.NewRegistry()

	// BuildInfo is a package-level GaugeVec and may already contain
	// a metric created during package initialization.
	BuildInfo.Reset()

	if err := registry.Register(BuildInfo); err != nil {
		t.Fatalf("register build info metric: %v", err)
	}

	// Use label names instead of positional values so this test does
	// not depend on the declaration order of BuildInfo's labels.
	BuildInfo.With(prometheus.Labels{
		"environment": "dev",
		"git_commit":  "local",
		"go_version":  "go1.26.5",
		"version":     "0.7.8",
	}).Set(1)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather build info metric: %v", err)
	}

	var found bool

	for _, family := range families {
		if family.GetName() != "village_build_info" {
			continue
		}

		found = true

		if family.GetType().String() != "GAUGE" {
			t.Fatalf(
				"expected gauge, got %s",
				family.GetType().String(),
			)
		}

		if len(family.GetMetric()) != 1 {
			t.Fatalf(
				"expected 1 metric sample, got %d",
				len(family.GetMetric()),
			)
		}

		metric := family.GetMetric()[0]

		if metric.GetGauge().GetValue() != 1 {
			t.Fatalf(
				"expected build info value 1, got %v",
				metric.GetGauge().GetValue(),
			)
		}

		expectedLabels := map[string]string{
			"environment": "dev",
			"git_commit":  "local",
			"go_version":  "go1.26.5",
			"version":     "0.7.8",
		}

		if len(metric.GetLabel()) != len(expectedLabels) {
			t.Fatalf(
				"expected %d labels, got %d",
				len(expectedLabels),
				len(metric.GetLabel()),
			)
		}

		for _, label := range metric.GetLabel() {
			expected, ok := expectedLabels[label.GetName()]
			if !ok {
				t.Fatalf(
					"unexpected label %q",
					label.GetName(),
				)
			}

			if label.GetValue() != expected {
				t.Fatalf(
					"label %q: expected %q, got %q",
					label.GetName(),
					expected,
					label.GetValue(),
				)
			}

			delete(expectedLabels, label.GetName())
		}

		if len(expectedLabels) != 0 {
			t.Fatalf(
				"missing expected labels: %v",
				expectedLabels,
			)
		}
	}

	if !found {
		t.Fatal("village_build_info metric not found")
	}
}

func TestPoolCollector(t *testing.T) {
	t.Skip("requires a live PostgreSQL connection")
}
