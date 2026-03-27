package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PaymentsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aurum_payments_created_total",
			Help: "Total number of payments created by currency",
		},
		[]string{"currency"},
	)

	PaymentTransitionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aurum_payment_transition_total",
			Help: "Total number of transitions on payment states",
		},
		[]string{"from", "to"},
	)

	OutboxPendingEvents = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "aurum_outbox_pending_events",
			Help: "Current number of unpublished outbox events.",
		},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aurum_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aurum_db_query_duration_seconds",
			Help:    "Database query duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

func TrackDB(operation string) func() {
	start := time.Now()
	return func() {
		DBQueryDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	}
}
