package graph

import "sync"

/////////////////////////////////////////////////////////////////////
// TYPES

type counter struct {
	sync.RWMutex
	value int
}

/////////////////////////////////////////////////////////////////////
// NEW

func NewCounter() *counter {
	return &counter{}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *counter) Dec(one bool) {
	if one {
		c.RWMutex.Lock()
		defer c.RWMutex.Unlock()
		c.value--
	}
}

func (c *counter) Inc(one bool) {
	if one {
		c.RWMutex.Lock()
		defer c.RWMutex.Unlock()
		c.value++
	}
}

func (c *counter) IsZero() bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	return c.value == 0
}
