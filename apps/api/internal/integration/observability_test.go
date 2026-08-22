//go:build integration

package integration

import (
	"bytes"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"testing"
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

func TestCommunityAPIObservabilityCorrelation(t *testing.T) {
	var logs bytes.Buffer

	app := newIntegrationAppWithLogger(
		t,
		newTestLogger(&logs),
	)

	route := "/api/v1/communities"

	beforeRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodPost,
			route,
			"201",
		),
	)

	beforeDuration := histogramSampleCountForRoute(
		t,
		http.MethodPost,
		route,
	)

	body := `{
		"name": "Observability Community",
		"slug": "observability-community-integration",
		"description": "Observability integration test",
		"external_source": "integration"
	}`

	resp := integrationRequest(
		t,
		app,
		http.MethodPost,
		route,
		body,
	)

	requestID := requireRequestID(t, resp)

	requireJSONResponse(
		t,
		resp,
		http.StatusCreated,
	)

	created := decodeCommunity(t, resp)

	if created.ID <= 0 {
		t.Fatalf(
			"expected created community ID > 0, got %d",
			created.ID,
		)
	}

	afterRequests := testutil.ToFloat64(
		metrics.RequestsTotal.WithLabelValues(
			http.MethodPost,
			route,
			"201",
		),
	)

	if got := afterRequests - beforeRequests; got != 1 {
		t.Fatalf(
			"expected POST request counter to increase by 1, got %v",
			got,
		)
	}

	afterDuration := histogramSampleCountForRoute(
		t,
		http.MethodPost,
		route,
	)

	if afterDuration != beforeDuration+1 {
		t.Fatalf(
			"expected POST duration sample to increase by 1; before=%d after=%d",
			beforeDuration,
			afterDuration,
		)
	}

	logOutput := logs.String()

	for _, expected := range []string{
		"http request completed",
		"request_id=" + requestID,
		"method=POST",
		"route=" + route,
		"status=201",
		"duration_ms=",
	} {
		if !strings.Contains(
			logOutput,
			expected,
		) {
			t.Fatalf(
				"expected log to contain %q, got:\n%s",
				expected,
				logOutput,
			)
		}
	}

	if strings.Contains(
		logOutput,
		"route="+
			"/api/v1/communities/"+strconv.FormatInt(created.ID, 10),
	) {
		t.Fatalf(
			"expected normalized route instead of concrete community ID, got:\n%s",
			logOutput,
		)
	}

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

	metricsBody, err := io.ReadAll(
		metricsResponse.Body,
	)
	if err != nil {
		t.Fatalf(
			"read /metrics response: %v",
			err,
		)
	}

	metricsText := string(metricsBody)

	for _, metricName := range []string{
		"village_http_requests_total",
		"village_http_request_duration_seconds",
		"village_http_requests_in_flight",
	} {
		if !strings.Contains(
			metricsText,
			metricName,
		) {
			t.Fatalf(
				"expected /metrics to expose %s",
				metricName,
			)
		}
	}
}

func TestCommunityObservabilityLifecycle(
	t *testing.T,
) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	app := newIntegrationAppWithLogger(
		t,
		logger,
	)

	// ---------------------------------------------------------------------
	// CREATE
	// ---------------------------------------------------------------------

	createBefore := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	createBody := `{
		"name": "Observability Lifecycle",
		"slug": "observability-lifecycle",
		"description": "Observability integration test",
		"external_source": "integration"
	}`

	createResp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		createBody,
	)

	requireJSONResponse(
		t,
		createResp,
		http.StatusCreated,
	)

	createRequestID := requireRequestID(
		t,
		createResp,
	)

	created := decodeCommunity(
		t,
		createResp,
	)

	createAfter := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	if got := createAfter - createBefore; got != 1 {
		t.Fatalf(
			"expected create metric +1, got %v",
			got,
		)
	}

	// ---------------------------------------------------------------------
	// UPDATE
	// ---------------------------------------------------------------------

	updateBefore := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	updateBody := `{
		"name": "Observability Lifecycle Updated",
		"slug": "observability-lifecycle-updated",
		"description": "Updated observability test",
		"external_source": "integration-updated"
	}`

	updateResp := integrationRequest(
		t,
		app,
		http.MethodPut,
		communityPath(created.ID),
		updateBody,
	)

	requireJSONResponse(
		t,
		updateResp,
		http.StatusOK,
	)

	updateRequestID := requireRequestID(
		t,
		updateResp,
	)

	_ = decodeCommunity(
		t,
		updateResp,
	)

	updateAfter := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	if got := updateAfter - updateBefore; got != 1 {
		t.Fatalf(
			"expected update metric +1, got %v",
			got,
		)
	}

	// ---------------------------------------------------------------------
	// DELETE
	// ---------------------------------------------------------------------

	deleteBefore := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	deleteResp := integrationRequest(
		t,
		app,
		http.MethodDelete,
		communityPath(created.ID),
		"",
	)

	if deleteResp.StatusCode != http.StatusNoContent {
		defer deleteResp.Body.Close()

		t.Fatalf(
			"expected DELETE status %d, got %d",
			http.StatusNoContent,
			deleteResp.StatusCode,
		)
	}

	deleteResp.Body.Close()

	deleteAfter := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	if got := deleteAfter - deleteBefore; got != 1 {
		t.Fatalf(
			"expected delete metric +1, got %v",
			got,
		)
	}

	// ---------------------------------------------------------------------
	// VALIDATION FAILURE
	// ---------------------------------------------------------------------

	validationBefore := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	invalidBody := `{
		"name": "",
		"slug": "valid-validation-slug"
	}`

	validationResp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		invalidBody,
	)

	requireJSONResponse(
		t,
		validationResp,
		http.StatusBadRequest,
	)

	validationRequestID := requireRequestID(
		t,
		validationResp,
	)

	validationResp.Body.Close()

	validationAfter := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	if got := validationAfter - validationBefore; got != 1 {
		t.Fatalf(
			"expected name validation metric +1, got %v",
			got,
		)
	}

	// ---------------------------------------------------------------------
	// LOG CORRELATION
	// ---------------------------------------------------------------------

	logOutput := logs.String()

	expectedLogs := []string{
		`msg="community operation completed"`,
		"operation=create",
		"operation=update",
		"operation=delete",

		"community_id=",
		"request_id=" + createRequestID,
		"request_id=" + updateRequestID,
		"request_id=" + validationRequestID,

		"status=201",
		"status=200",
		"status=204",
		"status=400",

		"duration_ms=",
	}

	for _, expected := range expectedLogs {
		if !strings.Contains(
			logOutput,
			expected,
		) {
			t.Fatalf(
				"expected logs to contain %q; got:\n%s",
				expected,
				logOutput,
			)
		}
	}

	// Confirm that raw request data isn't being logged.
	for _, forbidden := range []string{
		"Authorization",
		"Bearer",
		"password",
		"token",
	} {
		if strings.Contains(
			logOutput,
			forbidden,
		) {
			t.Fatalf(
				"sensitive data %q appeared in logs:\n%s",
				forbidden,
				logOutput,
			)
		}
	}
	// Community operation logs must be distinguishable from the generic
	// request-completed log.
	if !strings.Contains(
		logOutput,
		`msg="community operation completed"`,
	) {
		t.Fatalf(
			"expected community operation completion log, got:\n%s",
			logOutput,
		)
	}

	// /metrics must expose the Community metric families that the
	// lifecycle above exercised.
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

	metricsBody, err := io.ReadAll(
		metricsResponse.Body,
	)
	if err != nil {
		t.Fatalf(
			"read /metrics response: %v",
			err,
		)
	}

	metricsText := string(metricsBody)

	for _, metricName := range []string{
		"village_community_create_total",
		"village_community_update_total",
		"village_community_delete_total",
		"village_community_validation_failures_total",
	} {
		if !strings.Contains(metricsText, metricName) {
			t.Fatalf(
				"expected /metrics to expose %s; got:\n%s",
				metricName,
				metricsText,
			)
		}
	}

	if strings.Contains(metricsText, `request_id="`) {
		t.Fatalf(
			"request_id leaked into Prometheus labels:\n%s",
			metricsText,
		)
	}

	if strings.Contains(metricsText, `community_id="`) {
		t.Fatalf(
			"community_id leaked into Prometheus labels:\n%s",
			metricsText,
		)
	}

}
