#!/usr/bin/env bash

set -e

mkdir -p benchmarks/comparisons/results

go test -run=^$ \
-bench=. \
-benchmem \
./internal/evaluation/... \
| tee benchmarks/comparisons/results/comparative.txt

go test -run TestRuntimeSoak \
-v ./internal/workerpool/... \
| tee benchmarks/comparisons/results/soak.txt

go test -run=^$ \
-bench=PressureHotKey \
-benchmem \
./internal/workerpool/... \
| tee benchmarks/comparisons/results/hotkey.txt
