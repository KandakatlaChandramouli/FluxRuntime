package workerpool

import (
	"sync"
	"sync/atomic"
)

type RingBuffer struct {
	head atomic.Uint64
	tail atomic.Uint64

	enqueueMu sync.Mutex
	dequeueMu sync.Mutex

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

	r.enqueueMu.Lock()
	defer r.enqueueMu.Unlock()

	head := r.head.Load()
	tail := r.tail.Load()

	if head-tail >= uint64(len(r.buffer)) {
		return false
	}

	r.buffer[head&r.mask] = req

	r.head.Store(head + 1)

	return true
}

func (r *RingBuffer) Dequeue() (ReservationRequest, bool) {

	r.dequeueMu.Lock()
	defer r.dequeueMu.Unlock()

	head := r.head.Load()
	tail := r.tail.Load()

	if tail == head {
		return ReservationRequest{}, false
	}

	req := r.buffer[tail&r.mask]

	r.tail.Store(tail + 1)

	return req, true
}

func (r *RingBuffer) DequeueBatch(
	max int,
	dst []ReservationRequest,
) int {

	r.dequeueMu.Lock()
	defer r.dequeueMu.Unlock()

	head := r.head.Load()
	tail := r.tail.Load()

	n := 0

	for n < max {

		if tail == head {
			break
		}

		dst[n] = r.buffer[tail&r.mask]

		tail++

		n++
	}

	r.tail.Store(tail)

	return n
}

func (r *RingBuffer) Len() uint64 {

	head := r.head.Load()
	tail := r.tail.Load()

	return head - tail
}
