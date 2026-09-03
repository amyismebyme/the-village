package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestIngestListingRetryThenDeduplicatesFinalResponse(
	t *testing.T,
) {
	var calls int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				call := atomic.AddInt32(
					&calls,
					1,
				)

				if call == 1 {
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
							"children": [
								{
									"kind": "t3",
									"data": {
										"id": "retry-duplicate",
										"title": "First final item"
									}
								},
								{
									"kind": "t3",
									"data": {
										"id": "retry-duplicate",
										"title": "Duplicate final item"
									}
								},
								{
									"kind": "t3",
									"data": {
										"id": "retry-unique",
										"title": "Unique final item"
									}
								}
							]
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
		t.Fatalf(
			"create Reddit client: %v",
			err,
		)
	}

	backoff, err := external.NewBackoff(
		time.Millisecond,
		5*time.Millisecond,
		2,
		0,
	)
	if err != nil {
		t.Fatalf(
			"create backoff: %v",
			err,
		)
	}

	retryPolicy, err := external.NewRetryPolicy(
		2,
		backoff,
	)
	if err != nil {
		t.Fatalf(
			"create retry policy: %v",
			err,
		)
	}

	client.SetRetryPolicy(
		retryPolicy,
	)

	service := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	items, err := service.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"ingest listing: %v",
			err,
		)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf(
			"expected exactly two source attempts, got %d",
			got,
		)
	}

	if len(items) != 2 {
		t.Fatalf(
			"expected two unique items after retry, got %d",
			len(items),
		)
	}

	if items[0].ExternalID != "retry-duplicate" {
		t.Fatalf(
			"expected retry-duplicate first, got %q",
			items[0].ExternalID,
		)
	}

	if items[0].Title != "First final item" {
		t.Fatalf(
			"expected first occurrence to win, got %q",
			items[0].Title,
		)
	}

	if items[1].ExternalID != "retry-unique" {
		t.Fatalf(
			"expected retry-unique second, got %q",
			items[1].ExternalID,
		)
	}
}