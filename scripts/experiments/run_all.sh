#!/usr/bin/env bash

set -e

go test ./internal/benchmarks/...
go test ./internal/experiments/...

python3 scripts/visualization/generate_all.py
