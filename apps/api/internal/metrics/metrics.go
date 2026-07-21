package metrics

import (
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"runtime"
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
		Name:    "village_http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{
		"method",
		"path",
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

func Register() {

	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		BuildInfo,
	)
	BuildInfo.WithLabelValues(
		appruntime.BuildVersion,
		appruntime.GitCommit,
		runtime.Version(),
		appruntime.Environment,
	).Set(1)

}
