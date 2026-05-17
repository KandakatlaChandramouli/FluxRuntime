package workerpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

// liveBenchWorkerCount mirrors production formula.
const liveBenchWorkerCount = 8

// liveBenchQueueDepth matches config.QueueCapacity.
const liveBenchQueueDepth = 4096

// liveBenchRedisAddr is the local Redis instance assumed by Phase 1C.
const liveBenchRedisAddr = "localhost:6379"

// hotEventID is the fixed EventID used for single-shard collapse experiments.
const liveBenchHotEventID uint64 = 101

// liveBenchInventory is the pre-seeded ticket count per event.
// Large enough to survive a full benchmark run without exhaustion
// under normal b.N scaling.
const liveBenchInventory = 10_000_000

// liveBenchCtx is a shared background context for all benchmark requests.
var liveBenchCtx = context.Background()

// newLiveStore constructs a real inventory.Store connected to local Redis.
// Fails the benchmark immediately if Redis is unreachable.
func newLiveStore(b *testing.B) *inventory.Store {
	b.Helper()
	store, err := inventory.NewStore(liveBenchCtx, liveBenchRedisAddr, "", 0)
	if err != nil {
		b.Fatalf("inventory.NewStore: %v", err)
	}
	return store
}

// newLiveMetrics constructs a non-registering Metrics instance for benchmarks.
// Uses a fresh prometheus registry per call to avoid duplicate registration.
func newLiveMetrics() *telemetry.Metrics {
	return telemetry.NewMetrics(newBenchRegistry())
}

// newBenchRegistry returns a new isolated prometheus registry.
// Avoids "duplicate metrics collector" panics across benchmark runs.
func newBenchRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

// newLiveDispatcher constructs a real Dispatcher with live worker goroutines.
func newLiveDispatcher(
	b *testing.B,
	store *inventory.Store,
	metrics *telemetry.Metrics,
) *Dispatcher {
	b.Helper()

	d := NewDispatcher(
		liveBenchWorkerCount,
		store,
		metrics,
	)

	return d
}

// seedInventory writes a ticket inventory counter into Redis for the given eventID.
// Uses SET to unconditionally overwrite any prior state.
func seedInventory(b *testing.B, client *redis.Client, eventID uint64, count int64) {
	b.Helper()
	var buf [64]byte
	key := ticketsKeyBench(eventID, buf[:0])
	if err := client.Set(liveBenchCtx, key, count, 0).Err(); err != nil {
		b.Fatalf("seedInventory eventID=%d: %v", eventID, err)
	}
}

// cleanInventory deletes the tickets and orders keys for the given eventID.
func cleanInventory(b *testing.B, client *redis.Client, eventID uint64) {
	b.Helper()
	var buf [64]byte
	tk := ticketsKeyBench(eventID, buf[:0])
	ok := ordersKeyBench(eventID, buf[:0])
	client.Del(liveBenchCtx, tk, ok)
}

// ticketsKeyBench mirrors the keyspace logic for benchmark setup only.
// Not on the hot path.
func ticketsKeyBench(eventID uint64, buf []byte) string {
	buf = buf[:0]
	buf = append(buf, '{', 'e', 'v', 'e', 'n', 't', ':')
	buf = appendUintBench(buf, eventID)
	buf = append(buf, '}', ':', 't', 'i', 'c', 'k', 'e', 't', 's')
	return string(buf)
}

// ordersKeyBench mirrors the keyspace logic for benchmark setup only.
func ordersKeyBench(eventID uint64, buf []byte) string {
	buf = buf[:0]
	buf = append(buf, '{', 'e', 'v', 'e', 'n', 't', ':')
	buf = appendUintBench(buf, eventID)
	buf = append(buf, '}', ':', 'o', 'r', 'd', 'e', 'r', 's')
	return string(buf)
}

// appendUintBench appends the decimal representation of v to dst.
func appendUintBench(dst []byte, v uint64) []byte {
	if v == 0 {
		return append(dst, '0')
	}
	var tmp [20]byte
	pos := len(tmp)
	for v > 0 {
		pos--
		tmp[pos] = byte('0' + v%10)
		v /= 10
	}
	return append(dst, tmp[pos:]...)
}

// rawRedisClient returns a bare redis.Client for benchmark seed/cleanup.
// Not used on the hot path.
func rawRedisClient(b *testing.B) *redis.Client {
	b.Helper()
	c := redis.NewClient(&redis.Options{
		Addr: liveBenchRedisAddr,
		DB:   0,
	})
	if err := c.Ping(liveBenchCtx).Err(); err != nil {
		b.Fatalf("Redis unreachable at %s: %v", liveBenchRedisAddr, err)
	}
	return c
}

// newResponseCh returns a pre-allocated buffered response channel.
// Capacity 1: non-blocking worker send; handler reads after Dispatch returns.
func newResponseCh() chan ReservationResult {
	return make(chan ReservationResult, 1)
}

// drainResponses consumes exactly n results from responseCh.
// Blocks until all n results arrive or the deadline fires.
// Returns false if deadline exceeded (benchmark failure condition).
func drainResponses(responseCh chan ReservationResult, n int, deadline time.Duration) bool {
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	for i := 0; i < n; i++ {
		select {
		case <-responseCh:
		case <-timer.C:
			return false
		}
	}
	return true
}

// ============================================================
// BenchmarkLiveReservationHotKey
// ============================================================

// BenchmarkLiveReservationHotKey collapses all traffic onto a single shard
// via a fixed EventID. Exercises the full execution path:
//
//	Dispatcher -> Queue -> Worker -> Redis Lua -> ResponseCh
//
// Measures hot-shard contention, queue pressure, and execution latency
// amplification under single-shard skew.
func BenchmarkLiveReservationHotKey(b *testing.B) {
	client := rawRedisClient(b)
	defer client.Close()

	seedInventory(b, client, liveBenchHotEventID, liveBenchInventory)
	defer cleanInventory(b, client, liveBenchHotEventID)

	store := newLiveStore(b)
	defer store.Close()

	metrics := newLiveMetrics()

	d := newLiveDispatcher(b, store, metrics)
	defer d.Close()
	// Pre-allocate a pool of response channels.
	// One channel per concurrent in-flight request slot.
	// Bounded pool prevents goroutine-per-request explosion.
	const slots = 256
	chs := make([]chan ReservationResult, slots)
	for i := range chs {
		chs[i] = newResponseCh()
	}

	var dispatched atomic.Int64
	var rejected atomic.Int64

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ch := chs[i&(slots-1)]

		// Drain any prior result non-blocking before reuse.
		select {
		case <-ch:
		default:
		}

		req := ReservationRequest{
			EventID:    liveBenchHotEventID,
			Quantity:   1,
			UserIDHigh: 0,
			UserIDLow:  uint64(i),
			ResponseCh: ch,
			Ctx:        liveBenchCtx,
		}

		err := d.Dispatch(req)
		if err == nil {
			dispatched.Add(1)
			// Consume response to prevent channel backpressure.
			<-ch
		} else {
			rejected.Add(1)
		}
	}

	b.ReportMetric(float64(dispatched.Load()), "dispatched")
	b.ReportMetric(float64(rejected.Load()), "rejected")
}

// ============================================================
// BenchmarkLiveReservationColdKeys
// ============================================================

// BenchmarkLiveReservationColdKeys distributes traffic across 256 distinct
// EventIDs, spreading load across all shards. Measures balanced execution
// throughput and reduced shard contention relative to BenchmarkLiveReservationHotKey.
func BenchmarkLiveReservationColdKeys(b *testing.B) {
	client := rawRedisClient(b)
	defer client.Close()

	// 256 distinct event IDs — spread across all 8 shards.
	const numEvents = 256
	var coldIDs [numEvents]uint64
	for i := range coldIDs {
		coldIDs[i] = uint64(i + 1000)
	}

	for _, id := range coldIDs {
		seedInventory(b, client, id, liveBenchInventory)
	}
	defer func() {
		for _, id := range coldIDs {
			cleanInventory(b, client, id)
		}
	}()

	store := newLiveStore(b)
	defer store.Close()

	metrics := newLiveMetrics()

	stopCh := make(chan struct{})
	defer close(stopCh)

	d := newLiveDispatcher(b, store, metrics)

	const slots = 256
	chs := make([]chan ReservationResult, slots)
	for i := range chs {
		chs[i] = newResponseCh()
	}

	var dispatched atomic.Int64
	var rejected atomic.Int64

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ch := chs[i&(slots-1)]

		select {
		case <-ch:
		default:
		}

		req := ReservationRequest{
			EventID:    coldIDs[i&(numEvents-1)],
			Quantity:   1,
			UserIDHigh: 0,
			UserIDLow:  uint64(i),
			ResponseCh: ch,
			Ctx:        liveBenchCtx,
		}

		err := d.Dispatch(req)
		if err == nil {
			dispatched.Add(1)
			<-ch
		} else {
			rejected.Add(1)
		}
	}

	b.ReportMetric(float64(dispatched.Load()), "dispatched")
	b.ReportMetric(float64(rejected.Load()), "rejected")
}

// ============================================================
// BenchmarkRedisReserveOnly
// ============================================================

// BenchmarkRedisReserveOnly isolates Store.Reserve() execution cost.
// Bypasses Dispatcher, queue, and workerpool entirely.
// Measures raw Redis Lua round-trip latency, serialization overhead,
// and allocation behavior of the inventory coordination layer alone.
func BenchmarkRedisReserveOnly(b *testing.B) {
	client := rawRedisClient(b)
	defer client.Close()

	const benchEventID uint64 = 9999901
	seedInventory(b, client, benchEventID, liveBenchInventory)
	defer cleanInventory(b, client, benchEventID)

	store := newLiveStore(b)
	defer store.Close()

	// Worker-owned scratch buffer — reused across iterations.
	// Mirrors exact worker.go allocation pattern.
	var keyBuf [128]byte
	buf := keyBuf[:0]

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reservationID := inventory.NextReservationID()

		_ = store.Reserve(
			liveBenchCtx,
			benchEventID,
			1,
			reservationID,
			0,
			buf,
		)
	}
}

// ============================================================
// BenchmarkWorkerDrainExecution
// ============================================================

// BenchmarkWorkerDrainExecution measures sustained full-path execution
// throughput: dequeue + Redis Lua execution + response send.
//
// Uses a fixed sender concurrency pool to continuously feed the queue
// while live worker goroutines drain and execute. Measures drain
// throughput limits, execution bottlenecks, and contention amplification
// under sustained load.
func BenchmarkWorkerDrainExecution(b *testing.B) {
	client := rawRedisClient(b)
	defer client.Close()

	const drainEventID uint64 = 7777701
	seedInventory(b, client, drainEventID, liveBenchInventory)
	defer cleanInventory(b, client, drainEventID)

	store := newLiveStore(b)
	defer store.Close()

	metrics := newLiveMetrics()

	d := newLiveDispatcher(b, store, metrics)

	// Sender pool: fixed concurrency, no goroutine-per-request.
	// senderCount matches GOMAXPROCS to saturate dispatch without
	// scheduler amplification.
	senderCount := runtime.GOMAXPROCS(0)
	if senderCount < 2 {
		senderCount = 2
	}

	// workCh feeds request indices to sender goroutines.
	// Bounded capacity prevents backlog inflation.
	workCh := make(chan int, senderCount*2)

	// Pre-allocate response channels — one per sender goroutine.
	// Each sender owns its channel and waits for its response before
	// sending the next request. This bounds in-flight requests to
	// senderCount, preventing queue explosion.
	chs := make([]chan ReservationResult, senderCount)
	for i := range chs {
		chs[i] = newResponseCh()
	}

	var executed atomic.Int64
	var rejected atomic.Int64

	var wg sync.WaitGroup
	wg.Add(senderCount)

	for s := 0; s < senderCount; s++ {
		go func(senderID int) {
			defer wg.Done()
			ch := chs[senderID]
			for idx := range workCh {
				req := ReservationRequest{
					EventID:    drainEventID,
					Quantity:   1,
					UserIDHigh: 0,
					UserIDLow:  uint64(idx),
					ResponseCh: ch,
					Ctx:        liveBenchCtx,
				}
				err := d.Dispatch(req)
				if err == nil {
					<-ch
					executed.Add(1)
				} else {
					rejected.Add(1)
				}
			}
		}(s)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		workCh <- i
	}
	close(workCh)

	wg.Wait()

	b.ReportMetric(float64(executed.Load()), "executed")
	b.ReportMetric(float64(rejected.Load()), "rejected")
}
