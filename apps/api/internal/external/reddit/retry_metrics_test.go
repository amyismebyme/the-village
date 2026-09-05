package reddit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestFetchListingRetryMetricsCountPhysicalAttempts(t *testing.T) {
	metrics.ExternalRequestsTotal.Reset()
	metrics.ExternalRequestDuration.Reset()
	metrics.ExternalErrorsTotal.Reset()
	metrics.ExternalRetriesTotal.Reset()
	metrics.ExternalRetryDelay.Reset()
	metrics.ExternalRetryExhaustedTotal.Reset()

	var requests int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				call := atomic.AddInt32(&requests, 1)

				if call < 3 {
					http.Error(
						w,
						"temporary upstream failure",
						http.StatusServiceUnavailable,
					)
					return
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"data": {
							"after": null,
							"before": null,
							"children": []
						}
					}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		server.Client(),
		server.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf("create Reddit client: %v", err)
	}

	backoff, err := external.NewBackoff(
		time.Millisecond,
		5*time.Millisecond,
		2,
		0,
	)
	if err != nil {
		t.Fatalf("create backoff: %v", err)
	}

	policy, err := external.NewRetryPolicy(3, backoff)
	if err != nil {
		t.Fatalf("create retry policy: %v", err)
	}

	client.SetRetryPolicy(policy)

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf("fetch listing: %v", err)
	}

	if got := atomic.LoadInt32(&requests); got != 3 {
		t.Fatalf("expected 3 physical requests, got %d", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRequestsTotal.WithLabelValues(
			"reddit",
			"fetch",
			"5xx",
		),
	); got != 2 {
		t.Fatalf("expected two 503 request attempts, got %v", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRequestsTotal.WithLabelValues(
			"reddit",
			"fetch",
			"200",
		),
	); got != 1 {
		t.Fatalf("expected one successful request attempt, got %v", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalErrorsTotal.WithLabelValues(
			"reddit",
			"fetch",
			"upstream",
		),
	); got != 2 {
		t.Fatalf("expected two upstream request errors, got %v", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRetriesTotal.WithLabelValues(
			"reddit",
			"fetch",
			"upstream",
		),
	); got != 2 {
		t.Fatalf("expected two scheduled retries, got %v", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRetryExhaustedTotal.WithLabelValues(
			"reddit",
			"fetch",
		),
	); got != 0 {
		t.Fatalf("expected zero retry exhaustions, got %v", got)
	}
}

func TestFetchListingRetryExhaustionMetricIsOneLogicalEvent(
	t *testing.T,
) {
	metrics.ExternalRequestsTotal.Reset()
	metrics.ExternalRequestDuration.Reset()
	metrics.ExternalErrorsTotal.Reset()
	metrics.ExternalRetriesTotal.Reset()
	metrics.ExternalRetryDelay.Reset()
	metrics.ExternalRetryExhaustedTotal.Reset()

	var requests int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requests, 1)
				http.Error(
					w,
					"upstream failure",
					http.StatusServiceUnavailable,
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		server.Client(),
		server.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf("create Reddit client: %v", err)
	}

	backoff, err := external.NewBackoff(
		time.Millisecond,
		5*time.Millisecond,
		2,
		0,
	)
	if err != nil {
		t.Fatalf("create backoff: %v", err)
	}

	policy, err := external.NewRetryPolicy(3, backoff)
	if err != nil {
		t.Fatalf("create retry policy: %v", err)
	}

	client.SetRetryPolicy(policy)

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err == nil {
		t.Fatal("expected retry exhaustion error")
	}

	if !errors.Is(err, external.ErrRetryExhausted) {
		t.Fatalf("expected ErrRetryExhausted, got %v", err)
	}

	if got := atomic.LoadInt32(&requests); got != 3 {
		t.Fatalf("expected 3 physical requests, got %d", got)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRetriesTotal.WithLabelValues(
			"reddit",
			"fetch",
			"upstream",
		),
	); got != 2 {
		t.Fatalf(
			"expected two 5xx request attempts, got %v",
			got,
		)
	}

	if got := testutil.ToFloat64(
		metrics.ExternalRetryExhaustedTotal.WithLabelValues(
			"reddit",
			"fetch",
		),
	); got != 1 {
		t.Fatalf("expected one retry exhaustion, got %v", got)
	}
}
