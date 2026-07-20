package metrics

import(
"runtime"
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