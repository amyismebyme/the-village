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
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCommunityCreateMetricLogCorrelation(
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

	metrics.CommunityCreateTotal.Reset()

	beforeCreate := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	body := `{
		"name": "Community Correlation Test",
		"slug": "community-correlation-test",
		"description": "Community observability correlation test",
		"external_source": "integration"
	}`

	resp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	requestID := requireRequestID(
		t,
		resp,
	)

	requireJSONResponse(
		t,
		resp,
		http.StatusCreated,
	)

	created := decodeCommunity(
		t,
		resp,
	)

	if created.ID <= 0 {
		t.Fatalf(
			"expected created community ID > 0, got %d",
			created.ID,
		)
	}

	afterCreate := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	if got := afterCreate - beforeCreate; got != 1 {
		t.Fatalf(
			"expected create metric to increase by 1, got %v",
			got,
		)
	}

	logOutput := logs.String()

	for _, expected := range []string{
		`msg="community operation completed"`,
		"request_id=" + requestID,
		"operation=create",
		"community_id=",
		"status=201",
		"duration_ms=",
	} {
		if !strings.Contains(
			logOutput,
			expected,
		) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				logOutput,
			)
		}
	}

	if strings.Contains(
		logOutput,
		"password",
	) || strings.Contains(
		logOutput,
		"authorization",
	) || strings.Contains(
		logOutput,
		"token",
	) {
		t.Fatalf(
			"community operation log contains sensitive request data:\n%s",
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
		"village_community_create_total",
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

	if strings.Contains(
		metricsText,
		`request_id="`,
	) {
		t.Fatalf(
			"request_id leaked into Prometheus labels:\n%s",
			metricsText,
		)
	}

	if strings.Contains(
		metricsText,
		`community_id="`,
	) {
		t.Fatalf(
			"community_id leaked into Prometheus labels:\n%s",
			metricsText,
		)
	}
}
