#!/usr/bin/env bash

set -e

mkdir -p benchmarks/exported

go test -v ./internal/experiments/...

go test -bench=. -benchmem ./internal/evaluation/... \
| tee benchmarks/exported/comparisons.txt

go test -run TestRuntimeSoak -v ./internal/workerpool/... \
| tee benchmarks/exported/soak.txt
