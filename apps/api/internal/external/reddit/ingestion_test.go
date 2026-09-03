package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIngestionServiceIngestListing(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.URL.Path !=
					"/r/toronto/new" {
					t.Fatalf(
						"unexpected path %q",
						r.URL.Path,
					)
				}

				if got := r.Header.Get(
					"Authorization",
				); got != "Bearer test-token" {
					t.Fatalf(
						"unexpected authorization %q",
						got,
					)
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
										"id": "abc123",
										"name": "t3_abc123",
										"title": "Toronto Community",
										"selftext": "Example body",
										"url": "https://example.com/post",
										"permalink": "/r/toronto/comments/abc123",
										"subreddit": "toronto",
										"author": "test-user",
										"created_utc": 1234567890,
										"is_self": true
									}
								},
								{
									"kind": "other",
									"data": {
										"id": "ignored",
										"title": "Ignored"
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
			"create client: %v",
			err,
		)
	}

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

	if len(items) != 1 {
		t.Fatalf(
			"expected 1 normalized item, got %d",
			len(items),
		)
	}

	item := items[0]

	if item.Source != external.SourceReddit {
		t.Fatalf(
			"unexpected source %q",
			item.Source,
		)
	}

	if item.ExternalID != "abc123" {
		t.Fatalf(
			"unexpected external ID %q",
			item.ExternalID,
		)
	}

	if item.Title != "Toronto Community" {
		t.Fatalf(
			"unexpected title %q",
			item.Title,
		)
	}
}

func TestIngestionServicePropagatesCancellation(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				<-r.Context().Done()
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

	service := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = service.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if err == nil {
		t.Fatal(
			"expected cancellation error",
		)
	}
}

type invalidItemNormalizer struct{}

func (invalidItemNormalizer) Normalize(
	ctx context.Context,
	input RedditPost,
) (external.Item, error) {
	return external.Item{
		Source:     external.SourceReddit,
		ExternalID: "",
		Title:      "Looks valid",
	}, nil
}

func TestIngestionServiceRejectsInvalidNormalizedItem(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
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
										"id": "abc123",
										"title": "Example"
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
			"create client: %v",
			err,
		)
	}

	service := NewIngestionService(
		client,
		invalidItemNormalizer{},
	)

	_, err = service.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)

	if err == nil {
		t.Fatal(
			"expected normalized item validation error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"validate normalized item",
	) {
		t.Fatalf(
			"expected validation context, got %v",
			err,
		)
	}
}

func TestNewIngestionWorkerRejectsInvalidConfiguration(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
			},
		),
	)
	defer server.Close()

	authenticator, err := NewAuthenticator(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

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

	tests := []struct {
		name   string
		config WorkerConfig
	}{
		{
			name: "invalid limit",
			config: WorkerConfig{
				Subreddit: "toronto",
				Limit:     0,
				Interval:  time.Minute,
			},
		},
		{
			name: "invalid interval",
			config: WorkerConfig{
				Subreddit: "toronto",
				Limit:     10,
				Interval:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIngestionWorker(
				authenticator,
				ingestion,
				tt.config,
			)

			if err == nil {
				t.Fatal(
					"expected configuration error",
				)
			}
		})
	}
}

func TestIngestListingRejectsMissingSubreddit(
	t *testing.T,
) {
	server := testutil.NewRouteServer(
		map[string]testutil.Response{},
		testutil.Response{
			StatusCode: http.StatusNotFound,
		},
	)
	defer server.Close()

	client, err := NewClient(
		server.Server.Client(),
		server.URL(),
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create Reddit client: %v",
			err,
		)
	}

	service := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	_, err = service.IngestListing(
		context.Background(),
		"test-token",
		"",
		10,
		"",
	)

	if err == nil {
		t.Fatal(
			"expected missing subreddit error",
		)
	}

	if !errors.Is(
		err,
		external.ErrInvalidConfig,
	) {
		t.Fatalf(
			"expected ErrInvalidConfig, got %v",
			err,
		)
	}
}

func TestIngestListingCacheExpires(t *testing.T) {
	var requests int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests++

				if r.URL.Path != "/r/toronto/new" {
					t.Fatalf(
						"unexpected path %q",
						r.URL.Path,
					)
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
									"id": "abc123",
									"name": "t3_abc123",
									"title": "Toronto Community",
									"selftext": "Example body",
									"url": "https://example.com/post",
									"permalink": "/r/toronto/comments/abc123",
									"subreddit": "toronto",
									"author": "test-user",
									"created_utc": 1234567890,
									"is_self": true
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
		20*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	ctx := context.Background()

	// First ingestion should fetch from Reddit.
	if _, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	); err != nil {
		t.Fatalf(
			"first ingestion: %v",
			err,
		)
	}

	if requests != 1 {
		t.Fatalf(
			"expected one Reddit request after first ingestion, got %d",
			requests,
		)
	}

	// Second ingestion happens while the cache is valid.
	// It should use the cached response.
	if _, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	); err != nil {
		t.Fatalf(
			"second ingestion: %v",
			err,
		)
	}

	if requests != 1 {
		t.Fatalf(
			"expected cached response to avoid Reddit request, got %d requests",
			requests,
		)
	}

	// Wait for the cache entry to expire.
	time.Sleep(
		30 * time.Millisecond,
	)

	// Third ingestion should fetch from Reddit again.
	if _, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	); err != nil {
		t.Fatalf(
			"third ingestion: %v",
			err,
		)
	}

	if requests != 2 {
		t.Fatalf(
			"expected two Reddit requests after cache expiry, got %d",
			requests,
		)
	}

}

func TestIngestListingIgnoresCorruptCachedValue(
	t *testing.T,
) {
	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create key: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		[]byte("{bad json"),
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set corrupt cache: %v",
			err,
		)
	}

	// Run ingestion. It should delete the bad entry,
	// fetch from Reddit, and replace it with good data.
}

func TestIngestListingCacheBypassDoesNotUpdateCache(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
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
										"id": "new-id",
										"name": "t3_new-id",
										"title": "Fresh Reddit Post",
										"selftext": "fresh",
										"url": "https://example.com/new",
										"permalink": "/r/toronto/comments/new-id",
										"subreddit": "toronto",
										"author": "test-user",
										"created_utc": 1234567890,
										"is_self": true
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
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create cache key: %v",
			err,
		)
	}

	oldListing := ListingResponse{
		Data: ListingData{
			Children: []ListingChild{
				{
					Kind: "t3",
					Data: RedditPost{
						ID:        "old-id",
						Name:      "t3_old-id",
						Title:     "Cached Post",
						Subreddit: "toronto",
					},
				},
			},
		},
	}

	oldValue, err := json.Marshal(
		oldListing,
	)
	if err != nil {
		t.Fatalf(
			"encode old listing: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		oldValue,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"seed cache: %v",
			err,
		)
	}

	items, err := ingestion.IngestListingWithOptions(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
		IngestionOptions{
			CacheMode: CacheModeBypass,
		},
	)
	if err != nil {
		t.Fatalf(
			"bypass ingestion: %v",
			err,
		)
	}

	if len(items) != 1 {
		t.Fatalf(
			"expected one item, got %d",
			len(items),
		)
	}

	if items[0].ExternalID != "new-id" {
		t.Fatalf(
			"expected fresh item, got %q",
			items[0].ExternalID,
		)
	}

	// Bypass must not replace the existing cache.
	cached, ok, err := memoryCache.Get(
		context.Background(),
		key,
	)
	if err != nil {
		t.Fatalf(
			"get cache: %v",
			err,
		)
	}

	if !ok {
		t.Fatal(
			"expected original cache entry to remain",
		)
	}

	var cachedListing ListingResponse

	if err := DecodeCachedJSON(
		cached,
		&cachedListing,
	); err != nil {
		t.Fatalf(
			"decode cached listing: %v",
			err,
		)
	}

	if cachedListing.Data.Children[0].Data.ID !=
		"old-id" {
		t.Fatalf(
			"expected old cache entry to remain, got %q",
			cachedListing.Data.Children[0].Data.ID,
		)
	}
}

func TestIngestListingCacheRefreshReplacesCache(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
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
										"id": "fresh-id",
										"name": "t3_fresh-id",
										"title": "Fresh Post",
										"selftext": "fresh",
										"url": "https://example.com/fresh",
										"permalink": "/r/toronto/comments/fresh-id",
										"subreddit": "toronto",
										"author": "test-user",
										"created_utc": 1234567890,
										"is_self": true
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
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create cache key: %v",
			err,
		)
	}

	oldListing := ListingResponse{
		Data: ListingData{
			Children: []ListingChild{
				{
					Kind: "t3",
					Data: RedditPost{
						ID:        "old-id",
						Name:      "t3_old-id",
						Title:     "Old Post",
						Subreddit: "toronto",
					},
				},
			},
		},
	}

	oldValue, err := json.Marshal(
		oldListing,
	)
	if err != nil {
		t.Fatalf(
			"encode old listing: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		oldValue,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"seed cache: %v",
			err,
		)
	}

	// Refresh must skip the old cache and fetch Reddit.
	items, err := ingestion.IngestListingWithOptions(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
		IngestionOptions{
			CacheMode: CacheModeRefresh,
		},
	)
	if err != nil {
		t.Fatalf(
			"refresh ingestion: %v",
			err,
		)
	}

	if len(items) != 1 {
		t.Fatalf(
			"expected one item, got %d",
			len(items),
		)
	}

	if items[0].ExternalID != "fresh-id" {
		t.Fatalf(
			"expected fresh item, got %q",
			items[0].ExternalID,
		)
	}

	// A normal request after refresh should get the new cached value.
	items, err = ingestion.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"cached ingestion after refresh: %v",
			err,
		)
	}

	if len(items) != 1 {
		t.Fatalf(
			"expected one cached item, got %d",
			len(items),
		)
	}

	if items[0].ExternalID != "fresh-id" {
		t.Fatalf(
			"expected refreshed cache value, got %q",
			items[0].ExternalID,
		)
	}
}

func TestIngestListingCanceledBeforeCacheAccess(
	t *testing.T,
) {
	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	client, err := NewClient(
		http.DefaultClient,
		"http://example.test",
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

	if err := ingestion.SetCache(
		memoryCache,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestIngestListingCancellationPreventsSourceRequest(
	t *testing.T,
) {
	var requests int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests++

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
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	ingestion := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if requests != 0 {
		t.Fatalf(
			"expected zero Reddit requests, got %d",
			requests,
		)
	}
}

func TestIngestListingCacheHitStillHonorsCancellation(
	t *testing.T,
) {
	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	client, err := NewClient(
		http.DefaultClient,
		"http://example.test",
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

	if err := ingestion.SetCache(
		memoryCache,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create key: %v",
			err,
		)
	}

	listing := ListingResponse{
		Data: ListingData{
			Children: []ListingChild{},
		},
	}

	value, err := json.Marshal(listing)
	if err != nil {
		t.Fatalf(
			"marshal listing: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		value,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"seed cache: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestIngestListingDeduplicatesExternalIdentities(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
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
										"id": "same-id",
										"title": "First occurrence",
										"subreddit": "toronto"
									}
								},
								{
									"kind": "t3",
									"data": {
										"id": "same-id",
										"title": "Duplicate occurrence",
										"subreddit": "toronto"
									}
								},
								{
									"kind": "t3",
									"data": {
										"id": "unique-id",
										"title": "Unique occurrence",
										"subreddit": "toronto"
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

	if len(items) != 2 {
		t.Fatalf(
			"expected 2 unique items, got %d",
			len(items),
		)
	}

	if items[0].ExternalID != "same-id" {
		t.Fatalf(
			"expected first identity to be same-id, got %q",
			items[0].ExternalID,
		)
	}

	if items[0].Title != "First occurrence" {
		t.Fatalf(
			"expected first occurrence to win, got %q",
			items[0].Title,
		)
	}

	if items[1].ExternalID != "unique-id" {
		t.Fatalf(
			"expected second identity to be unique-id, got %q",
			items[1].ExternalID,
		)
	}
}
