package metrics

import "github.com/prometheus/client_golang/prometheus"

var ExternalRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_external_requests_total",
		Help: "Total outbound external integration requests.",
	},
	[]string{
		"source",
		"operation",
		"status",
	},
)

var ExternalRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "village_external_request_duration_seconds",
		Help: "Outbound external integration request duration.",
	},
	[]string{
		"source",
		"operation",
	},
)

var ExternalErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "village_external_errors_total",
		Help: "Total outbound external integration errors.",
	},
	[]string{
		"source",
		"operation",
		"type",
	},
)
