package workerpool

import (
	"context"
	"runtime"
	"time"

	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

// ErrQueueFull is re-exported here so dispatcher.go has a single import point.
var ErrQueueFull = inventory.ErrQueueFull

const (
	workerBatchSize = 64
	maxSpinCount    = 256
)

func worker(
	shard *Shard,
	store *inventory.Store,
	metrics *telemetry.Metrics,
	ctx context.Context,
) {

	var keyBuf [128]byte
	buf := keyBuf[:0]

	var batch [workerBatchSize]ReservationRequest

	var reserveBatch [workerBatchSize]struct {
		EventID       uint64
		Quantity      uint32
		ReservationID uint64
		ShardID       int
	}

	_ = metrics

	spinCount := 0

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

			if store == nil {
				return
			}

			spinCount++

			if spinCount < maxSpinCount {

				runtime.Gosched()

			} else {

				time.Sleep(time.Microsecond)

				spinCount = 0
			}

			continue
		}

		spinCount = 0

		batchStart := time.Now()

		for i := 0; i < n; i++ {

			reservationID := inventory.NextReservationID()

			reserveBatch[i] = struct {
				EventID       uint64
				Quantity      uint32
				ReservationID uint64
				ShardID       int
			}{
				EventID:       batch[i].EventID,
				Quantity:      batch[i].Quantity,
				ReservationID: reservationID,
				ShardID:       shard.ID,
			}
		}

		errs := store.ReserveBatch(
			context.Background(),
			reserveBatch[:n],
			buf,
		)

		for i := 0; i < n; i++ {

			req := batch[i]

			err := errs[i]

			req.ResponseCh <- ReservationResult{
				ReservationID: reserveBatch[i].ReservationID,
				TimestampNs:   batchStart.UnixNano(),
				Err:           err,
			}

			shard.Counters.TotalProcessed.Add(1)

			statusCode := 0

			if err != nil {
				statusCode = 1
			}

			telemetry.GlobalAsync.Emit(
				telemetry.Event{
					TimestampNs:   time.Now().UnixNano(),
					ShardID:       shard.ID,
					Quantity:      int(req.Quantity),
					EventID:       req.EventID,
					ReservationID: reserveBatch[i].ReservationID,
					LatencyNs:     time.Since(batchStart).Nanoseconds(),
					StatusCode:    statusCode,
				},
			)
		}
	}
}
