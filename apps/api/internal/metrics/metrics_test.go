package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestBuildInfoMetric(t *testing.T) {
	registry := prometheus.NewRegistry()

	BuildInfo.Reset()

	if err := registry.Register(BuildInfo); err != nil {
		t.Fatalf(
			"register build info metric: %v",
			err,
		)
	}

	BuildInfo.With(prometheus.Labels{
		"environment": "dev",
		"git_commit":  "local",
		"go_version":  "go1.26.5",
		"version":     "0.7.8",
	}).Set(1)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather build info metric: %v",
			err,
		)
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

			delete(
				expectedLabels,
				label.GetName(),
			)
		}

		if len(expectedLabels) != 0 {
			t.Fatalf(
				"missing expected labels: %v",
				expectedLabels,
			)
		}
	}

	if !found {
		t.Fatal(
			"village_build_info metric not found",
		)
	}
}

func TestPoolCollector(t *testing.T) {
	t.Skip(
		"requires a live PostgreSQL connection",
	)
}

func TestCommunityMetricsRegistration(t *testing.T) {
	tests := []struct {
		name      string
		collector prometheus.Collector
		metric    string
		labels    prometheus.Labels
	}{
		{
			name:      "create",
			collector: CommunityCreateTotal,
			metric:    "village_community_create_total",
			labels: prometheus.Labels{
				"status": "success",
			},
		},
		{
			name:      "update",
			collector: CommunityUpdateTotal,
			metric:    "village_community_update_total",
			labels: prometheus.Labels{
				"status": "success",
			},
		},
		{
			name:      "delete",
			collector: CommunityDeleteTotal,
			metric:    "village_community_delete_total",
			labels: prometheus.Labels{
				"status": "success",
			},
		},
		{
			name:      "validation failures",
			collector: CommunityValidationFailuresTotal,
			metric:    "village_community_validation_failures_total",
			labels: prometheus.Labels{
				"field": "name",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			resetCommunityMetrics()

			registry := prometheus.NewRegistry()

			if err := registry.Register(tt.collector); err != nil {
				t.Fatalf(
					"register %s metric: %v",
					tt.metric,
					err,
				)
			}

			switch collector := tt.collector.(type) {
			case *prometheus.CounterVec:
				collector.With(tt.labels).Add(0)

			default:
				t.Fatalf(
					"expected %s to be a CounterVec, got %T",
					tt.metric,
					tt.collector,
				)
			}

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf(
					"gather %s metric: %v",
					tt.metric,
					err,
				)
			}

			var found bool

			for _, family := range families {
				if family.GetName() != tt.metric {
					continue
				}

				found = true

				if family.GetType().String() != "COUNTER" {
					t.Fatalf(
						"expected %s to be COUNTER, got %s",
						tt.metric,
						family.GetType().String(),
					)
				}
			}

			if !found {
				t.Fatalf(
					"metric family %q not found",
					tt.metric,
				)
			}
		})
	}
}

func TestCommunityOperationMetricLabelsAreBounded(t *testing.T) {
	tests := []struct {
		name      string
		collector *prometheus.CounterVec
		metric    string
	}{
		{
			name:      "create",
			collector: CommunityCreateTotal,
			metric:    "village_community_create_total",
		},
		{
			name:      "update",
			collector: CommunityUpdateTotal,
			metric:    "village_community_update_total",
		},
		{
			name:      "delete",
			collector: CommunityDeleteTotal,
			metric:    "village_community_delete_total",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			resetCommunityMetrics()

			registry := prometheus.NewRegistry()

			if err := registry.Register(tt.collector); err != nil {
				t.Fatalf(
					"register %s metric: %v",
					tt.metric,
					err,
				)
			}

			tt.collector.
				WithLabelValues("success").
				Inc()

			tt.collector.
				WithLabelValues("failure").
				Inc()

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf(
					"gather %s metric: %v",
					tt.metric,
					err,
				)
			}

			var found bool

			for _, family := range families {
				if family.GetName() != tt.metric {
					continue
				}

				found = true

				if len(family.GetMetric()) != 2 {
					t.Fatalf(
						"expected 2 %s samples, got %d",
						tt.metric,
						len(family.GetMetric()),
					)
				}

				for _, metric := range family.GetMetric() {
					for _, label := range metric.GetLabel() {
						if label.GetName() != "status" {
							t.Fatalf(
								"unexpected label %q on %s",
								label.GetName(),
								tt.metric,
							)
						}

						switch label.GetValue() {
						case "success", "failure":
							// valid bounded value

						default:
							t.Fatalf(
								"unexpected status label value %q",
								label.GetValue(),
							)
						}
					}
				}
			}

			if !found {
				t.Fatalf(
					"metric family %q not found",
					tt.metric,
				)
			}
		})
	}
}

func TestCommunityValidationMetricUsesBoundedFieldLabel(t *testing.T) {
	resetCommunityMetrics()

	registry := prometheus.NewRegistry()

	if err := registry.Register(
		CommunityValidationFailuresTotal,
	); err != nil {
		t.Fatalf(
			"register validation metric: %v",
			err,
		)
	}

	fields := []string{
		"name",
		"slug",
		"description",
		"external_source",
	}

	for _, field := range fields {
		CommunityValidationFailuresTotal.
			WithLabelValues(field).
			Inc()
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather validation metric: %v",
			err,
		)
	}

	var found bool

	for _, family := range families {
		if family.GetName() != "village_community_validation_failures_total" {
			continue
		}

		found = true

		if family.GetType().String() != "COUNTER" {
			t.Fatalf(
				"expected validation metric to be COUNTER, got %s",
				family.GetType().String(),
			)
		}

		if len(family.GetMetric()) != len(fields) {
			t.Fatalf(
				"expected %d validation samples, got %d",
				len(fields),
				len(family.GetMetric()),
			)
		}

		expected := map[string]bool{
			"name":            false,
			"slug":            false,
			"description":     false,
			"external_source": false,
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() != "field" {
					t.Fatalf(
						"unexpected label %q on validation metric",
						label.GetName(),
					)
				}

				if _, ok := expected[label.GetValue()]; !ok {
					t.Fatalf(
						"unexpected validation field label %q",
						label.GetValue(),
					)
				}

				expected[label.GetValue()] = true
			}
		}

		for field, seen := range expected {
			if !seen {
				t.Fatalf(
					"missing validation field label %q",
					field,
				)
			}
		}
	}

	if !found {
		t.Fatal(
			"village_community_validation_failures_total metric not found",
		)
	}
}

func TestMetricsRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()

	Register(registry, nil)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	exposed := make(map[string]bool, len(families))
	for _, family := range families {
		exposed[family.GetName()] = true
	}

	if !exposed["village_build_info"] {
		t.Fatal("expected village_build_info to be exposed")
	}
}

func resetCommunityMetrics() {
	CommunityCreateTotal.Reset()
	CommunityUpdateTotal.Reset()
	CommunityDeleteTotal.Reset()
	CommunityValidationFailuresTotal.Reset()
}

func TestDatabaseQueriesMetric(t *testing.T) {
	registry := prometheus.NewRegistry()

	if err := registry.Register(DatabaseQueriesTotal); err != nil {
		t.Fatalf("register database queries metric: %v", err)
	}

	DatabaseQueriesTotal.
		WithLabelValues("select", "success").
		Add(1)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather database queries metric: %v", err)
	}

	var found bool

	for _, family := range families {
		if family.GetName() != "village_db_queries_total" {
			continue
		}

		found = true

		if family.GetType().String() != "COUNTER" {
			t.Fatalf(
				"expected COUNTER, got %s",
				family.GetType().String(),
			)
		}

		if len(family.GetMetric()) != 1 {
			t.Fatalf(
				"expected 1 metric, got %d",
				len(family.GetMetric()),
			)
		}
	}

	if !found {
		t.Fatal("village_db_queries_total metric not found")
	}
}
