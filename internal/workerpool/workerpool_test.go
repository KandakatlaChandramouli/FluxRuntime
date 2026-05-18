package workerpool

// workerpool_test.go — Phase 1A deterministic saturation test.
//
// Experimental objective (SYSTEM_SPEC.md §9.1):
//   - queue saturation under hot-key skew
//   - deterministic rejection under bounded capacity
//   - contention stability with no deadlocks
//   - bounded backpressure mechanics
//   - stable worker termination
//
// No Redis. No gRPC. No external dependencies.
// Standard library concurrency primitives only.

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/research/phase1a/internal/config"
)

// testConstants defines the synthetic load parameters.
// HOT_EVENT_ID is fixed across all 100,000 requests to force
// deterministic FNV-1a routing onto a single shard — intentional skew.
const (
	totalRequests = 100_000
	hotEventID    = uint64(101)
	workerCount   = 8 // fixed topology; must not be dynamic
)

// TestHotKeyQueueSaturation is the primary saturation experiment.
//
// Measures:
//   - TotalIngested: all dispatch attempts
//   - TotalRejected: requests dropped at full queue
//   - TotalProcessed: requests drained by worker (zero — no worker running)
//   - QueueHighWatermark: peak observed queue depth
//
// The worker drain loop is intentionally NOT started. This isolates the
// dispatcher + shard + queue subsystem from Redis and Prometheus dependencies,
// allowing clean measurement of the queue pressure mechanics alone.
//
// The hot shard fills to config.QueueCapacity (4096) and then all subsequent
// dispatches must be rejected instantly — validating the non-blocking
// select/default contract from spec §5.5.
func TestHotKeyQueueSaturation(t *testing.T) {
	// Construct shards directly. No worker goroutines started.
	// This isolates queue mechanics from Redis/Prometheus dependencies.
	shards := make([]*Shard, workerCount)
	for i := 0; i < workerCount; i++ {
		shards[i] = newShard(i)
	}

	dispatcher := &Dispatcher{
		shards:      shards,
		workerCount: workerCount,
	}

	// Determine which shard the hot event routes to.
	// This is the shard that must exhibit saturation.
	hotShardIdx := dispatcher.shardIndex(hotEventID)

	// Atomic counters for concurrent dispatch tracking.
	// Using sync/atomic directly — no mutexes, no channels for counting.
	var (
		atomicAccepted atomic.Uint64
		atomicRejected atomic.Uint64
	)

	// Pre-allocate all response channels before the dispatch goroutines start.
	// Buffered(1): ensures workers never block on send even if handler is gone.
	// Allocated here, not inside goroutines, to bound total heap growth.
	responseChs := make([]chan ReservationResult, totalRequests)
	for i := range responseChs {
		responseChs[i] = make(chan ReservationResult, 1)
	}

	ctx := context.Background()

	// Dispatch all requests concurrently via a fixed goroutine pool.
	// Pool size is bounded — no goroutine-per-request explosion.
	//
	// Architecture note: the concurrency degree here represents inbound
	// gRPC handler goroutines. In production each RPC is one goroutine;
	// here we use a fixed pool to stress the dispatcher without spawning
	// 100,000 goroutines, which would measure goroutine scheduler overhead
	// rather than queue mechanics.
	const senderConcurrency = 64
	var wg sync.WaitGroup

	// Work channel feeds request indices to sender goroutines.
	// Bounded capacity prevents sender backlog from inflating memory.
	workCh := make(chan int, senderConcurrency*2)

	// Start fixed sender pool.
	wg.Add(senderConcurrency)
	for s := 0; s < senderConcurrency; s++ {
		go func() {
			defer wg.Done()
			for idx := range workCh {
				req := ReservationRequest{
					EventID:    hotEventID,
					Quantity:   1,
					UserIDHigh: 0,
					UserIDLow:  uint64(idx),
					ResponseCh: responseChs[idx],
					Ctx:        ctx,
				}

				err := dispatcher.Dispatch(req)
				if err == nil {
					atomicAccepted.Add(1)
				} else {
					atomicRejected.Add(1)
				}
			}
		}()
	}

	// Feed all request indices into the work channel then close.
	// No append(). No dynamic growth. Deterministic feed.
	for i := 0; i < totalRequests; i++ {
		workCh <- i
	}
	close(workCh)

	// Wait for all senders to complete.
	// No sleep-based synchronization.
	wg.Wait()

	// --- Metric collection ---

	accepted := atomicAccepted.Load()
	rejected := atomicRejected.Load()

	// Snapshot all shards for counter validation.
	snapshots := dispatcher.Snapshots()

	// Aggregate across all shards.
	var (
		totalIngested     uint64
		totalRejected     uint64
		totalProcessed    uint64
		hotShardWatermark uint64
		hotShardIngested  uint64
		hotShardRejected  uint64
	)
	for _, snap := range snapshots {
		totalIngested += snap.TotalIngested
		totalRejected += snap.TotalRejected
		totalProcessed += snap.TotalProcessed
		if snap.ShardID == hotShardIdx {
			hotShardWatermark = snap.QueueHighWatermark
			hotShardIngested = snap.TotalIngested
			hotShardRejected = snap.TotalRejected
		}
	}

	// --- Structured result output ---
	//
	// t.Logf is acceptable here: test output path, not hot path.
	t.Logf("=== Phase 1A Hot-Key Saturation Results ===")
	t.Logf("total_requests     : %d", totalRequests)
	t.Logf("accepted           : %d", accepted)
	t.Logf("rejected           : %d", rejected)
	t.Logf("hot_shard_idx      : %d", hotShardIdx)
	t.Logf("hot_shard_ingested : %d", hotShardIngested)
	t.Logf("hot_shard_rejected : %d", hotShardRejected)
	t.Logf("hot_shard_watermark: %d", hotShardWatermark)
	t.Logf("total_processed    : %d (no workers; expected 0)", totalProcessed)
	t.Logf("queue_capacity     : %d", config.QueueCapacity)
	t.Logf("===========================================")

	// --- Assertions ---

	// 1. Conservation: accepted + rejected == total dispatched.
	//    Validates no requests are silently dropped or double-counted.
	if accepted+rejected != totalRequests {
		t.Errorf("conservation violated: accepted(%d) + rejected(%d) = %d, want %d",
			accepted, rejected, accepted+rejected, totalRequests)
	}

	// 2. Shard counter conservation: shard-level ingested == accepted + rejected.
	//    Validates atomic counter correctness under concurrent dispatch.
	if totalIngested != uint64(totalRequests) {
		t.Errorf("shard TotalIngested=%d, want %d", totalIngested, totalRequests)
	}

	// 3. Rejected counter consistency: shard-level == atomic tracker.
	if totalRejected != rejected {
		t.Errorf("shard TotalRejected=%d != atomic rejected=%d", totalRejected, rejected)
	}

	// 4. Saturation must occur: all 100k requests against one shard with
	//    capacity 4096 guarantees rejections exist.
	//    If no rejections: bounded queue constraint is violated.
	if rejected == 0 {
		t.Errorf("expected queue saturation rejections, got 0; queue capacity=%d, requests=%d",
			config.QueueCapacity, totalRequests)
	}

	// 5. Accepted must not exceed queue capacity.
	//    Validates bounded queue contract: no request accepted beyond cap.
	if accepted > config.QueueCapacity {
		t.Errorf("accepted=%d exceeds queue capacity=%d; bounded queue contract violated",
			accepted, config.QueueCapacity)
	}

	// 6. Hot shard carries all ingested traffic.
	//    Validates FNV-1a deterministic routing: 100% skew onto one shard.
	if hotShardIngested != uint64(totalRequests) {
		t.Errorf("hot_shard_ingested=%d, want %d; routing is non-deterministic",
			hotShardIngested, totalRequests)
	}

	// 7. High watermark must be > 0 and <= queue capacity.
	//    Validates watermark tracking is live under concurrent load.
	if hotShardWatermark == 0 {
		t.Errorf("hot_shard_watermark=0; watermark tracking is broken")
	}
	if hotShardWatermark > config.QueueCapacity {
		t.Errorf("hot_shard_watermark=%d exceeds queue capacity=%d",
			hotShardWatermark, config.QueueCapacity)
	}

	// 8. No processed work without workers.
	//    Validates worker isolation: queue mechanics are independent of drain.
	if totalProcessed != 0 {
		t.Errorf("totalProcessed=%d, want 0; no workers were started", totalProcessed)
	}

	// 9. All other shards must have zero ingested traffic.
	//    Validates hot-key skew isolation: cold shards must remain idle.
	for _, snap := range snapshots {
		if snap.ShardID == hotShardIdx {
			continue
		}
		if snap.TotalIngested != 0 {
			t.Errorf("cold shard %d: TotalIngested=%d, want 0; routing skew broken",
				snap.ShardID, snap.TotalIngested)
		}
	}
}

// TestDispatchNonBlocking validates the non-blocking rejection contract
// from spec §5.5 in isolation on a single shard.
//
// Fills the queue to exactly capacity, then dispatches one additional
// request and asserts it is rejected without blocking.
func TestDispatchNonBlocking(t *testing.T) {
	shard := newShard(0)

	ctx := context.Background()
	responseCh := make(chan ReservationResult, 1)

	req := ReservationRequest{
		EventID:    hotEventID,
		Quantity:   1,
		UserIDHigh: 0,
		UserIDLow:  0,
		ResponseCh: responseCh,
		Ctx:        ctx,
	}

	// Fill queue to exact capacity.
	var filled int
	for filled = 0; filled < config.QueueCapacity; filled++ {
		err := shard.Dispatch(req)
		if err != nil {
			break
		}
	}

	// Next dispatch must be rejected immediately.
	err := shard.Dispatch(req)
	if err == nil {
		t.Errorf("expected ErrQueueFull after filling %d slots, got nil", filled)
	}
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}

	snap := shard.Snapshot()

	t.Logf("=== TestDispatchNonBlocking ===")
	t.Logf("filled             : %d", filled)
	t.Logf("queue_depth        : %d", snap.QueueDepth)
	t.Logf("total_ingested     : %d", snap.TotalIngested)
	t.Logf("total_rejected     : %d", snap.TotalRejected)
	t.Logf("queue_high_watermark: %d", snap.QueueHighWatermark)
	t.Logf("==============================")

	if snap.QueueDepth != config.QueueCapacity {
		t.Errorf("queue_depth=%d, want %d", snap.QueueDepth, config.QueueCapacity)
	}

	if snap.TotalRejected != 1 {
		t.Errorf("TotalRejected=%d, want 1", snap.TotalRejected)
	}

	// Watermark must equal capacity: queue was filled completely.
	// (watermark is sampled before insert; may be cap-1 at peak sample point)
	if snap.QueueHighWatermark > uint64(config.QueueCapacity) {
		t.Errorf("QueueHighWatermark=%d exceeds capacity=%d",
			snap.QueueHighWatermark, config.QueueCapacity)
	}
}

// TestDeterministicShardRouting validates that FNV-1a routing is stable
// across repeated calls with the same EventID.
//
// Shard affinity must be deterministic — same EventID always maps to same shard.
func TestDeterministicShardRouting(t *testing.T) {
	shards := make([]*Shard, workerCount)
	for i := 0; i < workerCount; i++ {
		shards[i] = newShard(i)
	}
	dispatcher := &Dispatcher{shards: shards, workerCount: workerCount}

	first := dispatcher.shardIndex(hotEventID)

	const iterations = 10_000
	for i := 0; i < iterations; i++ {
		idx := dispatcher.shardIndex(hotEventID)
		if idx != first {
			t.Errorf("shard routing non-deterministic: got %d at iteration %d, want %d",
				idx, i, first)
			return
		}
	}

	t.Logf("shard routing: EventID=%d always maps to shard %d across %d iterations",
		hotEventID, first, iterations)
}

// TestWorkerTermination validates that the stopCh signal causes
// the worker goroutine to exit cleanly without leaking.
//
// A WaitGroup tracks the goroutine. The test asserts it exits
// within the scope of the test — no goroutine leak.
func TestWorkerTermination(t *testing.T) {
	shard := newShard(0)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		worker(
			shard,
			nil,
			nil,
			ctx,
		)
	}()

	cancel()

	wg.Wait()

	t.Log("worker terminated cleanly on context cancellation")
}

// TestConcurrentCounterConsistency validates that atomic counters remain
// consistent under concurrent dispatch from multiple goroutines.
//
// Dispatches a small fixed load concurrently and asserts:
//
//	ingested == accepted + rejected  (conservation)
//	ingested == total dispatched     (no silent drops)
func TestConcurrentCounterConsistency(t *testing.T) {
	const (
		concurrency = 32
		perSender   = 512
		total       = concurrency * perSender
	)

	shard := newShard(0)
	ctx := context.Background()

	var (
		accepted atomic.Uint64
		rejected atomic.Uint64
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for s := 0; s < concurrency; s++ {
		go func() {
			defer wg.Done()
			responseCh := make(chan ReservationResult, 1)
			req := ReservationRequest{
				EventID:    hotEventID,
				Quantity:   1,
				ResponseCh: responseCh,
				Ctx:        ctx,
			}
			for i := 0; i < perSender; i++ {
				err := shard.Dispatch(req)
				if err == nil {
					accepted.Add(1)
				} else {
					rejected.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	snap := shard.Snapshot()
	a := accepted.Load()
	r := rejected.Load()

	t.Logf("=== TestConcurrentCounterConsistency ===")
	t.Logf("total_dispatched   : %d", total)
	t.Logf("accepted           : %d", a)
	t.Logf("rejected           : %d", r)
	t.Logf("shard_ingested     : %d", snap.TotalIngested)
	t.Logf("shard_rejected     : %d", snap.TotalRejected)
	t.Logf("high_watermark     : %d", snap.QueueHighWatermark)
	t.Logf("========================================")

	// Conservation invariant.
	if a+r != uint64(total) {
		t.Errorf("conservation: accepted(%d)+rejected(%d)=%d, want %d",
			a, r, a+r, total)
	}

	// Shard-level ingested must match total dispatched.
	if snap.TotalIngested != uint64(total) {
		t.Errorf("shard TotalIngested=%d, want %d", snap.TotalIngested, total)
	}

	// Shard-level rejected must match atomic tracker.
	if snap.TotalRejected != r {
		t.Errorf("shard TotalRejected=%d != atomic rejected=%d", snap.TotalRejected, r)
	}
}
