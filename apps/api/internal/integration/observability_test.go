//go:build integration

package integration

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestObservabilityEndToEnd(t *testing.T) {
	var logs bytes.Buffer

	app := newIntegrationAppWithLogger(
		t,
		newTestLogger(&logs),
	)

	route := "/api/v1/communities"

	// Capture metric values before the request.
	beforeRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodGet,
			route,
			"200",
		),
	)

	beforeDuration := histogramSampleCountForRoute(
		t,
		http.MethodGet,
		route,
	)

	response := integrationRequest(
		t,
		app,
		http.MethodGet,
		route,
		"",
	)

	io.Copy(io.Discard, response.Body)
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	// -------------------------------------------------------------------------
	// Request ID
	// -------------------------------------------------------------------------

	requestID := response.Header.Get("X-Request-ID")

	if requestID == "" {
		t.Fatal("expected X-Request-ID header")
	}

	// -------------------------------------------------------------------------
	// HTTP request counter
	// -------------------------------------------------------------------------

	afterRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodGet,
			route,
			"200",
		),
	)

	if got := afterRequests - beforeRequests; got != 1 {
		t.Fatalf(
			"expected HTTP request counter to increase by 1, got %v",
			got,
		)
	}

	// -------------------------------------------------------------------------
	// HTTP request duration
	// -------------------------------------------------------------------------

	afterDuration := histogramSampleCountForRoute(
		t,
		http.MethodGet,
		route,
	)

	if afterDuration != beforeDuration+1 {
		t.Fatalf(
			"expected duration sample count to increase by 1, before=%d after=%d",
			beforeDuration,
			afterDuration,
		)
	}

	// -------------------------------------------------------------------------
	// Normalized route
	// -------------------------------------------------------------------------

	normalizedRoute := "/api/v1/communities"

	if !strings.Contains(
		logs.String(),
		"route="+normalizedRoute,
	) {
		t.Fatalf(
			"expected normalized route in log, got:\n%s",
			logs.String(),
		)
	}

	// -------------------------------------------------------------------------
	// Request ID correlation
	// -------------------------------------------------------------------------

	if !strings.Contains(
		logs.String(),
		"request_id="+requestID,
	) {
		t.Fatalf(
			"expected request ID %q in logs, got:\n%s",
			requestID,
			logs.String(),
		)
	}

	// -------------------------------------------------------------------------
	// Completed request log
	// -------------------------------------------------------------------------

	for _, expected := range []string{
		"method=GET",
		"status=200",
		"duration_ms=",
		"http request completed",
	} {
		if !strings.Contains(logs.String(), expected) {
			t.Fatalf(
				"expected log to contain %q, got:\n%s",
				expected,
				logs.String(),
			)
		}
	}

	// -------------------------------------------------------------------------
	// /metrics endpoint exposes the application metrics
	// -------------------------------------------------------------------------

	metricsResponse, err := app.server.Client().Get(
		app.server.URL + "/metrics",
	)
	if err != nil {
		t.Fatalf(
			"GET /metrics failed: %v",
			err,
		)
	}

	defer metricsResponse.Body.Close()

	if metricsResponse.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected /metrics status %d, got %d",
			http.StatusOK,
			metricsResponse.StatusCode,
		)
	}

	body, err := io.ReadAll(metricsResponse.Body)
	if err != nil {
		t.Fatalf(
			"read /metrics response: %v",
			err,
		)
	}

	metricsText := string(body)

	for _, metricName := range []string{
		"village_http_requests_total",
		"village_http_request_duration_seconds",
		"village_http_requests_in_flight",
	} {
		if !strings.Contains(metricsText, metricName) {
			t.Fatalf(
				"expected /metrics to expose %s",
				metricName,
			)
		}
	}
}

func histogramSampleCountForRoute(
	t *testing.T,
	method string,
	route string,
) uint64 {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() != "village_http_request_duration_seconds" {
			continue
		}

		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)

			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			if labels["method"] != method {
				continue
			}

			if labels["route"] != route {
				continue
			}

			if metric.GetHistogram() == nil {
				t.Fatalf(
					"expected histogram metric for %s %s",
					method,
					route,
				)
			}

			return metric.GetHistogram().GetSampleCount()
		}
	}

	return 0
}

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			buf,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)
}

func TestObservabilityParameterizedRouteNormalization(t *testing.T) {
	var logs bytes.Buffer

	app := newIntegrationAppWithLogger(
		t,
		newTestLogger(&logs),
	)

	route := "/api/v1/communities/{id}"

	beforeRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodGet,
			route,
			"404",
		),
	)

	beforeDuration := histogramSampleCountForRoute(
		t,
		http.MethodGet,
		route,
	)

	requestPath := "/api/v1/communities/999999"

	response := integrationRequest(
		t,
		app,
		http.MethodGet,
		requestPath,
		"",
	)

	io.Copy(io.Discard, response.Body)
	response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.StatusCode,
		)
	}

	afterRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodGet,
			route,
			"404",
		),
	)

	if got := afterRequests - beforeRequests; got != 1 {
		t.Fatalf(
			"expected parameterized route counter to increase by 1, got %v",
			got,
		)
	}

	afterDuration := histogramSampleCountForRoute(
		t,
		http.MethodGet,
		route,
	)

	if afterDuration != beforeDuration+1 {
		t.Fatalf(
			"expected parameterized route duration sample to increase by 1; before=%d after=%d",
			beforeDuration,
			afterDuration,
		)
	}

	logOutput := logs.String()

	if !strings.Contains(
		logOutput,
		"route="+route,
	) {
		t.Fatalf(
			"expected normalized route %q in logs, got:\n%s",
			route,
			logOutput,
		)
	}

	if strings.Contains(
		logOutput,
		"route="+requestPath,
	) {
		t.Fatalf(
			"expected concrete request path %q to be absent from route label, got:\n%s",
			requestPath,
			logOutput,
		)
	}
}
