# Contention Model — Phase 1A

## Hot-Key Saturation

FNV-1a routing guarantees deterministic shard affinity per EventID.
Under Phase 2 load (99% single EventID at 50k RPS), all hot-key traffic
routes to one shard. This creates the following contention path:

```
50k RPS → Dispatcher → ShardIndex = FNV1a(101) % WorkerCount
                             ↓
                    Single bounded queue (cap: 4096)
                             ↓
                    Single sequential worker
                             ↓
                    Redis Lua EVALSHA
```

Expected observation:
- Queue depth approaches 4096 rapidly
- TotalRejected counter rises sharply
- All other shards remain near-idle
- Oversell count: 0 (Lua atomicity guarantee)

## Queue Collapse Cascade (Expected Failure Mode)

If Redis RTT exceeds ~80µs under saturation, the single worker
falls behind the 4096-entry queue faster than it drains it.
At cap, ErrQueueFull fires instantly (<1ms per spec §5.5).

This is not a bug. It is the measurable behavior this system exists to study.

## Depletion Tail Stabilization

After inventory reaches 0, Lua returns -1 on every call.
Redis RTT for a cache-hot -1 return is substantially lower than
a successful DECRBY + HSET sequence.

Hypothesis: P99 latency decreases post-depletion due to reduced Redis write pressure.
The benchmark validates or falsifies this hypothesis.
