package telemetry

import (
	"sort"
	"sync"
	"sync/atomic"
)

type RuntimeStats struct {
	Completed atomic.Uint64
	Rejected  atomic.Uint64

	TotalLatencyNs atomic.Uint64
	MaxLatencyNs   atomic.Uint64

	mu sync.Mutex

	latencies []uint64

	BatchCount atomic.Uint64
	BatchTotal atomic.Uint64
}

func (r *RuntimeStats) RecordLatency(ns uint64) {

	r.Completed.Add(1)

	r.TotalLatencyNs.Add(ns)

	for {
		current := r.MaxLatencyNs.Load()

		if ns <= current {
			break
		}

		if r.MaxLatencyNs.CompareAndSwap(current, ns) {
			break
		}
	}

	r.mu.Lock()
	r.latencies = append(r.latencies, ns)
	r.mu.Unlock()
}

func (r *RuntimeStats) RecordBatch(size uint64) {

	r.BatchCount.Add(1)
	r.BatchTotal.Add(size)
}

type RuntimeSnapshot struct {
	Completed uint64
	Rejected  uint64

	AvgLatencyNs uint64
	MaxLatencyNs uint64

	P50 uint64
	P95 uint64
	P99 uint64

	AvgBatchSize float64
}

func (r *RuntimeStats) Snapshot() RuntimeSnapshot {

	completed := r.Completed.Load()

	totalLatency := r.TotalLatencyNs.Load()

	var avgLatency uint64

	if completed > 0 {
		avgLatency = totalLatency / completed
	}

	batchCount := r.BatchCount.Load()
	batchTotal := r.BatchTotal.Load()

	var avgBatch float64

	if batchCount > 0 {
		avgBatch = float64(batchTotal) / float64(batchCount)
	}

	r.mu.Lock()

	cp := make([]uint64, len(r.latencies))
	copy(cp, r.latencies)

	r.mu.Unlock()

	sort.Slice(cp, func(i, j int) bool {
		return cp[i] < cp[j]
	})

	p50 := percentile(cp, 0.50)
	p95 := percentile(cp, 0.95)
	p99 := percentile(cp, 0.99)

	return RuntimeSnapshot{
		Completed: completed,
		Rejected:  r.Rejected.Load(),

		AvgLatencyNs: avgLatency,
		MaxLatencyNs: r.MaxLatencyNs.Load(),

		P50: p50,
		P95: p95,
		P99: p99,

		AvgBatchSize: avgBatch,
	}
}

func percentile(data []uint64, p float64) uint64 {

	if len(data) == 0 {
		return 0
	}

	idx := int(float64(len(data)-1) * p)

	return data[idx]
}

var GlobalRuntimeStats RuntimeStats
