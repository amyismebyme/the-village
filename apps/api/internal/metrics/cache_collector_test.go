package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestCacheCollectorReportsStats(
	t *testing.T,
) {
	collector := NewCacheCollector(
		func() CacheStats {
			return CacheStats{
				Entries:   3,
				Hits:      10,
				Misses:    4,
				Evictions: 2,
			}
		},
	)

	if collector == nil {
		t.Fatal(
			"expected collector",
		)
	}

	stats := collector.provider()

	if stats.Entries != 3 {
		t.Fatalf(
			"expected 3 entries, got %d",
			stats.Entries,
		)
	}

	if stats.Hits != 10 {
		t.Fatalf(
			"expected 10 hits, got %d",
			stats.Hits,
		)
	}

	if stats.Misses != 4 {
		t.Fatalf(
			"expected 4 misses, got %d",
			stats.Misses,
		)
	}

	if stats.Evictions != 2 {
		t.Fatalf(
			"expected 2 evictions, got %d",
			stats.Evictions,
		)
	}
}

func TestCacheMetricsExposeExpectedFamilies(
	t *testing.T,
) {
	registry := prometheus.NewRegistry()

	collector := NewCacheCollector(
		func() CacheStats {
			return CacheStats{
				Entries:   2,
				Hits:      5,
				Misses:    3,
				Evictions: 1,
			}
		},
	)

	if err := registry.Register(
		collector,
	); err != nil {
		t.Fatalf(
			"register collector: %v",
			err,
		)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	expected := map[string]bool{
		"village_cache_entries":         false,
		"village_cache_hits_total":      false,
		"village_cache_misses_total":    false,
		"village_cache_evictions_total": false,
	}

	for _, family := range families {
		if _, ok := expected[family.GetName()]; ok {
			expected[family.GetName()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf(
				"expected metric family %q",
				name,
			)
		}
	}
}
