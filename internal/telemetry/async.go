package telemetry

import (
	"sync/atomic"
	"time"
)

type Event struct {
	TimestampNs   int64
	ShardID       int
	Quantity      int
	EventID       uint64
	ReservationID uint64
	LatencyNs     int64
	StatusCode    int
}

type AsyncTelemetry struct {
	queue chan Event

	dropped atomic.Uint64
}

func NewAsyncTelemetry(size int) *AsyncTelemetry {

	t := &AsyncTelemetry{
		queue: make(chan Event, size),
	}

	go t.run()

	return t
}

func (t *AsyncTelemetry) Emit(e Event) {

	select {

	case t.queue <- e:

	default:
		t.dropped.Add(1)
	}
}

func (t *AsyncTelemetry) run() {

	var batch [256]Event

	for {

		n := 0

		select {

		case ev := <-t.queue:
			batch[n] = ev
			n++

		case <-time.After(time.Millisecond):
			continue
		}

	DrainLoop:

		for n < len(batch) {

			select {

			case ev := <-t.queue:
				batch[n] = ev
				n++

			default:
				break DrainLoop
			}
		}

		for i := 0; i < n; i++ {

			_ = batch[i]

			// future exporters/sinks here
		}
	}
}

func (t *AsyncTelemetry) Dropped() uint64 {
	return t.dropped.Load()
}

var GlobalAsync = NewAsyncTelemetry(65536)
