package backpressure

type State struct {
	QueueDepth uint64
	Rejected   uint64
}
