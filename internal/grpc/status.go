package grpc

import (
	"errors"

	ticketv1 "github.com/research/phase1a/internal/gen/ticket/v1"
	"github.com/research/phase1a/internal/inventory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errorToResponse maps a hot-path error to the canonical gRPC response status
// and protobuf Status enum per spec §10.
//
// All error values are pre-allocated sentinels — no allocation on this path.
func errorToResponse(err error) (ticketv1.ReserveTicketResponse_Status, error) {
	if err == nil {
		return ticketv1.ReserveTicketResponse_STATUS_SUCCESS, nil
	}

	if errors.Is(err, inventory.ErrQueueFull) {
		return ticketv1.ReserveTicketResponse_STATUS_REJECTED,
			status.Error(codes.ResourceExhausted, "QUEUE_SATURATED")
	}

	if errors.Is(err, inventory.ErrInventoryExhausted) {
		return ticketv1.ReserveTicketResponse_STATUS_EXHAUSTED,
			status.Error(codes.FailedPrecondition, "STOCK_EXHAUSTED")
	}

	if errors.Is(err, inventory.ErrInventoryMissing) {
		return ticketv1.ReserveTicketResponse_STATUS_INTERNAL,
			status.Error(codes.NotFound, "INVENTORY_KEY_MISSING")
	}

	if errors.Is(err, inventory.ErrLuaExecution) {
		return ticketv1.ReserveTicketResponse_STATUS_INTERNAL,
			status.Error(codes.Internal, "LUA_EXECUTION_FAILURE")
	}

	return ticketv1.ReserveTicketResponse_STATUS_INTERNAL,
		status.Error(codes.Internal, "WORKER_PANIC_RECOVERED")
}
