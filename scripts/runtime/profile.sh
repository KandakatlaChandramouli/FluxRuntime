#!/usr/bin/env bash

go test \
-run=^$ \
-bench=. \
-cpuprofile=cpu.prof \
-memprofile=mem.prof \
./internal/workerpool/...
