package reddit

import (
	"encoding/json"
	"testing"
)

func TestListingResponseDecoding(t *testing.T) {
	payload := []byte(`{
		"data": {
			"after": "t3_next",
			"before": null,
			"children": [
				{
					"kind": "t3",
					"data": {
						"id": "abc123",
						"name": "t3_abc123",
						"title": "Toronto Community",
						"selftext": "Example post body",
						"url": "https://example.com/post",
						"permalink": "/r/toronto/comments/abc123/example",
						"subreddit": "toronto",
						"author": "example-user",
						"created_utc": 1234567890,
						"is_self": true,
						"removed_by": null
					}
				}
			]
		}
	}`)

	var response ListingResponse

	if err := json.Unmarshal(
		payload,
		&response,
	); err != nil {
		t.Fatalf(
			"decode listing: %v",
			err,
		)
	}

	if response.Data.After == nil {
		t.Fatal(
			"expected after cursor",
		)
	}

	if *response.Data.After != "t3_next" {
		t.Fatalf(
			"unexpected after cursor %q",
			*response.Data.After,
		)
	}

	if len(response.Data.Children) != 1 {
		t.Fatalf(
			"expected 1 child, got %d",
			len(response.Data.Children),
		)
	}

	post := response.Data.Children[0].Data

	if post.ID != "abc123" {
		t.Fatalf(
			"unexpected ID %q",
			post.ID,
		)
	}

	if post.Name != "t3_abc123" {
		t.Fatalf(
			"unexpected fullname %q",
			post.Name,
		)
	}

	if post.Title != "Toronto Community" {
		t.Fatalf(
			"unexpected title %q",
			post.Title,
		)
	}

	if post.Subreddit != "toronto" {
		t.Fatalf(
			"unexpected subreddit %q",
			post.Subreddit,
		)
	}
}

func TestListingResponseHandlesEmptyChildren(
	t *testing.T,
) {
	payload := []byte(`{
		"data": {
			"after": null,
			"before": null,
			"children": []
		}
	}`)

	var response ListingResponse

	if err := json.Unmarshal(
		payload,
		&response,
	); err != nil {
		t.Fatalf(
			"decode empty listing: %v",
			err,
		)
	}

	if response.Data.Children == nil {
		t.Fatal(
			"expected non-nil children slice",
		)
	}

	if len(response.Data.Children) != 0 {
		t.Fatalf(
			"expected zero children, got %d",
			len(response.Data.Children),
		)
	}
}
