package codel

import "time"

type Controller struct {
	Target time.Duration
}

func (c *Controller) Overloaded(latency time.Duration) bool {
	return latency > c.Target
}
