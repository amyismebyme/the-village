//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCommunityAPICreateMetric(t *testing.T) {
	app := newIntegrationApp(t)

	metrics.CommunityCreateTotal.Reset()

	before := testutil.ToFloat64(
		metrics.CommunityCreateTotal.WithLabelValues("success"),
	)

	body := `{
		"name": "Metric Test Community",
		"slug": "metric-test-community",
		"description": "Community create metric integration test",
		"external_source": "integration"
	}`

	resp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	requireJSONResponse(
		t,
		resp,
		http.StatusCreated,
	)

	requireRequestID(
		t,
		resp,
	)

	_ = decodeCommunity(
		t,
		resp,
	)

	after := testutil.ToFloat64(
		metrics.CommunityCreateTotal.WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected community create metric to increase by 1, got %v",
			got,
		)
	}
}
