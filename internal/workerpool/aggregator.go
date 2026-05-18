package workerpool

import (
	"context"
	"runtime"
	"time"

	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

const (
	aggregatorBatchSize = 128
	aggregatorQueueSize = 1 << 18

	aggregatorLaneCount = 16
)

type reserveJob struct {
	Request       ReservationRequest
	ReservationID uint64
	ShardID       int
}

type Aggregator struct {
	store *inventory.Store

	queue chan reserveJob
}

func NewAggregator(store *inventory.Store) *Aggregator {

	return &Aggregator{
		store: store,
		queue: make(chan reserveJob, aggregatorQueueSize),
	}
}

func (a *Aggregator) Submit(job reserveJob) bool {

	select {

	case a.queue <- job:
		return true

	default:
		return false
	}
}

func (a *Aggregator) Run(ctx context.Context) {

	var jobs [aggregatorBatchSize]reserveJob

	var reserveBatch [aggregatorBatchSize]struct {
		EventID       uint64
		Quantity      uint32
		ReservationID uint64
		ShardID       int
	}

	var keyBuf [256]byte
	buf := keyBuf[:0]

	for {

		select {

		case <-ctx.Done():
			return

		default:
		}

		n := 0

	collectLoop:

		for n < aggregatorBatchSize {

			select {

			case job := <-a.queue:

				jobs[n] = job

				reserveBatch[n] = struct {
					EventID       uint64
					Quantity      uint32
					ReservationID uint64
					ShardID       int
				}{
					EventID:       job.Request.EventID,
					Quantity:      job.Request.Quantity,
					ReservationID: job.ReservationID,
					ShardID:       job.ShardID,
				}

				n++

			default:
				break collectLoop
			}
		}

		telemetry.GlobalRuntimeStats.RecordBatch(uint64(n))

		if n == 0 {

			runtime.Gosched()

			continue
		}

		errs := a.store.ReserveBatch(
			context.Background(),
			reserveBatch[:n],
			buf,
		)

		now := time.Now().UnixNano()

		for i := 0; i < n; i++ {

			latency := uint64(time.Now().UnixNano() - jobs[i].Request.TimestampNs)

			telemetry.GlobalRuntimeStats.RecordLatency(latency)

			jobs[i].Request.ResponseCh <- ReservationResult{
				ReservationID: jobs[i].ReservationID,
				TimestampNs:   now,
				Err:           errs[i],
			}
		}
	}
}

var AggregatorLanes [aggregatorLaneCount]*Aggregator
