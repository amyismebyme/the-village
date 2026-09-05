package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/ratelimit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
)

func TestRedditWorkerFailureThenRecovery(
	t *testing.T,
) {
	var (
		mu          sync.Mutex
		listingRuns int
	)

	server := testutil.NewRouteServer(
		map[string]testutil.Response{
			"/api/v1/access_token": {
				StatusCode: http.StatusOK,
				Body: `{
					"access_token":"integration-token",
					"token_type":"bearer",
					"expires_in":3600,
					"scope":"*"
				}`,
			},

			"/r/toronto/new": {
				StatusCode: http.StatusOK,
				Body: `{
					"data":{
						"after":null,
						"before":null,
						"children":[]
					}
				}`,
			},
		},
		testutil.Response{
			StatusCode: http.StatusNotFound,
		},
	)

	defer server.Close()

	// The route server is extended by wrapping the HTTP client so
	// the first listing request returns an upstream failure.
	clientHTTP := &failureThenSuccessHTTPClient{
		Client:       server.Server.Client(),
		FailurePath:  "/r/toronto/new",
		FailureCount: &listingRuns,
		Mu:           &mu,
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
		1,
		backoff,
	)
	if err != nil {
		t.Fatalf(
			"create retry policy: %v",
			err,
		)
	}

	limiters := ratelimit.NewPerSource()

	redditLimiter, err := limiters.Register(
		external.SourceReddit,
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"register Reddit limiter: %v",
			err,
		)
	}

	authenticator, err := reddit.NewAuthenticator(
		clientHTTP,
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
		clientHTTP,
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

	authenticator.SetRateLimiter(
		redditLimiter,
	)

	authenticator.SetRetryPolicy(
		retryPolicy,
	)

	client.SetRateLimiter(
		redditLimiter,
	)

	client.SetRetryPolicy(
		retryPolicy,
	)

	ingestion := reddit.NewIngestionService(
		client,
		reddit.NewPostNormalizer(),
	)

	worker, err := reddit.NewIngestionWorker(
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

	// First run: listing request fails.
	err = worker.RunOnce(ctx)

	if err == nil {
		t.Fatal(
			"expected first worker run to fail",
		)
	}

	if !external.IsUpstream(err) {
		t.Fatalf(
			"expected upstream error, got %v",
			err,
		)
	}

	// Second run: listing request succeeds.
	err = worker.RunOnce(ctx)

	if err != nil {
		t.Fatalf(
			"expected worker to recover on second run: %v",
			err,
		)
	}

	mu.Lock()
	runs := listingRuns
	mu.Unlock()

	if runs != 2 {
		t.Fatalf(
			"expected 2 listing attempts across worker runs, got %d",
			runs,
		)
	}
}

type failureThenSuccessHTTPClient struct {
	Client *http.Client

	FailurePath string

	FailureCount *int

	Mu *sync.Mutex
}

func (c *failureThenSuccessHTTPClient) Do(
	req *http.Request,
) (*http.Response, error) {
	if req.URL.Path == c.FailurePath {
		c.Mu.Lock()

		*c.FailureCount++
		count := *c.FailureCount

		c.Mu.Unlock()

		if count == 1 {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       http.NoBody,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}
	}

	return c.Client.Do(req)
}
