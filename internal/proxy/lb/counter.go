package lb

import (
	"sync"
	"sync/atomic"
)

type ConnTracker interface {
	Acquire(host string)
	Release(host string)
}

type ConnCounter struct {
	counters map[string]*atomic.Int64
	mu       sync.RWMutex
}

func (c *ConnCounter) Acquire(host string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.counters[host].Add(1)
}

func (c *ConnCounter) Release(host string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v := c.counters[host].Load(); v > 0 {
		c.counters[host].Add(-1)
	}
}
