package reddit

import (
	"context"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"testing"
)

func BenchmarkNormalizeRedditPost(b *testing.B) {
	normalizer := NewPostNormalizer()
	post := RedditPost{
		ID:         "benchmark-123",
		Title:      "Benchmark Reddit post",
		SelfText:   "A representative post body for normalization.",
		URL:        "https://www.reddit.com/r/toronto/comments/benchmark-123/example/",
		Permalink:  "/r/toronto/comments/benchmark-123/example/",
		Subreddit:  "toronto",
		CreatedUTC: 1750000000,
	}
	ctx := context.Background()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := normalizer.Normalize(ctx, post); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeduplicateExternalItems(b *testing.B) {
	items := make([]external.Item, 0, 100)
	for i := 0; i < 100; i++ {
		items = append(items, external.Item{
			Source:     external.SourceReddit,
			ExternalID: fmt.Sprintf("item-%03d", i%26),
			Title:      "benchmark item",
		})
	}

	// Add deterministic duplicates.
	items = append(items,
		items[0], items[1], items[2], items[3], items[4],
	)

	ctx := context.Background()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := external.DeduplicateItems(ctx, items); err != nil {
			b.Fatal(err)
		}
	}
}
