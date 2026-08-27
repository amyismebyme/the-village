package reddit

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestPostNormalizerNormalize(t *testing.T) {
	normalizer := NewPostNormalizer()

	item, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			ID:         " abc123 ",
			Name:       "t3_abc123",
			Title:      "  Toronto Community  ",
			SelfText:   "  Example post body  ",
			URL:        " https://www.reddit.com/r/toronto/comments/abc123/ ",
			Subreddit:  "toronto",
			Author:     "example-user",
			CreatedUTC: 1234567890,
			IsSelf:     true,
		},
	)
	if err != nil {
		t.Fatalf(
			"normalize: %v",
			err,
		)
	}

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

	if item.Title != "Toronto Community" {
		t.Fatalf(
			"unexpected title %q",
			item.Title,
		)
	}

	if item.Description != "Example post body" {
		t.Fatalf(
			"unexpected description %q",
			item.Description,
		)
	}

	if item.URL !=
		"https://www.reddit.com/r/toronto/comments/abc123/" {
		t.Fatalf(
			"unexpected URL %q",
			item.URL,
		)
	}
}

func TestPostNormalizerUsesPermalinkWhenURLMissing(
	t *testing.T,
) {
	normalizer := NewPostNormalizer()

	item, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			ID:        "abc123",
			Title:     "Toronto Community",
			Permalink: "/r/toronto/comments/abc123/example/",
		},
	)
	if err != nil {
		t.Fatalf(
			"normalize: %v",
			err,
		)
	}

	expected :=
		"https://www.reddit.com/r/toronto/comments/abc123/example/"

	if item.URL != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			item.URL,
		)
	}
}

func TestPostNormalizerRequiresTitle(t *testing.T) {
	normalizer := NewPostNormalizer()

	_, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			ID:    "abc123",
			Title: " ",
		},
	)

	if err == nil {
		t.Fatal(
			"expected missing title error",
		)
	}
}

func TestPostNormalizerPropagatesCancellation(
	t *testing.T,
) {
	normalizer := NewPostNormalizer()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := normalizer.Normalize(
		ctx,
		RedditPost{
			ID:    "abc123",
			Title: "Example",
		},
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

func TestPostNormalizerRejectsMissingID(t *testing.T) {
	normalizer := NewPostNormalizer()

	_, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			Title: "Toronto Community",
		},
	)

	if err == nil {
		t.Fatal(
			"expected missing ID error",
		)
	}
}

func TestPostNormalizerRejectsMissingTitle(t *testing.T) {
	normalizer := NewPostNormalizer()

	_, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			ID: "abc123",
		},
	)

	if err == nil {
		t.Fatal(
			"expected missing title error",
		)
	}
}

func TestPostNormalizerRejectsInvalidURL(t *testing.T) {
	normalizer := NewPostNormalizer()

	_, err := normalizer.Normalize(
		context.Background(),
		RedditPost{
			ID:    "abc123",
			Title: "Toronto Community",
			URL:   "ftp://example.com/post",
		},
	)

	if err == nil {
		t.Fatal(
			"expected invalid URL error",
		)
	}
}
