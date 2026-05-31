#!/usr/bin/env bash
set -e

bash scripts/benchmarks/run_all.sh

python3 scripts/analysis/analyze.py
