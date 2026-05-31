# Experimental Methodology

## Hardware Environment

All experiments were executed on Apple Silicon M1 hardware under macOS Darwin ARM64.

## Runtime Environment

- Go runtime
- goroutine-based concurrency
- lock-free queue execution
- bounded queue pressure evaluation

## Workloads

The runtime was evaluated under:

- hot-key contention
- multicore scaling
- sustained overload ingress
- bounded queue saturation
- probabilistic overload shedding

## Metrics

The evaluation tracks:

- throughput
- p50 latency
- p95 latency
- p99 latency
- maximum latency
- queue depth
- rejection rate
- allocation behavior
- batch evolution

## Baselines

FluxRuntime is compared against:

- mutex-protected queues
- Go channel queues
- unbounded queues

to characterize scheduler contention, queue amplification, and overload stability behavior.
