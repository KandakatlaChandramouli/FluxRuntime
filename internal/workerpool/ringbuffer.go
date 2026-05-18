package workerpool

import (
	"sync/atomic"
)

type RingBuffer struct {
	head atomic.Uint64
	tail atomic.Uint64

	mask uint64

	buffer []ReservationRequest
}

func NewRingBuffer(size uint64) *RingBuffer {

	if size&(size-1) != 0 {
		panic("ring buffer size must be power of 2")
	}

	return &RingBuffer{
		mask:   size - 1,
		buffer: make([]ReservationRequest, size),
	}
}

func (r *RingBuffer) Enqueue(req ReservationRequest) bool {

	for {

		head := r.head.Load()
		tail := r.tail.Load()

		if head-tail >= uint64(len(r.buffer)) {
			return false
		}

		if r.head.CompareAndSwap(head, head+1) {

			slot := head & r.mask

			r.buffer[slot] = req

			return true
		}
	}
}

func (r *RingBuffer) Dequeue() (ReservationRequest, bool) {

	for {

		tail := r.tail.Load()
		head := r.head.Load()

		if tail == head {
			return ReservationRequest{}, false
		}

		if r.tail.CompareAndSwap(tail, tail+1) {

			slot := tail & r.mask

			req := r.buffer[slot]

			return req, true
		}
	}
}

func (r *RingBuffer) DequeueBatch(
	max int,
	dst []ReservationRequest,
) int {

	n := 0

	for n < max {

		req, ok := r.Dequeue()

		if !ok {
			break
		}

		dst[n] = req

		n++
	}

	return n
}

func (r *RingBuffer) Len() uint64 {

	head := r.head.Load()
	tail := r.tail.Load()

	return head - tail
}
