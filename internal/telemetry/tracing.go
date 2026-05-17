package telemetry

import (
	"io"
	"log/slog"
)

// Logger is the singleton structured logger for Phase 1A.
// All output is single-line, machine-parseable JSON.
// No multiline logs. No stack traces on success path.
var Logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// LogReservation emits a single structured log line for a reservation outcome.
// All fields are primitives to prevent heap escape.
func LogReservation(
	timestampNs int64,
	shardID int,
	workerID int,
	eventID uint64,
	reservationID uint64,
	latencyNs int64,
	grpcStatus int,
) {
	Logger.LogAttrs(
		nil,
		slog.LevelInfo,
		"reservation",
		slog.Int64("timestamp_ns", timestampNs),
		slog.Int("shard_id", shardID),
		slog.Int("worker_id", workerID),
		slog.Uint64("event_id", eventID),
		slog.Uint64("reservation_id", reservationID),
		slog.Int64("latency_ns", latencyNs),
		slog.Int("grpc_status", grpcStatus),
	)
}
