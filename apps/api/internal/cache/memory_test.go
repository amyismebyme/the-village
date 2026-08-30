package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemorySetAndGet(t *testing.T) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	value := []byte("hello")

	if err := cache.Set(
		context.Background(),
		"test-key",
		value,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set: %v",
			err,
		)
	}

	value[0] = 'X'

	got, ok, err := cache.Get(
		context.Background(),
		"test-key",
	)
	if err != nil {
		t.Fatalf(
			"get: %v",
			err,
		)
	}

	if !ok {
		t.Fatal(
			"expected cache hit",
		)
	}

	if string(got) != "hello" {
		t.Fatalf(
			"expected hello, got %q",
			string(got),
		)
	}

	got[0] = 'Y'

	gotAgain, ok, err := cache.Get(
		context.Background(),
		"test-key",
	)
	if err != nil {
		t.Fatalf(
			"second get: %v",
			err,
		)
	}

	if !ok {
		t.Fatal(
			"expected second cache hit",
		)
	}

	if string(gotAgain) != "hello" {
		t.Fatalf(
			"cached value was mutated: %q",
			string(gotAgain),
		)
	}
}

func TestMemoryTTLExpiry(t *testing.T) {
	now := time.Now()

	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	cache.now = func() time.Time {
		return now
	}

	if err := cache.Set(
		context.Background(),
		"test",
		[]byte("hello"),
		time.Second,
	); err != nil {
		t.Fatalf(
			"set: %v",
			err,
		)
	}

	_, ok, err := cache.Get(
		context.Background(),
		"test",
	)
	if err != nil {
		t.Fatalf(
			"get before expiry: %v",
			err,
		)
	}

	if !ok {
		t.Fatal(
			"expected hit",
		)
	}

	now = now.Add(
		2 * time.Second,
	)

	_, ok, err = cache.Get(
		context.Background(),
		"test",
	)
	if err != nil {
		t.Fatalf(
			"get after expiry: %v",
			err,
		)
	}

	if ok {
		t.Fatal(
			"expected expired entry to be removed",
		)
	}
}

func TestMemoryEvictsLeastRecentlyUsed(
	t *testing.T,
) {
	cache, err := NewMemory(2)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx := context.Background()

	if err := cache.Set(
		ctx,
		"a",
		[]byte("A"),
		time.Minute,
	); err != nil {
		t.Fatalf("set a: %v", err)
	}

	if err := cache.Set(
		ctx,
		"b",
		[]byte("B"),
		time.Minute,
	); err != nil {
		t.Fatalf("set b: %v", err)
	}

	// Make a the most recently used entry.
	if _, ok, err := cache.Get(
		ctx,
		"a",
	); err != nil || !ok {
		t.Fatalf(
			"expected a hit, ok=%v, err=%v",
			ok,
			err,
		)
	}

	if err := cache.Set(
		ctx,
		"c",
		[]byte("C"),
		time.Minute,
	); err != nil {
		t.Fatalf("set c: %v", err)
	}

	if _, ok, err := cache.Get(
		ctx,
		"b",
	); err != nil {
		t.Fatalf(
			"get b: %v",
			err,
		)
	} else if ok {
		t.Fatal(
			"expected b to be evicted",
		)
	}

	if _, ok, err := cache.Get(
		ctx,
		"a",
	); err != nil || !ok {
		t.Fatalf(
			"expected a to remain cached, ok=%v, err=%v",
			ok,
			err,
		)
	}

	if _, ok, err := cache.Get(
		ctx,
		"c",
	); err != nil || !ok {
		t.Fatalf(
			"expected c to remain cached, ok=%v, err=%v",
			ok,
			err,
		)
	}
}

func TestMemoryDelete(t *testing.T) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx := context.Background()

	if err := cache.Set(
		ctx,
		"test",
		[]byte("hello"),
		time.Minute,
	); err != nil {
		t.Fatalf(
			"set: %v",
			err,
		)
	}

	if err := cache.Delete(
		ctx,
		"test",
	); err != nil {
		t.Fatalf(
			"delete: %v",
			err,
		)
	}

	_, ok, err := cache.Get(
		ctx,
		"test",
	)
	if err != nil {
		t.Fatalf(
			"get: %v",
			err,
		)
	}

	if ok {
		t.Fatal(
			"expected cache miss",
		)
	}
}

func TestMemoryClear(t *testing.T) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx := context.Background()

	for _, key := range []string{
		"one",
		"two",
		"three",
	} {
		if err := cache.Set(
			ctx,
			key,
			[]byte(key),
			time.Minute,
		); err != nil {
			t.Fatalf(
				"set %s: %v",
				key,
				err,
			)
		}
	}

	if err := cache.Clear(ctx); err != nil {
		t.Fatalf(
			"clear: %v",
			err,
		)
	}

	for _, key := range []string{
		"one",
		"two",
		"three",
	} {
		_, ok, err := cache.Get(
			ctx,
			key,
		)
		if err != nil {
			t.Fatalf(
				"get %s: %v",
				key,
				err,
			)
		}

		if ok {
			t.Fatalf(
				"expected %s to be absent",
				key,
			)
		}
	}
}

func TestMemoryRejectsInvalidMaxEntries(
	t *testing.T,
) {
	if _, err := NewMemory(0); err == nil {
		t.Fatal(
			"expected max entries error",
		)
	}
}

func TestMemoryRejectsInvalidTTL(
	t *testing.T,
) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	err = cache.Set(
		context.Background(),
		"test",
		[]byte("hello"),
		0,
	)

	if err == nil {
		t.Fatal(
			"expected invalid TTL error",
		)
	}
}

func TestMemoryRejectsCanceledContext(
	t *testing.T,
) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, _, err = cache.Get(
		ctx,
		"test",
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

func TestMemoryConcurrentSetGet(
	t *testing.T,
) {
	cache, err := NewMemory(100)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	const (
		goroutines = 20
		iterations = 200
	)

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			ctx := context.Background()

			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf(
					"key-%d",
					j%10,
				)

				if err := cache.Set(
					ctx,
					key,
					[]byte("value"),
					time.Minute,
				); err != nil {
					t.Errorf(
						"set %q: %v",
						key,
						err,
					)

					return
				}

				_, _, err := cache.Get(
					ctx,
					key,
				)
				if err != nil {
					t.Errorf(
						"get %q: %v",
						key,
						err,
					)

					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryConcurrentDelete(
	t *testing.T,
) {
	cache, err := NewMemory(100)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx := context.Background()

	for i := 0; i < 20; i++ {
		if err := cache.Set(
			ctx,
			fmt.Sprintf("key-%d", i),
			[]byte("value"),
			time.Minute,
		); err != nil {
			t.Fatalf(
				"seed key: %v",
				err,
			)
		}
	}

	var wg sync.WaitGroup

	wg.Add(20)

	for i := 0; i < 20; i++ {
		go func(i int) {
			defer wg.Done()

			if err := cache.Delete(
				ctx,
				fmt.Sprintf("key-%d", i),
			); err != nil {
				t.Errorf(
					"delete: %v",
					err,
				)
			}
		}(i)
	}

	wg.Wait()

	stats := cache.Stats()

	if stats.Entries != 0 {
		t.Fatalf(
			"expected zero entries, got %d",
			stats.Entries,
		)
	}
}

func TestMemoryConcurrentClear(
	t *testing.T,
) {
	cache, err := NewMemory(100)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	var wg sync.WaitGroup

	wg.Add(20)

	for i := 0; i < 20; i++ {
		go func(i int) {
			defer wg.Done()

			ctx := context.Background()

			for j := 0; j < 50; j++ {
				key := fmt.Sprintf(
					"key-%d-%d",
					i,
					j,
				)

				_ = cache.Set(
					ctx,
					key,
					[]byte("value"),
					time.Minute,
				)

				if j%10 == 0 {
					_ = cache.Clear(ctx)
				}
			}
		}(i)
	}

	wg.Wait()

	if stats := cache.Stats(); stats.Entries >
		stats.MaxItems {
		t.Fatalf(
			"cache exceeded maximum entries: %d > %d",
			stats.Entries,
			stats.MaxItems,
		)
	}
}

func TestMemoryConcurrentLRUAccounting(
	t *testing.T,
) {
	cache, err := NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	const workers = 30

	var wg sync.WaitGroup

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			ctx := context.Background()

			for j := 0; j < 100; j++ {
				key := fmt.Sprintf(
					"key-%d",
					(i+j)%20,
				)

				_ = cache.Set(
					ctx,
					key,
					[]byte("value"),
					time.Minute,
				)

				_, _, _ = cache.Get(
					ctx,
					key,
				)
			}
		}(i)
	}

	wg.Wait()

	stats := cache.Stats()

	if stats.Entries > stats.MaxItems {
		t.Fatalf(
			"cache exceeded maximum entries: %d > %d",
			stats.Entries,
			stats.MaxItems,
		)
	}

	if stats.Evictions == 0 {
		t.Fatal(
			"expected evictions under cache pressure",
		)
	}
}

func TestMemoryStatsRemainConsistent(
	t *testing.T,
) {
	cache, err := NewMemory(3)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	ctx := context.Background()

	for i := 0; i < 20; i++ {
		if err := cache.Set(
			ctx,
			fmt.Sprintf("key-%d", i),
			[]byte("value"),
			time.Minute,
		); err != nil {
			t.Fatalf(
				"set: %v",
				err,
			)
		}
	}

	stats := cache.Stats()

	if stats.Entries < 0 {
		t.Fatalf(
			"invalid negative entry count",
		)
	}

	if stats.Entries > stats.MaxItems {
		t.Fatalf(
			"entries exceed maximum: %d > %d",
			stats.Entries,
			stats.MaxItems,
		)
	}

	if stats.Evictions != 17 {
		t.Fatalf(
			"expected 17 evictions, got %d",
			stats.Evictions,
		)
	}
}
