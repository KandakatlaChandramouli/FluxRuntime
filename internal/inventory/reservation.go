package inventory

import "sync/atomic"

// reservationCounter is the global monotonic reservation ID generator.
// Atomic increment; no locks; no heap allocation.
var reservationCounter atomic.Uint64

// NextReservationID returns the next unique reservation ID.
// IDs are process-scoped monotonic integers.
// Satisfies the uint64 reservation_id field in the protobuf contract.
func NextReservationID() uint64 {
	return reservationCounter.Add(1)
}
