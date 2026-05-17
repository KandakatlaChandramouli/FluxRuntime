package main

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/research/phase1a/internal/config"
	grpcinternal "github.com/research/phase1a/internal/grpc"
	"github.com/research/phase1a/internal/inventory"
	"github.com/research/phase1a/internal/telemetry"
	"github.com/research/phase1a/internal/workerpool"
)

// System holds all top-level components.
// Components are constructed in dependency order and shut down in reverse.
type System struct {
	store      *inventory.Store
	dispatcher *workerpool.Dispatcher
	grpcServer *grpcinternal.Server
	stopCh     chan struct{}
}

// bootstrap constructs the full Phase 1A system.
// All goroutines are spawned here. None are spawned after this returns.
func bootstrap(ctx context.Context, cfg config.RuntimeConfig) (*System, error) {
	// 1. Redis store + Lua preload.
	store, err := inventory.NewStore(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return nil, err
	}

	// 2. Prometheus registry.
	reg := prometheus.NewRegistry()
	metrics := telemetry.NewMetrics(reg)

	// 3. Dispatcher + workers.
	dispatcher := workerpool.NewDispatcher(
		cfg.WorkerCount,
		store,
		metrics,
	)

	// 4. gRPC handler + server.
	handler := grpcinternal.NewHandler(dispatcher)

	srv, err := grpcinternal.NewServer(cfg.GRPCAddr, handler)
	if err != nil {
		dispatcher.Close()
		store.Close()
		return nil, err
	}
	// 6. Prometheus HTTP exposition (non-hot-path goroutine).
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":9090", mux) //nolint:errcheck

	telemetry.Logger.Info("phase1a bootstrap complete",
		"worker_count", cfg.WorkerCount,
		"grpc_addr", cfg.GRPCAddr,
		"redis_addr", cfg.RedisAddr,
	)

	return &System{
		store:      store,
		dispatcher: dispatcher,
		grpcServer: srv,
	}, nil
}

// shutdown performs an ordered teardown.
func (s *System) shutdown() {
	s.grpcServer.Stop()
	close(s.stopCh)
	s.store.Close()
}
