#!/usr/bin/env bash
set -euo pipefail

# Phase 1A three-stage benchmark runner.
# All stages run sequentially per spec §9.3.

source "$(dirname "$0")/load-env.sh"

echo "[phase1a] Stage 1: Warm-Up (0→30s)"
k6 run benchmarks/k6/warmup.js

echo "[phase1a] Stage 2: Contention Shockwave (30→60s)"
k6 run benchmarks/k6/hotkey.js

echo "[phase1a] Stage 3: Depletion Tail (60→90s)"
k6 run benchmarks/k6/depletion.js

echo "[phase1a] Benchmark complete."
