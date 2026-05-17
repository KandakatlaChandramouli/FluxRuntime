package workerpool

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

const (
	pressureRedisAddr = "localhost:6379"

	pressureEventID   = uint64(20001)
	pressureInventory = 10_000_000

	pressureResponseQ = 8192
)

var pressureCtx = context.Background()

func newPressureStore(b *testing.B) *inventory.Store {
	b.Helper()

	store, err := inventory.NewStore(
		pressureCtx,
		pressureRedisAddr,
		"",
		0,
	)
	if err != nil {
		b.Fatalf("inventory.NewStore: %v", err)
	}

	return store
}

func newPressureMetrics() *telemetry.Metrics {
	return telemetry.NewMetrics(prometheus.NewRegistry())
}

func newPressureDispatcher(
	b *testing.B,
	store *inventory.Store,
	metrics *telemetry.Metrics,
	workerCount int,
) *Dispatcher {

	b.Helper()

	return NewDispatcher(
		workerCount,
		store,
		metrics,
	)
}

func BenchmarkPressureHotKey(b *testing.B) {

	store := newPressureStore(b)
	defer store.Close()

	metrics := newPressureMetrics()

	dispatcher := newPressureDispatcher(
		b,
		store,
		metrics,
		8,
	)
	defer dispatcher.Close()

	responseCh := make(chan ReservationResult, pressureResponseQ)

	var dispatched atomic.Int64
	var rejected atomic.Int64
	var completed atomic.Int64

	var latencyMu sync.Mutex
	latencies := make([]int64, 0, 1_000_000)

	go func() {
		for res := range responseCh {
			completed.Add(1)

			latency := time.Now().UnixNano() - res.TimestampNs

			latencyMu.Lock()
			latencies = append(latencies, latency)
			latencyMu.Unlock()
		}
	}()

	senderCount := runtime.GOMAXPROCS(0)

	if senderCount < 2 {
		senderCount = 2
	}

	workCh := make(chan int, senderCount*8)

	var senderWG sync.WaitGroup

	senderWG.Add(senderCount)

	for i := 0; i < senderCount; i++ {

		go func() {

			defer senderWG.Done()

			for range workCh {

				req := ReservationRequest{
					EventID:    pressureEventID,
					Quantity:   1,
					ResponseCh: responseCh,
				}

				err := dispatcher.Dispatch(req)

				if err != nil {
					rejected.Add(1)
					continue
				}

				dispatched.Add(1)
			}
		}()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		workCh <- i
	}

	close(workCh)

	senderWG.Wait()

	b.StopTimer()

	var totalIngested uint64
	var totalRejected uint64
	var maxWatermark uint64

	for _, shard := range dispatcher.shards {

		snap := shard.Snapshot()

		totalIngested += snap.TotalIngested
		totalRejected += snap.TotalRejected

		if snap.QueueHighWatermark > maxWatermark {
			maxWatermark = snap.QueueHighWatermark
		}
	}

	f, err := os.Create("benchmarks/results/latencies.csv")
	if err == nil {
		defer f.Close()

		for _, v := range latencies {
			fmt.Fprintf(f, "%d\n", v)
		}
	}

	b.ReportMetric(float64(dispatched.Load()), "dispatched")
	b.ReportMetric(float64(rejected.Load()), "rejected")
	b.ReportMetric(float64(completed.Load()), "completed")
	b.ReportMetric(float64(totalIngested), "shard_ingested")
	b.ReportMetric(float64(totalRejected), "shard_rejected")
	b.ReportMetric(float64(maxWatermark), "max_watermark")
}

func BenchmarkPressureMatrix(b *testing.B) {

	workerConfigs := []int{
		2,
		4,
		8,
		16,
	}

	for _, workers := range workerConfigs {

		b.Run(fmt.Sprintf("workers_%d", workers), func(b *testing.B) {

			store := newPressureStore(b)
			defer store.Close()

			metrics := newPressureMetrics()

			dispatcher := newPressureDispatcher(
				b,
				store,
				metrics,
				workers,
			)
			defer dispatcher.Close()

			responseCh := make(chan ReservationResult, pressureResponseQ)

			var dispatched atomic.Int64
			var rejected atomic.Int64
			var completed atomic.Int64

			go func() {
				for range responseCh {
					completed.Add(1)
				}
			}()

			senderCount := runtime.GOMAXPROCS(0)

			if senderCount < 2 {
				senderCount = 2
			}

			workCh := make(chan int, senderCount*8)

			var senderWG sync.WaitGroup

			senderWG.Add(senderCount)

			for i := 0; i < senderCount; i++ {

				go func() {

					defer senderWG.Done()

					for range workCh {

						req := ReservationRequest{
							EventID:    pressureEventID,
							Quantity:   1,
							ResponseCh: responseCh,
						}

						err := dispatcher.Dispatch(req)

						if err != nil {
							rejected.Add(1)
							continue
						}

						dispatched.Add(1)
					}
				}()
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				workCh <- i
			}

			close(workCh)

			senderWG.Wait()

			b.StopTimer()

			b.ReportMetric(
				float64(dispatched.Load()),
				"dispatched",
			)

			b.ReportMetric(
				float64(rejected.Load()),
				"rejected",
			)

			b.ReportMetric(
				float64(completed.Load()),
				"completed",
			)
		})
	}
}
