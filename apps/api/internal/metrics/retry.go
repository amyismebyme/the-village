package metrics

import "github.com/prometheus/client_golang/prometheus"

var ExternalRetriesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_external_retries_total",
		Help: "Total retries of outbound external integration requests.",
	},
	[]string{
		"source",
		"operation",
		"error_type",
	},
)

var ExternalRetryDelay = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "village_external_retry_delay_seconds",
		Help: "Delay before retrying outbound external integration requests.",
		Buckets: []float64{
			0.001,
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
			30,
		},
	},
	[]string{
		"source",
		"operation",
	},
)

var ExternalRetryExhaustedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_external_retry_exhausted_total",
		Help: "Total outbound external operations that exhausted their retry budget.",
	},
	[]string{
		"source",
		"operation",
	},
)
