package workerpool

import (
	"context"
	"runtime"
	"sync"

	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

// ErrQueueFull is re-exported here so dispatcher.go has a single import point.
var ErrQueueFull = inventory.ErrQueueFull

const (
	workerBatchSize = 64
)

var aggregatorOnce sync.Once

func worker(
	shard *Shard,
	store *inventory.Store,
	metrics *telemetry.Metrics,
	ctx context.Context,
) {

	_ = metrics

	aggregatorOnce.Do(func() {

		for i := 0; i < aggregatorLaneCount; i++ {

			AggregatorLanes[i] = NewAggregator(store)

			go AggregatorLanes[i].Run(ctx)
		}
	})

	var batch [workerBatchSize]ReservationRequest

	for {

		select {

		case <-ctx.Done():
			return

		default:
		}

		n := shard.queue.DequeueBatch(
			workerBatchSize,
			batch[:],
		)

		if n == 0 {

			runtime.Gosched()

			continue
		}

		for i := 0; i < n; i++ {

			req := batch[i]

			lane := int(req.EventID % aggregatorLaneCount)

			ok := AggregatorLanes[lane].Submit(
				reserveJob{
					Request:       req,
					ReservationID: inventory.NextReservationID(),
					ShardID:       shard.ID,
				},
			)

			if !ok {
				continue
			}

			shard.Counters.TotalProcessed.Add(1)
		}
	}
}
