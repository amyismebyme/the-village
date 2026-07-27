package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterDoesNotPanic(t *testing.T) {

	registry := prometheus.NewRegistry()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Register panicked: %v", r)
		}
	}()

	Register(registry)
}

/*
func TestBuildInfoMetricRegistered(t *testing.T) {

	registry := prometheus.NewRegistry()

	Register(registry)

	metrics, err := registry.Gather()

	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, metric := range metrics {

		if metric.GetName() == "village_build_info" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("village_build_info not registered")
	}
}
*/
/*
func TestBuildInfoMetricValue(t *testing.T) {

	registry := prometheus.NewRegistry()

	Register(registry)

	metrics, err := registry.Gather()

	if err != nil {
		t.Fatal(err)
	}

	for _, metric := range metrics {

		if metric.GetName() != "village_build_info" {
			continue
		}

		if len(metric.Metric) != 1 {
			t.Fatal("expected exactly one metric")
		}

		value := metric.Metric[0].GetGauge().GetValue()

		if value != 1 {
			t.Fatalf("expected value 1 got %v", value)
		}

		return
	}

	t.Fatal("build info metric missing")
}*/
