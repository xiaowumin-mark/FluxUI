package state

import (
	"fmt"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// State 是稳定绑定到组件上下文的泛型状态。
type State[T any] struct {
	key     string
	cell    stateCell[T]
	runtime *internal.Runtime
}

type stateCell[T any] interface {
	Value() T
	Set(T)
}

// Use 创建或读取当前作用域下的状态。
func Use[T any](ctx *internal.Context) *State[T] {
	return useWithInitial(ctx, false, *new(T))
}

// UseWithInitial 创建或读取当前作用域下的状态，并在首次创建时写入初始值。
func UseWithInitial[T any](ctx *internal.Context, initial T) *State[T] {
	return useWithInitial(ctx, true, initial)
}

func useWithInitial[T any](ctx *internal.Context, hasInitial bool, initial T) *State[T] {
	if ctx == nil {
		return &State[T]{}
	}
	key := nextKey(ctx)
	value := ctx.Persistent(key, func() any {
		return &slot[T]{
			value:       initial,
			initialized: hasInitial,
		}
	})

	cell, ok := value.(*slot[T])
	if !ok {
		panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUIstate: key %q 的状态类型发生变化", key))
	}
	if hasInitial {
		cell.mu.Lock()
		if !cell.initialized {
			cell.value = initial
			cell.initialized = true
		}
		cell.mu.Unlock()
	}

	return &State[T]{
		key:     key,
		cell:    cell,
		runtime: ctx.Runtime(),
	}
}

// FromHookSlot creates State backed by a component-owned hook slot.
// This is used by the experimental React-style API while legacy Use keeps
// using Context path keys for compatibility.
func FromHookSlot[T any](ctx *internal.Context, key string, hook *internal.HookSlot, initial T) *State[T] {
	if hook == nil {
		return &State[T]{}
	}
	cell, ok := hook.Value.(*hookSlotCell[T])
	if !ok {
		if hook.Initialized && hook.Value != nil {
			panic(fmt.Sprintf("github.com/xiaowumin-mark/FluxUIstate: hook slot %q 的状态类型发生变化", key))
		}
		cell = &hookSlotCell[T]{value: initial}
		hook.Value = cell
		hook.Initialized = true
	} else if !hook.Initialized {
		cell.Set(initial)
		hook.Initialized = true
	}

	var rt *internal.Runtime
	if ctx != nil {
		rt = ctx.Runtime()
	}
	return &State[T]{
		key:     key,
		cell:    cell,
		runtime: rt,
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
	if s.runtime != nil {
		s.runtime.RecordFrameSection(internal.PerfState, 1)
	}
	return s.cell.Value()
}

// Set 更新状态并请求重绘。可从任意 goroutine 安全调用。
func (s *State[T]) Set(v T) {
	if s == nil || s.cell == nil {
		return
	}
	if s.runtime != nil {
		s.runtime.RecordFrameSection(internal.PerfState, 1)
	}
	s.cell.Set(v)
	if s.runtime != nil {
		s.runtime.RequestRedrawReason("state.Set")
	}
}
