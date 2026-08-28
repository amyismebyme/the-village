package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestIngestionWorkerRunsRedditIngestion(
	t *testing.T,
) {
	var requests atomic.Int32

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests.Add(1)

				switch r.URL.Path {
				case "/api/v1/access_token":
					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					_, _ = w.Write(
						[]byte(`{
							"access_token":"test-token",
							"token_type":"bearer",
							"expires_in":3600,
							"scope":"*"
						}`),
					)

				case "/r/toronto/new":
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
										"kind":"t3",
										"data": {
											"id":"abc123",
											"title":"Toronto Community",
											"selftext":"Example",
											"url":"https://example.com/post",
											"subreddit":"toronto",
											"created_utc":1234567890
										}
									}
								]
							}
						}`),
					)

				default:
					http.NotFound(
						w,
						r,
					)
				}
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
			"create Reddit client: %v",
			err,
		)
	}

	ingestion := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	redditWorker, err := NewIngestionWorker(
		authenticator,
		ingestion,
		WorkerConfig{
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

	done := make(chan error, 1)

	go func() {
		done <- redditWorker.Run(ctx)
	}()

	// The scheduler performs the first run immediately.
	deadline := time.After(
		500 * time.Millisecond,
	)

	for requests.Load() < 2 {
		select {
		case <-deadline:
			t.Fatalf(
				"expected authentication and listing requests, got %d",
				requests.Load(),
			)

		case <-time.After(5 * time.Millisecond):
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal(
			"Reddit worker did not stop after cancellation",
		)
	}
}
