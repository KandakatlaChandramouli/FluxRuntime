package red

import "math/rand"

type Controller struct {
	Min uint64
	Max uint64
}

func (c *Controller) Allow(depth uint64) bool {
	if depth <= c.Min {
		return true
	}

	if depth >= c.Max {
		return false
	}

	p := float64(depth-c.Min) / float64(c.Max-c.Min)

	return rand.Float64() > p
}
