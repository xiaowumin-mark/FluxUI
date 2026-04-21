package state

import "sync"

type slot[T any] struct {
	mu    sync.Mutex
	value T
}
