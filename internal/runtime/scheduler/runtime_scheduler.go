package scheduler

import "runtime"

func Goroutines() int {
	return runtime.NumGoroutine()
}
