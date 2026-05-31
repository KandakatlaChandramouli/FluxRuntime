package residency

const (
	MaxReasonableLatencyNs = 60 * 1e9
)

func ValidLatency(ns uint64) bool {
	return ns > 0 && ns < MaxReasonableLatencyNs
}
