# Latency Analysis — Phase 1A

## Latency Budget (P99 < 10ms @ 50k RPS)

| Segment                    | Budget     |
|----------------------------|------------|
| gRPC decode                | ~50µs      |
| Dispatcher shard routing   | ~1µs       |
| Queue insert (non-blocking)| ~1µs       |
| Worker queue drain wait    | variable   |
| Redis Lua EVALSHA RTT      | < 2ms      |
| gRPC encode + send         | ~50µs      |

Queue drain wait dominates under saturation. At 50k RPS hitting one shard
with one sequential worker, the drain rate is bounded by Redis RTT.

## Tail Amplification Risk

Under hot-key skew:
- One shard absorbs ~50k RPS
- Worker processes ~500 req/s (assuming 2ms Redis RTT)
- Queue fills in ~8.2 seconds at full load
- After fill: all new requests rejected in <1ms

P99 tail behavior is dominated by queueing time before saturation
and by instant rejection latency after saturation.

## Measurement Methodology

All latency measurements are recorded as nanosecond timestamps
using time.Now() at RPC entry and time.Since() at response emit.
No sampling. No aggregation loss on the critical path.

Prometheus histograms use fixed buckets (0.1ms to 100ms) to bound
cardinality and scrape cost.
