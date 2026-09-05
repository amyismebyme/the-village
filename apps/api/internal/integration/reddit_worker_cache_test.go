package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
)

func TestRedditWorkerUsesCacheAcrossRuns(
	t *testing.T,
) {
	var (
		mu              sync.Mutex
		tokenRequests   int
		listingRequests int
	)

	server := testutil.NewRouteServer(
		map[string]testutil.Response{
			"/api/v1/access_token": {
				StatusCode: http.StatusOK,
				Body: `{
					"access_token":"worker-cache-token",
					"token_type":"bearer",
					"expires_in":3600,
					"scope":"*"
				}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			"/r/toronto/new": {
				StatusCode: http.StatusOK,
				Body: `{
					"data":{
						"after":null,
						"before":null,
						"children":[
							{
								"kind":"t3",
								"data":{
									"id":"cache-worker-post",
									"name":"t3_cache-worker-post",
									"title":"Cached worker post",
									"selftext":"Cached worker test",
									"url":"https://example.test/post",
									"permalink":"/r/toronto/comments/cache-worker-post/",
									"subreddit":"toronto",
									"author":"integration-test",
									"created_utc":1234567890,
									"is_self":true
								}
							}
						]
					}
				}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		testutil.Response{
			StatusCode: http.StatusNotFound,
		},
	)
	defer server.Close()

	recordingClient := &countingHTTPClient{
		Client: server.Server.Client(),
		Record: func(path string) {
			mu.Lock()
			defer mu.Unlock()

			switch path {
			case "/api/v1/access_token":
				tokenRequests++

			case "/r/toronto/new":
				listingRequests++
			}
		},
	}

	authenticator, err := reddit.NewAuthenticator(
		recordingClient,
		server.URL(),
		"client-id",
		"client-secret",
		"the-village/integration",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

	client, err := reddit.NewClient(
		recordingClient,
		server.URL(),
		"the-village/integration",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create Reddit client: %v",
			err,
		)
	}

	ingestion := reddit.NewIngestionService(
		client,
		reddit.NewPostNormalizer(),
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
			"configure cache: %v",
			err,
		)
	}

	redditWorker, err := reddit.NewIngestionWorker(
		authenticator,
		ingestion,
		&integrationExternalItemRepository{},
		reddit.WorkerConfig{
			Subreddit: "toronto",
			Limit:     10,
			Interval:  time.Hour,
		},
	)
	if err != nil {
		t.Fatalf(
			"create Reddit worker: %v",
			err,
		)
	}

	ctx := context.Background()

	// First run:
	// OAuth request + Reddit listing request.
	if err := redditWorker.RunOnce(ctx); err != nil {
		t.Fatalf(
			"first worker run: %v",
			err,
		)
	}

	// Second run:
	// OAuth token is already cached by the authenticator.
	// Reddit listing should come from the ingestion cache.
	if err := redditWorker.RunOnce(ctx); err != nil {
		t.Fatalf(
			"second worker run: %v",
			err,
		)
	}

	mu.Lock()
	gotTokenRequests := tokenRequests
	gotListingRequests := listingRequests
	mu.Unlock()

	if gotTokenRequests != 1 {
		t.Fatalf(
			"expected one OAuth request, got %d",
			gotTokenRequests,
		)
	}

	if gotListingRequests != 1 {
		t.Fatalf(
			"expected one Reddit listing request across two worker runs, got %d",
			gotListingRequests,
		)
	}
}

type countingHTTPClient struct {
	Client *http.Client
	Record func(path string)
}

func (c *countingHTTPClient) Do(
	req *http.Request,
) (*http.Response, error) {
	if c.Record != nil {
		c.Record(req.URL.Path)
	}

	return c.Client.Do(req)
}

// Ensure the wrapper continues to satisfy the external contract.
var _ external.HTTPClient = (*countingHTTPClient)(nil)
