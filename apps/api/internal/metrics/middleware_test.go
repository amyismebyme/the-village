package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPMetricsMiddleware(t *testing.T) {
	beforeTotal := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			http.MethodGet,
			"/test",
			"200",
		),
	)

	beforeInFlight := testutil.ToFloat64(RequestsInFlight)

	handler := Middleware(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	afterTotal := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			http.MethodGet,
			"/test",
			"200",
		),
	)

	if got := afterTotal - beforeTotal; got != 1 {
		t.Fatalf(
			"expected requests total to increase by 1, got %v",
			got,
		)
	}

	if got := testutil.ToFloat64(RequestsInFlight); got != beforeInFlight {
		t.Fatalf(
			"expected in-flight requests to return to %v, got %v",
			beforeInFlight,
			got,
		)
	}
}

func TestHTTPMetricsStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantStatus string
	}{
		{
			name:       "ok",
			status:     http.StatusOK,
			wantStatus: "200",
		},
		{
			name:       "not found",
			status:     http.StatusNotFound,
			wantStatus: "404",
		},
		{
			name:       "internal server error",
			status:     http.StatusInternalServerError,
			wantStatus: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := testutil.ToFloat64(
				RequestsTotal.WithLabelValues(
					http.MethodGet,
					"/status-test",
					tt.wantStatus,
				),
			)

			handler := Middleware(
				http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					w.WriteHeader(tt.status)
				}),
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/status-test",
				nil,
			)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			after := testutil.ToFloat64(
				RequestsTotal.WithLabelValues(
					http.MethodGet,
					"/status-test",
					tt.wantStatus,
				),
			)

			if got := after - before; got != 1 {
				t.Fatalf(
					"expected status %d counter to increase by 1, got %v",
					tt.status,
					got,
				)
			}
		})
	}
}

func TestHTTPMetricsInFlight(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	before := testutil.ToFloat64(RequestsInFlight)

	handler := Middleware(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			close(started)

			<-release

			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/in-flight",
		nil,
	)

	rec := httptest.NewRecorder()

	done := make(chan struct{})

	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	<-started

	during := testutil.ToFloat64(RequestsInFlight)

	if during != before+1 {
		t.Fatalf(
			"expected in-flight metric to increase by 1 while request executes; before=%v during=%v",
			before,
			during,
		)
	}

	close(release)

	<-done

	after := testutil.ToFloat64(RequestsInFlight)

	if after != before {
		t.Fatalf(
			"expected in-flight metric to return to baseline; before=%v after=%v",
			before,
			after,
		)
	}
}

func TestHTTPMetricsDuration(t *testing.T) {
	handler := Middleware(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/duration-test",
		nil,
	)

	rec := httptest.NewRecorder()

	before := histogramSampleCount(
		t,
		http.MethodGet,
		"/duration-test",
	)

	handler.ServeHTTP(rec, req)

	after := histogramSampleCount(
		t,
		http.MethodGet,
		"/duration-test",
	)

	if after != before+1 {
		t.Fatalf(
			"expected duration sample count to increase by 1, before=%d after=%d",
			before,
			after,
		)
	}
}

func histogramSampleCount(
	t *testing.T,
	method string,
	route string,
) uint64 {
	t.Helper()

	registry := prometheus.NewRegistry()

	if err := registry.Register(RequestDuration); err != nil {
		t.Fatalf("register request duration collector: %v", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather request duration metrics: %v", err)
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

func TestHTTPMetricsNormalizesServeMuxRoutePattern(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		func(
			w http.ResponseWriter,
			_ *http.Request,
		) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	before := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			http.MethodGet,
			"/api/v1/communities/{id}",
			"404",
		),
	)

	handler := Middleware(mux)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}

	after := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			http.MethodGet,
			"/api/v1/communities/{id}",
			"404",
		),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected normalized route counter to increase by 1, got %v",
			got,
		)
	}
}
