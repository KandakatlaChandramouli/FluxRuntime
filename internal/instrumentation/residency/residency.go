package residency

import (
	"sync/atomic"
	"time"
)

type Tracker struct {
	totalNs atomic.Uint64
	count   atomic.Uint64
	maxNs   atomic.Uint64
}

func (t *Tracker) Record(start int64) {
	if start <= 0 {
		return
	}

	now := time.Now().UnixNano()

	if now <= start {
		return
	}

	lat := uint64(now - start)

	t.totalNs.Add(lat)
	t.count.Add(1)

	for {
		cur := t.maxNs.Load()

		if lat <= cur {
			return
		}

		if t.maxNs.CompareAndSwap(cur, lat) {
			return
		}
	}
}

func (t *Tracker) Avg() uint64 {
	c := t.count.Load()

	if c == 0 {
		return 0
	}

	return t.totalNs.Load() / c
}

func (t *Tracker) Max() uint64 {
	return t.maxNs.Load()
}
