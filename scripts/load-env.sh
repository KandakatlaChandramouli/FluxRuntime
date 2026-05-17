#!/usr/bin/env bash
# Load environment variables for Phase 1A benchmark runs.
# Override these for isolated benchmark hosts.

export GRPC_ADDR="${GRPC_ADDR:-:50051}"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"
export BASE_URL="${BASE_URL:-grpc://localhost:50051}"
