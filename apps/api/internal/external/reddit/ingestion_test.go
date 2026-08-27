package reddit

import (
	"context"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
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
