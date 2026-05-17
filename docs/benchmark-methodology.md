# Benchmark Methodology — Phase 1A

## Scientific Objective

This benchmark suite studies:
- contention collapse under hot-key skew
- queue saturation behavior at bounded capacity
- tail-latency explosion vs. post-depletion stabilization
- oversell prevention under extreme concurrency

It does not produce synthetic throughput vanity metrics.

## Environment Requirements

- Isolated hosts (no shared tenant noise)
- CPU pinning (taskset or cgroup cpuset)
- Background processes disabled
- Fixed runtime.NumCPU() topology
- Redis on local network (<1ms baseline RTT)

## Three-Stage Protocol

### Stage 1: Warm-Up (0–30s)
Ramp from 0 to 5k RPS across randomized event IDs.
Purpose: eliminate JIT cold-start noise, warm Redis keyspace,
fill connection pools, stabilize goroutine scheduler.

### Stage 2: Contention Shockwave (30–60s)
Spike to 50k+ RPS with 99% traffic on EventID=101, inventory=100.
Purpose: force hot-key saturation, measure queue imbalance,
validate zero oversell guarantee under extreme concurrency.

### Stage 3: Depletion Tail (60–90s)
Maintain peak RPS after inventory exhaustion.
Purpose: measure rejection path efficiency, validate Lua -1 fast-path,
observe P99 stabilization hypothesis.

## Required Metric Collection

All metrics listed in SYSTEM_SPEC.md §9.4 must be captured per run.
Runs without complete metric snapshots are invalid.
