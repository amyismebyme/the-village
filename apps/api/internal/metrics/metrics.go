package metrics

import (
	"runtime"
	"sync"

	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/prometheus/client_golang/prometheus"
)

var RequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_http_requests_total",
		Help: "Total HTTP requests.",
	},
	[]string{
		"method",
		"path",
		"status",
	},
)

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "village_http_request_duration_seconds",
		Help: "HTTP request latency.",
		Buckets: []float64{
			0.001,
			0.0025,
			0.005,
			0.01,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
			2.5,
			5,
			10,
		},
	},
	[]string{
		"method",
		"path",
	},
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
	[]string{
		"type",
	},
)

var DatabaseQueriesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_db_queries_total",
		Help: "Total database queries.",
	},
	[]string{
		"operation",
	},
)

var DatabaseQueryDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "village_db_query_duration_seconds",
		Help:    "Database query latency.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{
		"operation",
	},
)

var BuildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "village_build_info",
		Help: "Build and runtime information for the running application.",
	},
	[]string{
		"version",
		"git_commit",
		"go_version",
		"environment",
	},
)

var registerOnce sync.Once

func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			RequestsTotal,
			RequestDuration,
			RequestsInFlight,
			PanicsTotal,
			ErrorsTotal,
			DatabaseQueriesTotal,
			DatabaseQueryDuration,
			BuildInfo,

			// Standard Go & process metrics.
			//prometheus.NewGoCollector(),
			//prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		)

		BuildInfo.WithLabelValues(
			appruntime.BuildVersion,
			appruntime.GitCommit,
			runtime.Version(),
			appruntime.Environment,
		).Set(1)
	})
}