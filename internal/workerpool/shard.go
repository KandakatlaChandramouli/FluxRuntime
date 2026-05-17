package workerpool

import (
	"github.com/research/phase1a/internal/config"
	"github.com/research/phase1a/internal/telemetry"
)

// Shard owns exactly one bounded queue and one set of atomic counters.
// One goroutine (the worker) drains the queue sequentially.
// No other goroutine may write to the queue except via Dispatch.
type Shard struct {
	ID       int
	queue    *RingBuffer
	Counters telemetry.ShardCounters
}

// newShard initializes a Shard with a bounded queue of spec-mandated capacity.
func newShard(id int) *Shard {
	return &Shard{
		ID:    id,
		queue: NewRingBuffer(uint64(config.QueueCapacity)),
	}
}

// Dispatch attempts a non-blocking insert into the shard's queue.
// Returns ErrQueueFull immediately if the queue is at capacity.
// Blocking insertion is forbidden per spec §5.5.
func (s *Shard) Dispatch(req ReservationRequest) error {

	s.Counters.TotalIngested.Add(1)

	if s.queue.Enqueue(req) {

		current := s.queue.Len()

		for {

			high := s.Counters.QueueHighWatermark.Load()

			if current <= high {
				break
			}

			if s.Counters.QueueHighWatermark.CompareAndSwap(
				high,
				current,
			) {
				break
			}
		}

		return nil
	}

	s.Counters.TotalRejected.Add(1)

	return ErrQueueFull
}
