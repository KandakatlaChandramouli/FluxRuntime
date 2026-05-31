package snapshot

type RuntimeSnapshot struct {
	ThroughputReqSec float64
	P50Ns            uint64
	P95Ns            uint64
	P99Ns            uint64
	AvgNs            uint64
	MaxNs            uint64
	QueueDepth       uint64
	Rejected         uint64
	Completed        uint64
	BatchSize        float64
}
