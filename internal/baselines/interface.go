package baselines

import "github.com/research/phase1a/internal/workerpool"

type Queue interface {
	Enqueue(workerpool.ReservationRequest) bool
	DequeueBatch(int, []workerpool.ReservationRequest) int
	Len() uint64
	Capacity() uint64
}
