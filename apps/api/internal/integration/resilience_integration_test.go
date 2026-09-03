//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestResilienceIntegrationRetryExhaustionPreservesCause(
	t *testing.T,
) {
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

	policy, err := external.NewRetryPolicy(
		3,
		backoff,
	)
	if err != nil {
		t.Fatalf(
			"create retry policy: %v",
			err,
		)
	}

	var attempts int

	err = policy.Do(
		context.Background(),
		func(context.Context) error {
			attempts++

			return external.ErrUpstream
		},
		nil,
	)

	if !external.IsRetryExhausted(err) {
		t.Fatalf(
			"expected retry exhaustion, got %v",
			err,
		)
	}

	if !errors.Is(
		err,
		external.ErrUpstream,
	) {
		t.Fatalf(
			"expected original upstream cause, got %v",
			err,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"expected three attempts, got %d",
			attempts,
		)
	}
}

func TestResilienceIntegrationExpiredCacheFailsClosed(
	t *testing.T,
) {
	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	key := "resilience:test"
	value := []byte("stale")

	if err := memoryCache.Set(
		context.Background(),
		key,
		value,
		5*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"seed cache: %v",
			err,
		)
	}

	time.Sleep(
		15 * time.Millisecond,
	)

	_, ok, err := memoryCache.Get(
		context.Background(),
		key,
	)
	if err != nil {
		t.Fatalf(
			"read cache: %v",
			err,
		)
	}

	if ok {
		t.Fatal(
			"expired cache entry must not be served",
		)
	}
}
