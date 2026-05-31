package scheduler

import (
	"runtime"
	"sync/atomic"
)

type SchedulerStats struct {
	Goroutines uint64
	CgoCalls   uint64
}

var Global SchedulerStats

func Capture() {
	atomic.StoreUint64(
		&Global.Goroutines,
		uint64(runtime.NumGoroutine()),
	)

	atomic.StoreUint64(
		&Global.CgoCalls,
		uint64(runtime.NumCgoCall()),
	)
}
