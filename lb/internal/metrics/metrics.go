package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts the total number of requests handled by the load balancer.
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "teeter_requests_total",
			Help: "The total number of requests handled by Teeter",
		},
		[]string{"route", "method", "status"},
	)

	// RequestDuration tracks the latency of requests forwarded to backends.
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "teeter_request_duration_seconds",
			Help:    "Latency of requests forwarded to backends (seconds)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route"},
	)

	// BackendStatus indicates the health status of a specific backend.
	// 1 = Up, 0 = Down
	BackendStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "teeter_backend_status",
			Help: "Health status of a backend (1 for UP, 0 for DOWN)",
		},
		[]string{"route", "url"},
	)

	// ActiveConnections tracks the number of currently active connections to a backend.
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "teeter_active_connections",
			Help: "Number of active connections to a backend",
		},
		[]string{"route", "url"},
	)
)
