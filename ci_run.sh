#!/usr/bin/env bash
set -e

echo "== TESTS =="

go test ./...

echo "== BENCHMARKS =="

go test \
-run=^$ \
-bench=. \
-benchmem \
-benchtime=1s \
-count=3 \
./... | tee benchmarks/raw/full_benchmark.txt

echo "== BENCHSTAT =="

benchstat benchmarks/raw/full_benchmark.txt \
> benchmarks/statistics/benchstat.txt

echo "== FIGURES =="

python3 scripts/visualization/generate_all.py

echo "== COMPLETE =="
