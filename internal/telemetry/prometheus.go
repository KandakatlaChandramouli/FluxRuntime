package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus instrumentation for Phase 1A.
// All histograms use fixed bucket sets to bound cardinality.
type Metrics struct {
	ReservationLatency *prometheus.HistogramVec
	QueueSaturation    *prometheus.CounterVec
	OversellAttempts   prometheus.Counter
	RedisRTT           prometheus.Histogram
	WorkerQueueDepth   *prometheus.GaugeVec
}

var buckets = []float64{
	0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1,
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		ReservationLatency: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "reservation_latency_seconds",
			Help:    "End-to-end reservation latency in seconds.",
			Buckets: buckets,
		}, []string{"status"}),

		QueueSaturation: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "queue_saturation_total",
			Help: "Total requests rejected due to queue saturation.",
		}, []string{"shard_id"}),

		OversellAttempts: factory.NewCounter(prometheus.CounterOpts{
			Name: "oversell_attempts_total",
			Help: "Total inventory exhaustion hits (must remain 0 for actual oversells).",
		}),

		RedisRTT: factory.NewHistogram(prometheus.HistogramOpts{
			Name:    "redis_rtt_seconds",
			Help:    "Redis Lua script round-trip latency.",
			Buckets: buckets,
		}),

		WorkerQueueDepth: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "worker_queue_depth",
			Help: "Current depth of each worker's bounded queue.",
		}, []string{"shard_id"}),
	}
}
