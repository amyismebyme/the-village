package external

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestObserveOperationSuccess(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{},
		),
	)

	registry := prometheus.NewRegistry()

	registerExternalCollectors(
		t,
		registry,
	)

	metrics.ExternalRequestsTotal.
		WithLabelValues(
			"test",
			"fetch",
			"200",
		).
		Add(0)

	start := time.Now()

	ObserveOperation(
		logger,
		Source("test"),
		"fetch",
		"external-123",
		"200",
		true,
		start,
		nil,
	)

	assertMetricFamily(t, registry, "village_external_requests_total")
	assertMetricFamily(
		t,
		registry,
		"village_external_request_duration_seconds",
	)

	output := logs.String()

	for _, expected := range []string{
		"source=test",
		"operation=fetch",
		"external_id=external-123",
		"status=200",
		"duration_ms=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestObserveOperationTimeout(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{},
		),
	)

	registry := prometheus.NewRegistry()

	registerExternalCollectors(
		t,
		registry,
	)

	err := errors.Join(
		ErrTimeout,
		context.DeadlineExceeded,
	)

	ObserveOperation(
		logger,
		Source("test"),
		"fetch",
		"external-timeout",
		"timeout",
		true,
		time.Now(),
		err,
	)

	assertMetricFamily(
		t,
		registry,
		"village_external_requests_total",
	)

	assertMetricFamily(
		t,
		registry,
		"village_external_errors_total",
	)

	output := logs.String()

	for _, expected := range []string{
		"source=test",
		"operation=fetch",
		"external_id=external-timeout",
		"status=timeout",
		"error_type=timeout",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestExternalTestServer(t *testing.T) {
	server := testutil.NewServer(
		testutil.Response{
			StatusCode: http.StatusOK,
			Body:       `{"ok":true}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
	)
	defer server.Close()

	client := NewClient(
		server.Server.Client(),
		time.Second,
	)

	req, err := http.NewRequest(
		http.MethodGet,
		server.URL()+"/example",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	response, err := client.Do(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf(
			"external request: %v",
			err,
		)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf(
				"close response body: %v",
				err,
			)
		}
	}()

	count, method, path, headers := server.Snapshot()

	if count != 1 {
		t.Fatalf(
			"expected 1 request, got %d",
			count,
		)
	}

	if method != http.MethodGet {
		t.Fatalf(
			"expected GET, got %s",
			method,
		)
	}

	if path != "/example" {
		t.Fatalf(
			"expected /example, got %s",
			path,
		)
	}

	if headers == nil {
		t.Fatal("expected request headers")
	}
}

func registerExternalCollectors(
	t *testing.T,
	registry *prometheus.Registry,
) {
	t.Helper()

	collectors := []prometheus.Collector{
		metrics.ExternalRequestsTotal,
		metrics.ExternalRequestDuration,
		metrics.ExternalErrorsTotal,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			t.Fatalf(
				"register external collector: %v",
				err,
			)
		}
	}
}

func assertMetricFamily(
	t *testing.T,
	registry *prometheus.Registry,
	name string,
) {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() == name {
			return
		}
	}

	t.Fatalf(
		"metric family %q not found",
		name,
	)
}

func TestObserveOperationDoesNotCountUnattemptedRequest(
	t *testing.T,
) {
	registry := prometheus.NewRegistry()

	registerExternalCollectors(
		t,
		registry,
	)

	metrics.ExternalRequestsTotal.Reset()

	ObserveOperation(
		nil,
		Source("test"),
		"fetch",
		"",
		"invalid_config",
		false,
		time.Now(),
		ErrInvalidConfig,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() !=
			"village_external_requests_total" {
			continue
		}

		if len(family.GetMetric()) != 0 {
			t.Fatal(
				"unattempted operation should not expose request metric",
			)
		}
	}
}

func TestObserveOperationRecordsRateLimitError(
	t *testing.T,
) {
	metrics.ExternalRequestsTotal.Reset()
	metrics.ExternalRequestDuration.Reset()
	metrics.ExternalErrorsTotal.Reset()

	registry := prometheus.NewRegistry()

	registerExternalCollectors(
		t,
		registry,
	)

	start := time.Now()

	ObserveOperation(
		nil,
		Source("reddit"),
		"fetch",
		"",
		"429",
		true,
		start,
		ErrRateLimited,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	var found bool

	for _, family := range families {
		if family.GetName() !=
			"village_external_errors_total" {
			continue
		}

		found = true

		if len(family.GetMetric()) != 1 {
			t.Fatalf(
				"expected 1 error metric sample, got %d",
				len(family.GetMetric()),
			)
		}

		metric := family.GetMetric()[0]

		if metric.GetCounter().GetValue() != 1 {
			t.Fatalf(
				"expected error counter 1, got %v",
				metric.GetCounter().GetValue(),
			)
		}

		labels := make(
			map[string]string,
		)

		for _, label := range metric.GetLabel() {
			labels[label.GetName()] =
				label.GetValue()
		}

		if labels["source"] != "reddit" {
			t.Fatalf(
				"expected source=reddit, got %q",
				labels["source"],
			)
		}

		if labels["operation"] != "fetch" {
			t.Fatalf(
				"expected operation=fetch, got %q",
				labels["operation"],
			)
		}

		if labels["type"] != "rate_limited" {
			t.Fatalf(
				"expected type=rate_limited, got %q",
				labels["type"],
			)
		}
	}

	if !found {
		t.Fatal(
			"metric family village_external_errors_total not found",
		)
	}
}
