//go:build smoke

package reddit

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestRedditLiveSmoke(t *testing.T) {
	clientID := strings.TrimSpace(os.Getenv("REDDIT_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("REDDIT_CLIENT_SECRET"))
	userAgent := strings.TrimSpace(os.Getenv("REDDIT_USER_AGENT"))

	if clientID == "" || clientSecret == "" || userAgent == "" {
		t.Skip("set REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET, and REDDIT_USER_AGENT to run live Reddit smoke test")
	}

	subreddit := strings.TrimSpace(os.Getenv("REDDIT_INGEST_SUBREDDIT"))
	if subreddit == "" {
		subreddit = "toronto"
	}

	limit := 5
	requestTimeout := 15 * time.Second

	httpClient := &http.Client{}

	authenticator, err := NewAuthenticator(
		httpClient,
		"https://www.reddit.com",
		clientID,
		clientSecret,
		userAgent,
		requestTimeout,
	)
	if err != nil {
		t.Fatalf("create Reddit authenticator: %v", err)
	}

	client, err := NewClient(
		httpClient,
		"https://oauth.reddit.com",
		userAgent,
		requestTimeout,
	)
	if err != nil {
		t.Fatalf("create Reddit client: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		45*time.Second,
	)
	defer cancel()

	token, err := authenticator.Token(ctx)
	if err != nil {
		t.Fatalf("authenticate with Reddit: %v", err)
	}

	service := NewIngestionService(
		client,
		NewPostNormalizer(),
	)

	items, err := service.IngestListing(
		ctx,
		token,
		subreddit,
		limit,
		"",
	)
	if err != nil {
		t.Fatalf("fetch and normalize Reddit listing: %v", err)
	}

	for i, item := range items {
		if err := item.Validate(); err != nil {
			t.Fatalf("validate live item %d: %v", i, err)
		}
	}

	if len(items) == 0 {
		t.Fatalf("Reddit returned zero normalized items for r/%s", subreddit)
	}

	identities := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := item.Identity().Key()
		if _, exists := identities[key]; exists {
			t.Fatalf("duplicate identity survived live ingestion: %s", key)
		}
		identities[key] = struct{}{}
	}

	t.Logf("Reddit live smoke passed: source=%s subreddit=%s items=%d", external.SourceReddit, subreddit, len(items))
}
