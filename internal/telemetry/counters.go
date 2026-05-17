package telemetry

import "sync/atomic"

// ShardCounters holds per-shard saturation metrics.
// All fields must be accessed via atomic operations only.
// No locks. No channels. No goroutines.
type ShardCounters struct {
	TotalIngested      atomic.Uint64
	TotalRejected      atomic.Uint64
	TotalProcessed     atomic.Uint64
	QueueHighWatermark atomic.Uint64
}

// UpdateHighWatermark performs a lock-free max update.
func (c *ShardCounters) UpdateHighWatermark(current uint64) {
	for {
		old := c.QueueHighWatermark.Load()
		if current <= old {
			return
		}
		if c.QueueHighWatermark.CompareAndSwap(old, current) {
			return
		}
	}
}
