package mutexqueue

import (
	"sync"

	"github.com/research/phase1a/internal/workerpool"
)

type Queue struct {
	mu       sync.Mutex
	data     []workerpool.ReservationRequest
	head     int
	tail     int
	size     int
	capacity int
}

func New(capacity int) *Queue {
	return &Queue{
		data:     make([]workerpool.ReservationRequest, capacity),
		capacity: capacity,
	}
}

func (q *Queue) Enqueue(r workerpool.ReservationRequest) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size == q.capacity {
		return false
	}

	q.data[q.tail] = r
	q.tail = (q.tail + 1) % q.capacity
	q.size++

	return true
}

func (q *Queue) DequeueBatch(max int, out []workerpool.ReservationRequest) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size == 0 {
		return 0
	}

	n := max
	if q.size < n {
		n = q.size
	}

	for i := 0; i < n; i++ {
		out[i] = q.data[q.head]
		q.head = (q.head + 1) % q.capacity
	}

	q.size -= n

	return n
}

func (q *Queue) Len() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return uint64(q.size)
}

func (q *Queue) Capacity() uint64 {
	return uint64(q.capacity)
}
