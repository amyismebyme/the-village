package reddit

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
)

func TestRedditIngestionEndToEnd(t *testing.T) {
	server := testutil.NewRouteServer(
		map[string]testutil.Response{
			"/api/v1/access_token": {
				StatusCode: http.StatusOK,
				Body: `{
					"access_token": "test-access-token",
					"token_type": "bearer",
					"expires_in": 3600,
					"scope": "*"
				}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			"/r/toronto/new": {
				StatusCode: http.StatusOK,
				Body: `{
					"data": {
						"after": null,
						"before": null,
						"children": [
							{
								"kind": "t3",
								"data": {
									"id": "abc123",
									"name": "t3_abc123",
									"title": "  Toronto Men Community  ",
									"selftext": "  Example Reddit post.  ",
									"url": "https://www.reddit.com/r/toronto/comments/abc123/example/",
									"permalink": "/r/toronto/comments/abc123/example/",
									"subreddit": "toronto",
									"author": "test-user",
									"created_utc": 1234567890,
									"is_self": true,
									"removed_by": null
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

	authenticator, err := NewAuthenticator(
		server.Server.Client(),
		server.URL(),
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

	token, err := authenticator.Token(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"get Reddit token: %v",
			err,
		)
	}

	if token != "test-access-token" {
		t.Fatalf(
			"unexpected token %q",
			token,
		)
	}

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

	ingestion := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	items, err := ingestion.IngestListing(
		context.Background(),
		token,
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"ingest Reddit listing: %v",
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
			"expected source %q, got %q",
			external.SourceReddit,
			item.Source,
		)
	}

	if item.ExternalID != "abc123" {
		t.Fatalf(
			"expected external ID %q, got %q",
			"abc123",
			item.ExternalID,
		)
	}

	if item.Title != "Toronto Men Community" {
		t.Fatalf(
			"unexpected title %q",
			item.Title,
		)
	}

	if item.Description != "Example Reddit post." {
		t.Fatalf(
			"unexpected description %q",
			item.Description,
		)
	}

	expectedURL :=
		"https://www.reddit.com/r/toronto/comments/abc123/example/"

	if item.URL != expectedURL {
		t.Fatalf(
			"unexpected URL %q",
			item.URL,
		)
	}

	identity := item.Identity()

	if identity.Source != external.SourceReddit {
		t.Fatalf(
			"unexpected identity source %q",
			identity.Source,
		)
	}

	if identity.ExternalID != "abc123" {
		t.Fatalf(
			"unexpected identity external ID %q",
			identity.ExternalID,
		)
	}

	if identity.Key() != "reddit:abc123" {
		t.Fatalf(
			"unexpected identity key %q",
			identity.Key(),
		)
	}
}

func TestRedditIngestionEndToEndRejectsInvalidPayload(
	t *testing.T,
) {
	server := testutil.NewRouteServer(
		map[string]testutil.Response{
			"/r/toronto/new": {
				StatusCode: http.StatusOK,
				Body: `{
					"data": {
						"after": null,
						"before": null,
						"children": [
							{
								"kind": "t3",
								"data": {
									"id": "",
									"title": "Invalid Reddit post"
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

	ingestion := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	_, err = ingestion.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err == nil {
		t.Fatal(
			"expected invalid payload error",
		)
	}

	if !errors.Is(
		err,
		external.ErrInvalidPayload,
	) {
		t.Fatalf(
			"expected ErrInvalidPayload, got %v",
			err,
		)
	}
}
