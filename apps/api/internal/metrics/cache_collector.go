package metrics

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type CacheStats struct {
	Entries     int
	Hits        uint64
	Misses      uint64
	Evictions   uint64
	Sets        uint64
	Deletes     uint64
	Expirations uint64
}

type CacheStatsProvider func() CacheStats

type CacheCollector struct {
	provider CacheStatsProvider

	entries     *prometheus.Desc
	hits        *prometheus.Desc
	misses      *prometheus.Desc
	evictions   *prometheus.Desc
	sets        *prometheus.Desc
	deletes     *prometheus.Desc
	expirations *prometheus.Desc
}

func NewCacheCollector(
	provider CacheStatsProvider,
) *CacheCollector {
	return &CacheCollector{
		provider: provider,

		entries: prometheus.NewDesc(
			"village_cache_entries",
			"Current number of cache entries.",
			nil,
			nil,
		),

		hits: prometheus.NewDesc(
			"village_cache_hits_total",
			"Total cache hits.",
			nil,
			nil,
		),

		misses: prometheus.NewDesc(
			"village_cache_misses_total",
			"Total cache misses.",
			nil,
			nil,
		),

		sets: prometheus.NewDesc(
			"village_cache_sets_total",
			"Total cache set operations.",
			nil,
			nil,
		),

		deletes: prometheus.NewDesc(
			"village_cache_deletes_total",
			"Total cache delete operations that removed an entry.",
			nil,
			nil,
		),

		expirations: prometheus.NewDesc(
			"village_cache_expirations_total",
			"Total cache entries found expired during access.",
			nil,
			nil,
		),

		evictions: prometheus.NewDesc(
			"village_cache_evictions_total",
			"Total cache entries evicted due to capacity limits.",
			nil,
			nil,
		),
	}
}

func (c *CacheCollector) Describe(
	ch chan<- *prometheus.Desc,
) {
	ch <- c.entries
	ch <- c.hits
	ch <- c.misses
	ch <- c.evictions
	ch <- c.sets
	ch <- c.deletes
	ch <- c.expirations
}

func (c *CacheCollector) Collect(
	ch chan<- prometheus.Metric,
) {
	stats := c.provider()

	ch <- prometheus.MustNewConstMetric(
		c.entries,
		prometheus.GaugeValue,
		float64(stats.Entries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.hits,
		prometheus.CounterValue,
		float64(stats.Hits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.misses,
		prometheus.CounterValue,
		float64(stats.Misses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.evictions,
		prometheus.CounterValue,
		float64(stats.Evictions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.sets,
		prometheus.CounterValue,
		float64(stats.Sets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.deletes,
		prometheus.CounterValue,
		float64(stats.Deletes),
	)

	ch <- prometheus.MustNewConstMetric(
		c.expirations,
		prometheus.CounterValue,
		float64(stats.Expirations),
	)
}

func RegisterCache(
	reg prometheus.Registerer,
	provider CacheStatsProvider,
) error {
	if reg == nil {
		return errors.New(
			"metrics: registerer is required",
		)
	}

	if provider == nil {
		return errors.New(
			"metrics: cache stats provider is required",
		)
	}

	if err := reg.Register(
		NewCacheCollector(provider),
	); err != nil {
		var alreadyRegistered prometheus.AlreadyRegisteredError

		if errors.As(
			err,
			&alreadyRegistered,
		) {
			return nil
		}

		return fmt.Errorf(
			"register cache metrics: %w",
			err,
		)
	}

	return nil
}
