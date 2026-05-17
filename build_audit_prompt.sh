#!/bin/bash

OUTPUT="redis_audit_prompt.txt"

cat > $OUTPUT <<'PROMPT'
You are acting as a hostile low-level distributed systems performance reviewer.

This is NOT a redesign task.

This is NOT a feature-generation task.

You are forbidden from inventing abstractions, APIs, wrappers, or alternate architectures.

Your ONLY responsibility is to perform an adversarial hot-path performance audit against the REAL existing implementation.

The system already experimentally validated:
- deterministic bounded queues
- non-blocking rejection
- zero-allocation dispatch
- zero-allocation shard routing
- zero-allocation queue enqueue
- stable hot-key contention behavior

Microbenchmark results already achieved:
- BenchmarkDispatchHotKey: 13.24 ns/op, 0 allocs/op
- BenchmarkDispatchColdKeys: 13.59 ns/op, 0 allocs/op
- BenchmarkSaturatedDispatch: 13.99 ns/op, 0 allocs/op
- BenchmarkShardRouting: 5.009 ns/op, 0 allocs/op
- BenchmarkAtomicCounters: 42.31 ns/op, 0 allocs/op

Your job is to determine whether the Redis coordination layer breaks these guarantees.
PROMPT

cat audit_context.txt >> $OUTPUT

cat >> $OUTPUT <<'PROMPT'

==================================================
AUDIT OBJECTIVES
==================================================

You MUST identify:
1. Hidden heap allocations
2. Interface boxing
3. Slice growth allocations
4. string([]byte) allocation boundaries
5. []interface{} allocation pressure
6. []string allocation pressure
7. strconv.AppendUint behavior
8. EvalSha argument escape behavior
9. Lua payload allocation contamination
10. Redis key generation costs
11. sync.Pool opportunities
12. escape-analysis failures
13. scheduler amplification risks
14. cache-line contention risks
15. telemetry contamination inside hot paths

==================================================
IMPORTANT CONSTRAINTS
==================================================

You MUST preserve architecture exactly.

Forbidden:
- redesigns
- dependency injection
- wrappers
- interfaces
- ORMs
- middleware
- microservices
- speculative infrastructure

==================================================
OUTPUT FORMAT
==================================================

1. Allocation Risk Map
2. Escape Analysis Risks
3. Hot Path Contamination Analysis
4. Mechanical Optimization Opportunities
5. Estimated allocs/op Impact
6. Highest Priority Optimization Targets
PROMPT

echo "Generated: $OUTPUT"
