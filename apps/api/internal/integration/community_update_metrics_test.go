//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCommunityAPIUpdateMetric(t *testing.T) {
	app := newIntegrationApp(t)

	metrics.CommunityUpdateTotal.Reset()

	// Create the record that will be updated.
	createBody := `{
		"name": "Update Metric Community",
		"slug": "update-metric-community",
		"description": "Before update",
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

	created := decodeCommunity(
		t,
		createResp,
	)

	before := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.WithLabelValues("success"),
	)

	updateBody := `{
		"name": "Update Metric Community Updated",
		"slug": "update-metric-community-updated",
		"description": "After update",
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

	requireRequestID(
		t,
		updateResp,
	)

	updated := decodeCommunity(
		t,
		updateResp,
	)

	if updated.ID != created.ID {
		t.Fatalf(
			"expected updated ID %d, got %d",
			created.ID,
			updated.ID,
		)
	}

	if updated.Name != "Update Metric Community Updated" {
		t.Fatalf(
			"expected updated name, got %q",
			updated.Name,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected update metric to increase by 1, got %v",
			got,
		)
	}
}
