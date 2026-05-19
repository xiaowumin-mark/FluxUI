package state

import "sync"

type hookSlotCell[T any] struct {
	mu    sync.Mutex
	value T
}

func (c *hookSlotCell[T]) Value() T {
	c.mu.Lock()
	v := c.value
	c.mu.Unlock()
	return v
}

func (c *hookSlotCell[T]) Set(v T) {
	c.mu.Lock()
	c.value = v
	c.mu.Unlock()
}
