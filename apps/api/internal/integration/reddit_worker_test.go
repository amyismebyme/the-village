package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
)

func TestRedditWorkerIntegration(t *testing.T) {
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
									"id":"worker-post-1",
									"name":"t3_worker-post-1",
									"title":"Worker integration post",
									"selftext":"Worker integration test",
									"url":"https://example.test/post/1",
									"permalink":"/r/toronto/comments/worker-post-1/",
									"subreddit":"toronto",
									"author":"integration-test",
									"created_utc":1234567890
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

	authenticator, err := reddit.NewAuthenticator(
		server.Server.Client(),
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
		server.Server.Client(),
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

	select {
	case <-time.After(500 * time.Millisecond):
		cancel()
		t.Fatal(
			"worker did not execute within expected time",
		)

	default:
	}

	// The scheduler performs its first run immediately.
	cancel()

	select {
	case <-done:

	case <-time.After(500 * time.Millisecond):
		t.Fatal(
			"worker did not stop after cancellation",
		)
	}
}
