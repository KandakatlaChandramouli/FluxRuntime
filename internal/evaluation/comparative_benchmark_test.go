package evaluation

import (
	"testing"

	"github.com/research/phase1a/internal/baselines/channelqueue"
	"github.com/research/phase1a/internal/baselines/mutexqueue"
	"github.com/research/phase1a/internal/baselines/unboundedqueue"
	"github.com/research/phase1a/internal/workerpool"
)

func benchmarkQueue(
	b *testing.B,
	name string,
	queue interface {
		Enqueue(workerpool.ReservationRequest) bool
	},
) {

	req := workerpool.ReservationRequest{
		EventID:  101,
		Quantity: 1,
	}

	b.Run(name, func(b *testing.B) {

		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = queue.Enqueue(req)
		}
	})
}

func BenchmarkQueueComparisons(b *testing.B) {

	benchmarkQueue(
		b,
		"mutex_queue",
		mutexqueue.New(4096),
	)

	benchmarkQueue(
		b,
		"channel_queue",
		channelqueue.New(4096),
	)

	benchmarkQueue(
		b,
		"unbounded_queue",
		unboundedqueue.New(4096),
	)
}
