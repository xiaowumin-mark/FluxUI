package state

import (
	"fmt"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// State 是稳定绑定到组件上下文的泛型状态。
type State[T any] struct {
	key     string
	cell    *slot[T]
	runtime *internal.Runtime
}

// Use 创建或读取当前作用域下的状态。
func Use[T any](ctx *internal.Context) *State[T] {
	key := nextKey(ctx)
	value := ctx.Persistent(key, func() any {
		return &slot[T]{}
	})

	cell, ok := value.(*slot[T])
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUIstate: key %q 的状态类型发生变化", key))
	}

	return &State[T]{
		key:     key,
		cell:    cell,
		runtime: ctx.Runtime(),
	}
}

// Key 返回当前状态的稳定 key。
func (s *State[T]) Key() string {
	if s == nil {
		return ""
	}
	return s.key
}

// Value 返回当前状态值。
func (s *State[T]) Value() T {
	if s == nil || s.cell == nil {
		var zero T
		return zero
	}
	return s.cell.value
}

// Set 更新状态并请求重绘。
func (s *State[T]) Set(v T) {
	if s == nil || s.cell == nil {
		return
	}
	s.cell.value = v
	if s.runtime != nil {
		s.runtime.RequestRedraw()
	}
}
