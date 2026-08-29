package ratelimit

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestPacerFirstRequestIsImmediate(
	t *testing.T,
) {
	pacer, err := NewPacer(
		100 * time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create pacer: %v",
			err,
		)
	}

	start := time.Now()

	if err := pacer.Wait(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"wait: %v",
			err,
		)
	}

	if elapsed := time.Since(start); elapsed >=
		50*time.Millisecond {
		t.Fatalf(
			"expected immediate first request, took %s",
			elapsed,
		)
	}
}

func TestPacerSeparatesRequests(
	t *testing.T,
) {
	interval := 40 * time.Millisecond

	pacer, err := NewPacer(interval)
	if err != nil {
		t.Fatalf(
			"create pacer: %v",
			err,
		)
	}

	ctx := context.Background()

	if err := pacer.Wait(ctx); err != nil {
		t.Fatalf(
			"first wait: %v",
			err,
		)
	}

	start := time.Now()

	if err := pacer.Wait(ctx); err != nil {
		t.Fatalf(
			"second wait: %v",
			err,
		)
	}

	elapsed := time.Since(start)

	if elapsed < interval-5*time.Millisecond {
		t.Fatalf(
			"expected at least %s, got %s",
			interval,
			elapsed,
		)
	}
}

func TestPacerCancellationDoesNotConsumeSlot(
	t *testing.T,
) {
	interval := 50 * time.Millisecond

	pacer, err := NewPacer(interval)
	if err != nil {
		t.Fatalf(
			"create pacer: %v",
			err,
		)
	}

	if err := pacer.Wait(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"first wait: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err = pacer.Wait(ctx)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	start := time.Now()

	if err := pacer.Wait(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"third wait: %v",
			err,
		)
	}

	if elapsed := time.Since(start); elapsed <
		interval-5*time.Millisecond {
		t.Fatalf(
			"expected pacing after cancellation, got %s",
			elapsed,
		)
	}
}

func TestPacerSerializesConcurrentRequests(
	t *testing.T,
) {
	interval := 25 * time.Millisecond

	pacer, err := NewPacer(interval)
	if err != nil {
		t.Fatalf(
			"create pacer: %v",
			err,
		)
	}

	const requestCount = 4

	startTimes := make(
		[]time.Time,
		0,
		requestCount,
	)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	wg.Add(requestCount)

	for range requestCount {
		go func() {
			defer wg.Done()

			if err := pacer.Wait(
				context.Background(),
			); err != nil {
				t.Errorf(
					"wait: %v",
					err,
				)
				return
			}

			mu.Lock()

			startTimes = append(
				startTimes,
				time.Now(),
			)

			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(startTimes) != requestCount {
		t.Fatalf(
			"expected %d starts, got %d",
			requestCount,
			len(startTimes),
		)
	}

	sort.Slice(
		startTimes,
		func(i, j int) bool {
			return startTimes[i].Before(
				startTimes[j],
			)
		},
	)

	for i := 1; i < len(startTimes); i++ {
		elapsed := startTimes[i].Sub(
			startTimes[i-1],
		)

		if elapsed < interval-5*time.Millisecond {
			t.Fatalf(
				"requests too close together: %s",
				elapsed,
			)
		}
	}
}
