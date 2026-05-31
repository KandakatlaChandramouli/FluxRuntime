package unboundedqueue

import (
	"sync"

	"github.com/research/phase1a/internal/workerpool"
)

type Queue struct {
	mu   sync.Mutex
	data []workerpool.ReservationRequest
}

func New(initial int) *Queue {
	return &Queue{
		data: make([]workerpool.ReservationRequest, 0, initial),
	}
}

func (q *Queue) Enqueue(r workerpool.ReservationRequest) bool {
	q.mu.Lock()
	q.data = append(q.data, r)
	q.mu.Unlock()
	return true
}

func (q *Queue) DequeueBatch(max int, out []workerpool.ReservationRequest) int {

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.data) == 0 {
		return 0
	}

	n := max
	if len(q.data) < n {
		n = len(q.data)
	}

	copy(out[:n], q.data[:n])

	q.data = q.data[n:]

	return n
}

func (q *Queue) Len() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return uint64(len(q.data))
}

func (q *Queue) Capacity() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return uint64(cap(q.data))
}
