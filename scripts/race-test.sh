#!/usr/bin/env bash
set -euo pipefail

# Race detector + benchmark memory test per spec §8.2.

echo "[phase1a] Running race detector..."
go test -race ./...

echo "[phase1a] Running benchmark with memory profiling..."
go test -bench=. -benchmem ./...

echo "[phase1a] All checks passed."
