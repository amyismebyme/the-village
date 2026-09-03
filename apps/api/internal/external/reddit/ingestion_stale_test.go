package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestIngestListingDoesNotServeExpiredCacheOnSourceFailure(
	t *testing.T,
) {
	var requests atomic.Int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.URL.Path != "/r/toronto/new" {
					http.NotFound(w, r)
					return
				}

				if requests.Add(1) == 1 {
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

					return
				}

				w.WriteHeader(
					http.StatusInternalServerError,
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
			"create client: %v",
			err,
		)
	}

	ingestion := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	if err := ingestion.SetCache(
		memoryCache,
		time.Millisecond,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	ctx := context.Background()

	// First call populates the cache.
	if _, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	); err != nil {
		t.Fatalf(
			"seed ingestion: %v",
			err,
		)
	}

	// Ensure the cached value is expired.
	time.Sleep(
		5 * time.Millisecond,
	)

	// The source now fails. The expired cached value must NOT be served.
	_, err = ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)
	if err == nil {
		t.Fatal(
			"expected source failure after cache expiry",
		)
	}

	if !external.IsUpstream(err) {
		t.Fatalf(
			"expected upstream failure, got %v",
			err,
		)
	}

	if requests.Load() != 2 {
		t.Fatalf(
			"expected source to be called twice, got %d",
			requests.Load(),
		)
	}
}
