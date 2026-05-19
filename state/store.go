package state

import "sync"

type slot[T any] struct {
	mu          sync.Mutex
	value       T
	initialized bool
}

func (s *slot[T]) Value() T {
	s.mu.Lock()
	v := s.value
	s.mu.Unlock()
	return v
}

func (s *slot[T]) Set(v T) {
	s.mu.Lock()
	s.value = v
	s.mu.Unlock()
}
