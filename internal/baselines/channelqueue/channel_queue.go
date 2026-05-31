package channelqueue

import (
	"github.com/research/phase1a/internal/workerpool"
)

type Queue struct {
	ch chan workerpool.ReservationRequest
}

func New(capacity int) *Queue {
	return &Queue{
		ch: make(chan workerpool.ReservationRequest, capacity),
	}
}

func (q *Queue) Enqueue(r workerpool.ReservationRequest) bool {
	select {
	case q.ch <- r:
		return true
	default:
		return false
	}
}

func (q *Queue) DequeueBatch(max int, out []workerpool.ReservationRequest) int {

	n := 0

	for n < max {

		select {

		case req := <-q.ch:
			out[n] = req
			n++

		default:
			return n
		}
	}

	return n
}

func (q *Queue) Len() uint64 {
	return uint64(len(q.ch))
}

func (q *Queue) Capacity() uint64 {
	return uint64(cap(q.ch))
}
