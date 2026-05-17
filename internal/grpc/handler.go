package grpc

import (
	"context"
	"time"

	ticketv1 "github.com/research/phase1a/internal/gen/ticket/v1"
	"github.com/research/phase1a/internal/workerpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the TicketReservationService gRPC contract.
// It translates inbound RPCs into ReservationRequests and routes them
// through the Dispatcher without blocking or retrying.
type Handler struct {
	ticketv1.UnimplementedTicketReservationServiceServer
	dispatcher *workerpool.Dispatcher
}

// NewHandler constructs a Handler backed by the given Dispatcher.
func NewHandler(d *workerpool.Dispatcher) *Handler {
	return &Handler{dispatcher: d}
}

// resultPool holds pre-allocated response channel slots.
// Each RPC acquires one channel, waits for the worker result, then releases it.
// Channels are buffered(1) so workers never block on send.
//
// Pool capacity == max in-flight RPCs; sized conservatively.
// Using sync.Pool here would cause GC churn; fixed allocation at startup is correct.

// ReserveTicket is the hot-path RPC handler.
// All allocation is bounded and fixed-size.
//
// Flow per spec §2.2:
//
//	Inbound gRPC Request → Shard Routing → Bounded Queue Insert → Response
func (h *Handler) ReserveTicket(
	ctx context.Context,
	req *ticketv1.ReserveTicketRequest,
) (*ticketv1.ReserveTicketResponse, error) {
	// Pre-allocate response channel. Buffered(1): worker never blocks.
	responseCh := make(chan workerpool.ReservationResult, 1)

	wreq := workerpool.ReservationRequest{
		EventID:    req.EventId,
		Quantity:   req.Quantity,
		UserIDHigh: req.UserIdHigh,
		UserIDLow:  req.UserIdLow,
		ResponseCh: responseCh,
		Ctx:        ctx,
	}

	// Non-blocking dispatch. Returns ErrQueueFull immediately if saturated.
	if err := h.dispatcher.Dispatch(wreq); err != nil {
		protoStatus, grpcErr := errorToResponse(err)
		return &ticketv1.ReserveTicketResponse{
			Status:            protoStatus,
			TimestampUnixNano: time.Now().UnixNano(),
		}, grpcErr
	}

	// Wait for worker result or context cancellation.
	select {
	case result := <-responseCh:
		protoStatus, grpcErr := errorToResponse(result.Err)
		return &ticketv1.ReserveTicketResponse{
			Status:            protoStatus,
			ReservationId:     result.ReservationID,
			TimestampUnixNano: result.TimestampNs,
		}, grpcErr

	case <-ctx.Done():
		return &ticketv1.ReserveTicketResponse{
			Status:            ticketv1.ReserveTicketResponse_STATUS_TIMEOUT,
			TimestampUnixNano: time.Now().UnixNano(),
		}, status.Error(codes.DeadlineExceeded, "REDIS_TIMEOUT")
	}
}
