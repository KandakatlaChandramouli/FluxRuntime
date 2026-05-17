package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/research/phase1a/internal/telemetry"
)

func main() {
	cfg := loadConfig()

	ctx := context.Background()

	sys, err := bootstrap(ctx, cfg)
	if err != nil {
		telemetry.Logger.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}

	// Block until SIGINT or SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start serving in background; main blocks on signal.
	errCh := make(chan error, 1)
	go func() {
		errCh <- sys.grpcServer.Serve()
	}()

	select {
	case sig := <-sigCh:
		telemetry.Logger.Info("received signal, shutting down", "signal", sig)
	case err := <-errCh:
		telemetry.Logger.Error("grpc server error", "error", err)
	}

	sys.shutdown()
}
