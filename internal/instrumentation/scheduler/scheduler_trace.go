package scheduler

import "runtime"

type SchedulerStats struct {
	Goroutines uint64
	CgoCalls   uint64
}

func Capture() SchedulerStats {
	return SchedulerStats{
		Goroutines: uint64(runtime.NumGoroutine()),
		CgoCalls:   uint64(runtime.NumCgoCall()),
	}
}
