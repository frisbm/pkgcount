package models

import (
	"sync"
	"sync/atomic"
)

type Counter struct {
	c map[string]*atomic.Uint32
	sync.RWMutex
}

func NewCounter() *Counter {
	return &Counter{
		c: make(map[string]*atomic.Uint32),
	}
}

func (c *Counter) Add(pkg string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.c[pkg]; !ok {
		c.c[pkg] = new(atomic.Uint32)
	}
	c.c[pkg].Add(1)
}

func (c *Counter) Counts() map[string]*atomic.Uint32 {
	c.RLock()
	defer c.RUnlock()
	if c.c == nil {
		return make(map[string]*atomic.Uint32)
	}
	return c.c
}
