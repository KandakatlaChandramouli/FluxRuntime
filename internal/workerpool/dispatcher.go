package workerpool

import (
	"context"
	"hash/fnv"
	"math/rand"

	"github.com/research/phase1a/internal/config"
	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
)

const (
	softQueueLimit = 2048
)

type Dispatcher struct {
	shards      []*Shard
	workerCount int
	cancel      context.CancelFunc
}

func NewDispatcher(
	workerCount int,
	store *inventory.Store,
	metrics *telemetry.Metrics,
) *Dispatcher {

	ctx, cancel := context.WithCancel(context.Background())

	d := &Dispatcher{
		shards:      make([]*Shard, workerCount),
		workerCount: workerCount,
		cancel:      cancel,
	}

	for i := 0; i < workerCount; i++ {

		shard := newShard(i)

		d.shards[i] = shard

		go worker(
			shard,
			store,
			metrics,
			ctx,
		)
	}

	return d
}

func (d *Dispatcher) Close() {
	d.cancel()
}

// Dispatch routes a request deterministically to a shard.
func (d *Dispatcher) Dispatch(req ReservationRequest) error {

	shard := d.route(req.EventID)

	shard.Counters.TotalIngested.Add(1)

	depth := shard.queue.Len()

	if depth > softQueueLimit {

		overload := float64(depth-softQueueLimit) / float64(config.QueueCapacity-softQueueLimit)

		if rand.Float64() < overload {

			shard.Counters.TotalRejected.Add(1)

			return ErrQueueFull
		}
	}

	if shard.queue.Enqueue(req) {

		depth := shard.queue.Len()

		for {

			current := shard.Counters.QueueHighWatermark.Load()

			if depth <= current {
				break
			}

			if shard.Counters.QueueHighWatermark.CompareAndSwap(
				current,
				depth,
			) {
				break
			}
		}

		return nil
	}

	shard.Counters.TotalRejected.Add(1)

	return ErrQueueFull
}

func (d *Dispatcher) route(eventID uint64) *Shard {
	idx := d.shardIndex(eventID)
	return d.shards[idx]
}

// route maps an EventID deterministically to a shard.
func (d *Dispatcher) shardIndex(eventID uint64) int {

	h := fnv.New64a()

	var b [8]byte

	b[0] = byte(eventID)
	b[1] = byte(eventID >> 8)
	b[2] = byte(eventID >> 16)
	b[3] = byte(eventID >> 24)
	b[4] = byte(eventID >> 32)
	b[5] = byte(eventID >> 40)
	b[6] = byte(eventID >> 48)
	b[7] = byte(eventID >> 56)

	_, _ = h.Write(b[:])

	return int(h.Sum64() % uint64(d.workerCount))
}

func (d *Dispatcher) Snapshots() []ShardSnapshot {

	out := make([]ShardSnapshot, len(d.shards))

	for i, shard := range d.shards {
		out[i] = shard.Snapshot()
	}

	return out
}
