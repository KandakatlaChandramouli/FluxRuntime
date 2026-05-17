package telemetry

import (
	"sync/atomic"
	"time"
)

// LatencyTracker records nanosecond-resolution latency samples
// using atomic accumulators only. No heap allocation on record path.
type LatencyTracker struct {
	totalNs     atomic.Int64
	sampleCount atomic.Int64
}

// Record stores a single latency observation.
func (t *LatencyTracker) Record(start time.Time) {
	ns := time.Since(start).Nanoseconds()
	t.totalNs.Add(ns)
	t.sampleCount.Add(1)
}

// MeanNs returns mean latency in nanoseconds.
// Returns 0 if no samples recorded.
func (t *LatencyTracker) MeanNs() int64 {
	count := t.sampleCount.Load()
	if count == 0 {
		return 0
	}
	return t.totalNs.Load() / count
}

// SampleCount returns the total number of recorded samples.
func (t *LatencyTracker) SampleCount() int64 {
	return t.sampleCount.Load()
}
