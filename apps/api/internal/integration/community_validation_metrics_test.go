//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCommunityAPIValidationFailureMetric(t *testing.T) {
	app := newIntegrationApp(t)

	metrics.CommunityValidationFailuresTotal.Reset()

	before := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	body := `{
		"name": "",
		"slug": "valid-community",
		"description": "Invalid community"
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
		http.StatusBadRequest,
	)

	requireRequestID(
		t,
		resp,
	)

	resp.Body.Close()

	after := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected name validation metric to increase by 1, got %v",
			got,
		)
	}
}
