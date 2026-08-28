package metrics

import "github.com/prometheus/client_golang/prometheus"

var WorkerRunsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_worker_runs_total",
		Help: "Total background worker runs.",
	},
	[]string{
		"worker",
		"status",
	},
)

var WorkerFailuresTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_worker_failures_total",
		Help: "Total background worker run failures.",
	},
	[]string{
		"worker",
	},
)

var WorkerDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "village_worker_duration_seconds",
		Help: "Background worker run duration.",
	},
	[]string{
		"worker",
	},
)

var WorkersInFlight = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "village_worker_in_flight",
		Help: "Current number of active background worker runs.",
	},
	[]string{
		"worker",
	},
)
