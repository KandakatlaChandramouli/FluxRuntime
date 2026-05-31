package memory

import "runtime"

type Snapshot struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

func Capture() Snapshot {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)

	return Snapshot{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}
