High-Throughput Reservation Runtime

Research-oriented high-performance reservation runtime implemented in Go.

## Features

- Deterministic shard routing
- Lock-free ring buffers
- Adaptive batching
- Overload control
- Runtime telemetry
- Zero-allocation hotpath execution

## Benchmark Highlights

- ~10M req/sec throughput
- ~115 ns/op
- 0 allocs/op
- Adaptive rejection stabilization

## Repository Structure

internal/
docs/
benchmarks/
scripts/

## Status

Research/runtime prototype complete.
