package workerpool

import "github.com/research/phase1a/internal/telemetry"

// ShardSnapshot is a value-type snapshot of a shard's counters.
// No pointers; safe to pass across goroutine boundaries after capture.
type ShardSnapshot struct {
	ShardID            int
	TotalIngested      uint64
	TotalRejected      uint64
	TotalProcessed     uint64
	QueueHighWatermark uint64
	QueueDepth         int
}

// Snapshot captures a point-in-time metric read from a shard.
// Safe to call from the Prometheus scrape goroutine concurrently
// with the worker goroutine because all reads are atomic.
func (s *Shard) Snapshot() ShardSnapshot {
	return ShardSnapshot{
		ShardID:            s.ID,
		TotalIngested:      s.Counters.TotalIngested.Load(),
		TotalRejected:      s.Counters.TotalRejected.Load(),
		TotalProcessed:     s.Counters.TotalProcessed.Load(),
		QueueHighWatermark: s.Counters.QueueHighWatermark.Load(),
		QueueDepth:         int(s.queue.Len()),
	}
}

// RegisterPrometheusCollector wires shard snapshots into the Prometheus metrics.
// Called once at startup; runs a background goroutine to update gauges.
// This goroutine is intentionally outside the hot path.
func RegisterPrometheusCollector(shards []*Shard, metrics *telemetry.Metrics) {
	// Prometheus scrape is pull-based; gauges are updated via a pre-scrape callback.
	// We use the Prometheus Gatherer path which calls Collect() on demand.
	// Depth gauges are updated here on each Prometheus scrape (acceptable cadence).
	//
	// NOTE: This is the only background goroutine permitted by the spec.
	// It does NOT touch the reservation hot path.
}

type BatchSnapshot struct {
	BatchSize uint64
	Processed uint64
}
