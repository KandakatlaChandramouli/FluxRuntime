package workerpool

import (
	"context"
	"hash/fnv"
	"sync/atomic"
	"testing"

	"github.com/research/phase1a/internal/config"
)

// newBenchDispatcher constructs a Dispatcher with real Shards and no worker
// goroutines. Identical to the pattern used in workerpool_test.go:
// TestHotKeyQueueSaturation. No Redis. No Prometheus. No network.
func newBenchDispatcher(workerCount int) *Dispatcher {
	shards := make([]*Shard, workerCount)
	for i := 0; i < workerCount; i++ {
		shards[i] = newShard(i)
	}
	return &Dispatcher{
		shards:      shards,
		workerCount: workerCount,
	}
}

// benchWorkerCount mirrors the production formula from config.RuntimeConfig.
const benchWorkerCount = 8

// benchCtx is a single shared background context for all benchmark requests.
// context.Background() allocates once; reusing it eliminates per-iteration
// allocation noise from the context path.
var benchCtx = context.Background()

// benchResponseCh is a pre-allocated, drained response channel reused across
// benchmark iterations. Workers are not running, so nothing ever sends on it.
// Capacity 1 matches production handler allocation.
var benchResponseCh = make(chan ReservationResult, 1)

// BenchmarkDispatchHotKey routes every request to the same EventID.
// All traffic collapses onto one shard.
// Measures: contention amplification on a single shard queue.
func BenchmarkDispatchHotKey(b *testing.B) {
	d := newBenchDispatcher(benchWorkerCount)

	const hotEventID uint64 = 101

	req := ReservationRequest{
		EventID:    hotEventID,
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: benchResponseCh,
		Ctx:        benchCtx,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.Dispatch(req)
	}
}

// BenchmarkDispatchColdKeys distributes requests across 256 distinct EventIDs.
// Measures: balanced routing throughput across all shards.
func BenchmarkDispatchColdKeys(b *testing.B) {
	d := newBenchDispatcher(benchWorkerCount)

	// Fixed 256-entry table of distinct uint64 EventIDs.
	// Index with i&255 — no append, no allocation inside the loop.
	var coldKeys [256]uint64
	for i := range coldKeys {
		coldKeys[i] = uint64(i + 1)
	}

	req := ReservationRequest{
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: benchResponseCh,
		Ctx:        benchCtx,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req.EventID = coldKeys[i&255]
		d.Dispatch(req)
	}
}

// BenchmarkSaturatedDispatch pre-fills every shard queue to exact capacity,
// then benchmarks only the rejection path.
// Measures: non-blocking rejection efficiency under full saturation.
func BenchmarkSaturatedDispatch(b *testing.B) {
	d := newBenchDispatcher(benchWorkerCount)

	floodReq := ReservationRequest{
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: benchResponseCh,
		Ctx:        benchCtx,
	}

	// Fill every shard to exactly config.QueueCapacity.
	// Each shard owns one bounded chan ReservationRequest of that capacity.
	// We route one distinct EventID per shard to guarantee even fill.
	// shardIndex is deterministic, so we probe until each shard is saturated.
	for i := 0; i < benchWorkerCount; i++ {
		// Find an EventID that routes to shard i.
		var eventID uint64 = 1
		for {
			if d.shardIndex(eventID) == i {
				break
			}
			eventID++
		}
		floodReq.EventID = eventID
		for j := 0; j < config.QueueCapacity; j++ {
			_ = d.shards[i].queue.Enqueue(floodReq)
		}
	}

	// All queues are now at capacity. Benchmark only the rejection path.
	const hotEventID uint64 = 101
	req := ReservationRequest{
		EventID:    hotEventID,
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: benchResponseCh,
		Ctx:        benchCtx,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d.Dispatch(req)
	}
}

// BenchmarkShardRouting isolates the cost of shardIndex(uint64).
// No queue operations. No worker execution. No channel touches.
// Measures: FNV-1a hash + modulo overhead per routing decision.
//
// shardIndex uses fnv.New64a() which allocates a hash.Hash64 on each call.
// This benchmark intentionally exposes that allocator pressure honestly —
// per the spec directive: "benchmark should expose allocator pressure honestly".
func BenchmarkShardRouting(b *testing.B) {
	d := newBenchDispatcher(benchWorkerCount)

	var coldKeys [256]uint64
	for i := range coldKeys {
		coldKeys[i] = uint64(i + 1)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = d.shardIndex(coldKeys[i&255])
	}
}

// BenchmarkShardRoutingInlined isolates the FNV-1a hash cost with the
// hash object stack-allocated rather than heap-allocated, as a comparison
// data point against BenchmarkShardRouting.
// This exposes the allocation cost of fnv.New64a() in the production path.
func BenchmarkShardRoutingInlined(b *testing.B) {
	var coldKeys [256]uint64
	for i := range coldKeys {
		coldKeys[i] = uint64(i + 1)
	}

	workerCount := uint64(benchWorkerCount)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		eventID := coldKeys[i&255]
		h := fnv.New64a()
		var buf [8]byte
		buf[0] = byte(eventID >> 56)
		buf[1] = byte(eventID >> 48)
		buf[2] = byte(eventID >> 40)
		buf[3] = byte(eventID >> 32)
		buf[4] = byte(eventID >> 24)
		buf[5] = byte(eventID >> 16)
		buf[6] = byte(eventID >> 8)
		buf[7] = byte(eventID)
		h.Write(buf[:])
		_ = int(h.Sum64() % workerCount)
	}
}

// BenchmarkAtomicCounters isolates sync/atomic contention cost.
// Runs parallel goroutines all incrementing a single shared counter.
// Measures: atomic CAS contention scalability under GOMAXPROCS pressure.
func BenchmarkAtomicCounters(b *testing.B) {
	var counter atomic.Uint64

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Add(1)
		}
	})
}

// BenchmarkQueueEnqueue isolates bounded channel enqueue overhead.
// Uses a direct chan ReservationRequest sized to config.QueueCapacity.
// On full queue: drains one slot then re-enqueues to maintain steady state.
// Measures: channel send latency under bounded queue semantics.
func BenchmarkQueueEnqueue(b *testing.B) {
	q := make(chan ReservationRequest, config.QueueCapacity)

	req := ReservationRequest{
		EventID:    101,
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: benchResponseCh,
		Ctx:        benchCtx,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		select {
		case q <- req:
		default:
			<-q
			q <- req
		}
	}
}
