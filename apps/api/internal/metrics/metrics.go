package metrics

import (
	"runtime"
	"sync"

	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var RequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_http_requests_total",
		Help: "Total HTTP requests..",
	},
	[]string{"method", "route", "status"},
)

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "village_http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"method", "route"},
)

var RequestsInFlight = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "village_http_requests_in_flight",
		Help: "Current number of HTTP requests being processed.",
	},
)

var PanicsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "village_panics_total",
		Help: "Total recovered panics.",
	},
)

var ErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_errors_total",
		Help: "Total application errors.",
	},
	[]string{"type"},
)

var DatabaseQueriesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_db_queries_total",
		Help: "Total database queries by operation and status.",
	},
	[]string{
		"operation",
		"status",
	},
)

var DatabaseQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "village_db_query_duration_seconds",
		Help: "Database query duration in seconds.",
	},
	[]string{
		"operation",
	},
)

var BuildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "village_build_info",
		Help: "Build and runtime information.",
	},
	[]string{
		"version",
		"git_commit",
		"go_version",
		"environment",
	},
)

var registerOnce sync.Once

func Register(reg prometheus.Registerer, pool *pgxpool.Pool) {
	registerOnce.Do(func() {
		collectors := []prometheus.Collector{
			RequestsTotal,
			RequestDuration,
			RequestsInFlight,
			PanicsTotal,
			ErrorsTotal,
			DatabaseQueriesTotal,
			DatabaseQueryDuration,
			BuildInfo,
		}

		if pool != nil {
			collectors = append(collectors, NewPoolCollector(pool))
		}

		reg.MustRegister(collectors...)

		BuildInfo.WithLabelValues(
			appruntime.BuildVersion,
			appruntime.GitCommit,
			runtime.Version(),
			appruntime.Environment,
		).Set(1)
	})
}

type PoolCollector struct {
	pool *pgxpool.Pool

	acquired *prometheus.Desc
	idle     *prometheus.Desc
	total    *prometheus.Desc
	max      *prometheus.Desc

	acquireCount      *prometheus.Desc
	acquireDuration   *prometheus.Desc
	emptyAcquireCount *prometheus.Desc
	constructing      *prometheus.Desc
}

func NewPoolCollector(pool *pgxpool.Pool) *PoolCollector {
	return &PoolCollector{
		pool: pool,

		acquired: prometheus.NewDesc(
			"village_db_pool_acquired_connections",
			"Number of acquired connections.",
			nil,
			nil,
		),

		idle: prometheus.NewDesc(
			"village_db_pool_idle_connections",
			"Number of idle connections.",
			nil,
			nil,
		),

		total: prometheus.NewDesc(
			"village_db_pool_total_connections",
			"Total connections in the pool.",
			nil,
			nil,
		),

		acquireCount: prometheus.NewDesc(
			"village_db_pool_acquire_total",
			"Total successful connection acquisitions.",
			nil,
			nil,
		),

		acquireDuration: prometheus.NewDesc(
			"village_db_pool_acquire_duration_ms_total",
			"Total acquisition wait time in milliseconds.",
			nil,
			nil,
		),

		emptyAcquireCount: prometheus.NewDesc(
			"village_db_pool_empty_acquire_total",
			"Number of waits caused by exhausted pool.",
			nil,
			nil,
		),

		constructing: prometheus.NewDesc(
			"village_db_pool_constructing_connections",
			"Connections currently being established.",
			nil,
			nil,
		),

		max: prometheus.NewDesc(
			"village_db_pool_max_connections",
			"Configured maximum connections.",
			nil,
			nil,
		),
	}
}

func (c *PoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquired
	ch <- c.idle
	ch <- c.total
	ch <- c.max

	ch <- c.acquireCount
	ch <- c.acquireDuration
	ch <- c.emptyAcquireCount
	ch <- c.constructing
}

func (c *PoolCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(
		c.acquired,
		prometheus.GaugeValue,
		float64(s.AcquiredConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.idle,
		prometheus.GaugeValue,
		float64(s.IdleConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.total,
		prometheus.GaugeValue,
		float64(s.TotalConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.acquireCount,
		prometheus.CounterValue,
		float64(s.AcquireCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.acquireDuration,
		prometheus.CounterValue,
		float64(s.AcquireDuration().Milliseconds()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.emptyAcquireCount,
		prometheus.CounterValue,
		float64(s.EmptyAcquireCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.constructing,
		prometheus.GaugeValue,
		float64(s.ConstructingConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.max,
		prometheus.GaugeValue,
		float64(s.MaxConns()),
	)
}
