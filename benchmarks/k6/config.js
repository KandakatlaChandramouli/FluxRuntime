// Phase 1A k6 benchmark configuration.
// All parameters match SYSTEM_SPEC.md §9.

export const BASE_URL = __ENV.BASE_URL || 'grpc://localhost:50051';

export const TARGET_EVENT_HOT  = 101;   // §9.3 Phase 2: hot-key target
export const TARGET_INVENTORY  = 100;   // §9.3 Phase 2: tickets available
export const PEAK_RPS          = 50000;
export const WARMUP_RPS        = 5000;
