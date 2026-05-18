package workerpool

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/research/phase1a/internal/telemetry"

	"github.com/research/phase1a/internal/config"
	"github.com/research/phase1a/internal/inventory"
)

func TestRuntimeSoak(t *testing.T) {

	store, err := inventory.NewStore(
		context.Background(),
		"localhost:6379",
		"",
		0,
	)

	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	defer store.Close()

	metrics := newPressureMetrics()

	dispatcher := NewDispatcher(
		8,
		store,
		metrics,
	)

	defer dispatcher.Close()

	duration := 60 * time.Second

	ctx, cancel := context.WithTimeout(
		context.Background(),
		duration,
	)

	defer cancel()

	responseCh := make(chan ReservationResult, 1<<20)

	var dispatched atomic.Uint64
	var rejected atomic.Uint64
	var completed atomic.Uint64

	senderCount := runtime.GOMAXPROCS(0) * 2

	var wg sync.WaitGroup

	wg.Add(senderCount)

	start := time.Now()

	for i := 0; i < senderCount; i++ {

		go func(id int) {

			defer wg.Done()

			eventID := uint64(id % 16)

			for {

				select {

				case <-ctx.Done():
					return

				default:
				}

				req := ReservationRequest{
					EventID:     eventID,
					Quantity:    1,
					UserIDHigh:  uint64(id),
					UserIDLow:   uint64(time.Now().UnixNano()),
					ResponseCh:  responseCh,
					Ctx:         context.Background(),
					TimestampNs: time.Now().UnixNano(),
				}

				err := dispatcher.Dispatch(req)

				dispatched.Add(1)

				if err != nil {
					rejected.Add(1)
				}
			}
		}(i)
	}

	go func() {

		for range responseCh {
			completed.Add(1)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {

		select {

		case <-ctx.Done():

			wg.Wait()

			totalDispatched := dispatched.Load()
			totalRejected := rejected.Load()
			totalCompleted := completed.Load()

			elapsed := time.Since(start).Seconds()

			t.Log("========================================")
			t.Log("Runtime Soak Test Results")
			t.Log("========================================")

			t.Logf("duration_sec      : %.2f", elapsed)
			t.Logf("total_dispatched  : %d", totalDispatched)
			t.Logf("total_completed   : %d", totalCompleted)
			t.Logf("total_rejected    : %d", totalRejected)

			t.Logf(
				"throughput_reqsec : %.2f",
				float64(totalDispatched)/elapsed,
			)

			t.Logf(
				"rejection_rate    : %.4f",
				float64(totalRejected)/float64(totalDispatched),
			)

			t.Logf(
				"queue_capacity    : %d",
				config.QueueCapacity,
			)

			t.Log("========================================")

			return

		case <-ticker.C:

			var totalDepth int

			for _, shard := range dispatcher.Snapshots() {
				totalDepth += shard.QueueDepth
			}

			t.Logf(
				"[LIVE] dispatched=%d completed=%d rejected=%d queue_depth=%d",
				dispatched.Load(),
				completed.Load(),
				rejected.Load(),
				totalDepth,
			)

			stats := telemetry.GlobalRuntimeStats.Snapshot()

			t.Logf(
				"[LATENCY] p50=%dns p95=%dns p99=%dns avg=%dns max=%dns avg_batch=%.2f",
				stats.P50,
				stats.P95,
				stats.P99,
				stats.AvgLatencyNs,
				stats.MaxLatencyNs,
				stats.AvgBatchSize,
			)
		}
	}
}
