package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHTTPMetricsDoNotUseRequestIDLabel(t *testing.T) {
	const requestID = "metrics-request-id-test"

	handler := middleware.RequestID(
		metrics.Middleware(
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	req.Header.Set(
		"X-Request-ID",
		requestID,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	registry := prometheus.NewRegistry()

	if err := registry.Register(metrics.RequestsTotal); err != nil {
		t.Fatalf(
			"register RequestsTotal: %v",
			err,
		)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "request_id" {
					t.Fatalf(
						"request_id must never be a Prometheus label",
					)
				}

				if strings.Contains(
					label.GetValue(),
					requestID,
				) {
					t.Fatalf(
						"request ID leaked into metric label value",
					)
				}
			}
		}
	}
}
