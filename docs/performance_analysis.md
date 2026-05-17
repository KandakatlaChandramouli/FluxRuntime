# High-Throughput Deterministic Workerpool Evaluation

## 1. Overview

This project implements a high-throughput deterministic workerpool architecture designed for contention-heavy workloads under sustained pressure.

The system uses deterministic sharding to route requests into bounded queues while maintaining strict backpressure semantics and zero-allocation dispatch behavior.

Primary design goals:

- Deterministic event routing
- Queue isolation under hot-key contention
- Bounded memory growth
- Explicit overload rejection
- Low-latency dispatch
- Concurrent telemetry instrumentation
- Race-safe metrics collection

---

# 2. Architecture

## Dispatcher

The dispatcher maps EventIDs deterministically to shards using FNV-1a hashing.

Properties:

- Stable routing
- Contention localization
- O(1) shard lookup
- Lock-free fast path

---

## Shards

Each shard contains:

- Dedicated bounded queue
- Worker goroutine
- Atomic telemetry counters

Isolation properties:

- Prevents cross-shard queue interference
- Maintains deterministic hot-key ownership
- Enables independent queue saturation analysis

---

## Worker Execution Model

Each worker processes requests sequentially from its assigned shard queue.

Constraints enforced:

- No dynamic goroutine spawning
- No heap amplification
- No reflection
- No JSON serialization
- No per-request allocation

This preserves allocator stability under pressure.

---

# 3. Benchmark Methodology

## Hardware

- Apple M1
- macOS Darwin ARM64

---

## Workload

Benchmarks simulate sustained hot-key contention against a single EventID.

Characteristics:

- Deterministic routing
- Queue saturation pressure
- High rejection rates
- Fixed-capacity shard queues
- Concurrent request generation

---

## Benchmark Types

### PressureHotKey

Measures:

- Dispatch throughput
- Queue saturation
- Rejection behavior
- Allocation stability
- Latency distribution

---

### PressureMatrix

Evaluates scaling behavior across worker counts:

- 2 workers
- 4 workers
- 8 workers
- 16 workers

---

# 4. Results

## Throughput

Observed dispatch throughput exceeded several million requests per benchmark interval while maintaining:

- Zero allocations per operation
- Stable bounded queues
- Deterministic rejection semantics

---

## Saturation Behavior

Hot-key saturation tests demonstrated:

- Deterministic shard ownership
- Stable queue high-watermarks
- Predictable overload rejection
- Isolation of contention domains

---

## Latency Analysis

Latency instrumentation captured:

- P50 latency
- P95 latency
- P99 latency
- P99.9 latency

Observed characteristics:

- Stable median latency
- Increased tail latency near saturation
- Bounded latency growth under overload

---

# 5. Profiling Analysis

CPU profiling identified major runtime costs:

- Scheduler synchronization
- Channel operations
- Queue contention
- Runtime condition waits

Minimal overhead observed from:

- Hash routing
- Metrics accounting
- Atomic instrumentation

---

# 6. Key Findings

## Deterministic Sharding

Deterministic routing localizes contention efficiently while preserving stable ownership semantics.

---

## Explicit Backpressure

Bounded queues prevent unbounded memory growth during overload scenarios.

---

## Allocation Stability

The dispatch path maintains zero allocations per operation under sustained pressure.

---

## Observability

Atomic telemetry counters provide consistent concurrent metrics without introducing significant synchronization overhead.

---

# 7. Future Work

Potential future improvements include:

- Adaptive shard scaling
- Lock-free ring buffers
- NUMA-aware scheduling
- Distributed shard ownership
- Dynamic queue resizing
- Priority-aware dispatching
- Multi-node contention simulation

---

# 8. Conclusion

This system demonstrates that deterministic sharding combined with bounded queues and explicit overload semantics can maintain stable behavior under sustained contention-heavy workloads while preserving allocator stability and low dispatch overhead.
