package ratelimit

import (
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestPerSourceRegistersIndependentLimiters(
	t *testing.T,
) {
	registry := NewPerSource()

	redditLimiter, err := registry.Register(
		external.SourceReddit,
		50*time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"register Reddit limiter: %v",
			err,
		)
	}

	otherLimiter, err := registry.Register(
		external.Source("other"),
		100*time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"register other limiter: %v",
			err,
		)
	}

	if redditLimiter == otherLimiter {
		t.Fatal(
			"expected independent limiters",
		)
	}

	got, ok := registry.Get(
		external.SourceReddit,
	)
	if !ok {
		t.Fatal(
			"expected Reddit limiter",
		)
	}

	if got != redditLimiter {
		t.Fatal(
			"unexpected Reddit limiter",
		)
	}
}

func TestPerSourceRejectsDuplicateSource(
	t *testing.T,
) {
	registry := NewPerSource()

	if _, err := registry.Register(
		external.SourceReddit,
		time.Second,
	); err != nil {
		t.Fatalf(
			"first registration failed: %v",
			err,
		)
	}

	if _, err := registry.Register(
		external.SourceReddit,
		time.Second,
	); err == nil {
		t.Fatal(
			"expected duplicate registration error",
		)
	}
}

func TestPerSourceMissingSource(
	t *testing.T,
) {
	registry := NewPerSource()

	if _, ok := registry.Get(
		external.Source("missing"),
	); ok {
		t.Fatal(
			"expected source to be missing",
		)
	}
}
