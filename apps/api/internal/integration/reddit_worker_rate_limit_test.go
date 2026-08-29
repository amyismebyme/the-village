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

func TestRedditWorkerUsesSharedRateLimiter(
	t *testing.T,
) {
	var (
		mu           sync.Mutex
		requestTimes []time.Time
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

	// Use a transport wrapper to record actual outbound request starts.
	recordingClient := &recordingHTTPClient{
		Client: server.Server.Client(),
		Record: func() {
			mu.Lock()
			defer mu.Unlock()

			requestTimes = append(
				requestTimes,
				time.Now(),
			)
		},
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
		30*time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"register Reddit limiter: %v",
			err,
		)
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

	err = worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"worker run: %v",
			err,
		)
	}

	mu.Lock()
	times := append(
		[]time.Time(nil),
		requestTimes...,
	)
	mu.Unlock()

	if len(times) != 2 {
		t.Fatalf(
			"expected 2 outbound requests, got %d",
			len(times),
		)
	}

	elapsed := times[1].Sub(
		times[0],
	)

	if elapsed < 25*time.Millisecond {
		t.Fatalf(
			"expected shared Reddit pacing of at least 30ms, got %s",
			elapsed,
		)
	}
}

type recordingHTTPClient struct {
	Client *http.Client
	Record func()
}

func (c *recordingHTTPClient) Do(
	req *http.Request,
) (*http.Response, error) {
	if c.Record != nil {
		c.Record()
	}

	return c.Client.Do(req)
}
