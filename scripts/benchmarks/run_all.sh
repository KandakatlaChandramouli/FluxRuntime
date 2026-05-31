#!/usr/bin/env bash
set -e

mkdir -p benchmarks/results

go test -bench=. -benchmem ./... \
| tee benchmarks/results/full.txt

go test -run=^$ \
-bench=BenchmarkThroughput \
-cpuprofile=benchmarks/results/cpu.prof \
-memprofile=benchmarks/results/mem.prof \
./...

go tool pprof -top benchmarks/results/cpu.prof \
> benchmarks/results/cpu_top.txt

go tool pprof -top benchmarks/results/mem.prof \
> benchmarks/results/mem_top.txt
