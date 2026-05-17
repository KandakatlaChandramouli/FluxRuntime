package workerpool

import "context"

// ReservationRequest is the unit of work enqueued into a worker's bounded channel.
// All fields are value types; no pointers, no interfaces, no closures.
// This keeps the struct off the heap when passed through the channel.
type ReservationRequest struct {
	EventID    uint64
	Quantity   uint32
	UserIDHigh uint64
	UserIDLow  uint64
	ResponseCh chan ReservationResult
	Ctx        context.Context
}

// ReservationResult carries the outcome back to the gRPC handler.
type ReservationResult struct {
	ReservationID uint64
	TimestampNs   int64
	Err           error
}
