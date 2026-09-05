//go:build integration

package integration

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/ratelimit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
)

func TestRedditWorkerRecoversAfterDependencyRestoration(
	t *testing.T,
) {
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

	clientHTTP := newScriptedHTTPClient(
		server.Server.Client(),
	)

	clientHTTP.SetScript(
		"/r/toronto/new",
		scriptedHTTPOutcome{
			StatusCode: http.StatusServiceUnavailable,
			Body:       "dependency unavailable",
		},
		scriptedHTTPOutcome{
			StatusCode: http.StatusOK,
			Body: `{
				"data":{
					"after":null,
					"before":null,
					"children":[]
				}
			}`,
		},
	)

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

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx)
	}()

	deadline := time.NewTimer(
		2 * time.Second,
	)
	defer deadline.Stop()

	for {
		if clientHTTP.Calls(
			"/r/toronto/new",
		) >= 2 {
			cancel()
			break
		}

		select {
		case <-deadline.C:
			t.Fatalf(
				"worker did not reach the restored dependency",
			)

		case <-time.After(
			5 * time.Millisecond,
		):
		}
	}

	if err := <-done; !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled after shutdown, got %v",
			err,
		)
	}

	if calls := clientHTTP.Calls(
		"/r/toronto/new",
	); calls < 2 {
		t.Fatalf(
			"expected at least two listing calls, got %d",
			calls,
		)
	}
}

type scriptedHTTPOutcome struct {
	StatusCode int
	Body       string
}

type scriptedHTTPClient struct {
	client *http.Client

	mu      sync.Mutex
	scripts map[string][]scriptedHTTPOutcome
	calls   map[string]int
}

func newScriptedHTTPClient(
	client *http.Client,
) *scriptedHTTPClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &scriptedHTTPClient{
		client:  client,
		scripts: make(map[string][]scriptedHTTPOutcome),
		calls:   make(map[string]int),
	}
}

func (c *scriptedHTTPClient) SetScript(
	path string,
	outcomes ...scriptedHTTPOutcome,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.scripts[path] = append(
		[]scriptedHTTPOutcome(nil),
		outcomes...,
	)
}

func (c *scriptedHTTPClient) Calls(
	path string,
) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.calls[path]
}

func (c *scriptedHTTPClient) Do(
	req *http.Request,
) (*http.Response, error) {
	if req == nil {
		return nil, errors.New(
			"scripted HTTP client: request is required",
		)
	}

	path := req.URL.Path

	c.mu.Lock()

	c.calls[path]++

	outcomes := c.scripts[path]

	if len(outcomes) > 0 {
		outcome := outcomes[0]

		c.scripts[path] = outcomes[1:]

		c.mu.Unlock()

		return newScriptedResponse(
			req,
			outcome,
		), nil
	}

	c.mu.Unlock()

	return c.client.Do(req)
}

func newScriptedResponse(
	req *http.Request,
	outcome scriptedHTTPOutcome,
) *http.Response {
	return &http.Response{
		StatusCode: outcome.StatusCode,
		Status:     http.StatusText(outcome.StatusCode),
		Body: io.NopCloser(
			strings.NewReader(outcome.Body),
		),
		Header: http.Header{
			"Content-Type": []string{
				"application/json",
			},
		},
		Request: req,
	}
}

var _ external.HTTPClient = (*scriptedHTTPClient)(nil)
