//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCommunityAPIDeleteMetric(t *testing.T) {
	app := newIntegrationApp(t)

	createBody := `{
		"name": "Delete Metric Community",
		"slug": "delete-metric-community",
		"description": "Delete metric integration test",
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

	created := decodeCommunity(t, createResp)

	metrics.CommunityDeleteTotal.Reset()

	before := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.WithLabelValues("success"),
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

	after := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected delete metric to increase by 1, got %v",
			got,
		)
	}
}
