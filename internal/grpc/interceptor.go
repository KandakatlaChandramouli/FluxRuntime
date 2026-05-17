package grpc

import (
	"context"
	"runtime/debug"

	"github.com/research/phase1a/internal/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PanicRecoveryInterceptor catches any panic that escapes a handler,
// logs the incident with error code 5001 (WORKER_PANIC_RECOVERED),
// and returns gRPC Internal rather than crashing the process.
//
// Stack traces are emitted only on panic — never on the success path.
func PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				telemetry.Logger.Error(
					"WORKER_PANIC_RECOVERED",
					"error_code", 5001,
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()),
				)
				err = status.Error(codes.Internal, "WORKER_PANIC_RECOVERED")
			}
		}()
		return handler(ctx, req)
	}
}
