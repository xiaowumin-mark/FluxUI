package state

import (
	"fmt"
	"sync"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
)

// AsyncStatus 表示异步操作的状态。
type AsyncStatus int

const (
	AsyncIdle    AsyncStatus = iota
	AsyncLoading
	AsyncSuccess
	AsyncError
)

type asyncSlot[T any] struct {
	mu     sync.Mutex
	gen    uint64
	status AsyncStatus
	data   T
	err    error
}

// AsyncHandle 是异步操作的句柄，提供 Run/Loading/Data/Error/Status/Reset 方法。
type AsyncHandle[T any] struct {
	key     string
	cell    *asyncSlot[T]
	runtime *internal.Runtime
}

// UseAsync 创建或读取当前作用域下的异步状态。
func UseAsync[T any](ctx *internal.Context) *AsyncHandle[T] {
	key := ctx.NextKey("async")
	value := ctx.Persistent(key, func() any {
		return &asyncSlot[T]{}
	})

	cell, ok := value.(*asyncSlot[T])
	if !ok {
		panic(fmt.Sprintf("FluxUI/state: key %q 的异步状态类型发生变化", key))
	}

	return &AsyncHandle[T]{
		key:     key,
		cell:    cell,
		runtime: ctx.Runtime(),
	}
}

// Run 启动异步操作。重复调用会忽略上一次未完成的结果。
func (h *AsyncHandle[T]) Run(fn func() (T, error)) {
	if h == nil || h.cell == nil || fn == nil {
		return
	}

	h.cell.mu.Lock()
	h.cell.gen++
	gen := h.cell.gen
	h.cell.status = AsyncLoading
	var zero T
	h.cell.data = zero
	h.cell.err = nil
	h.cell.mu.Unlock()

	if h.runtime != nil {
		h.runtime.RequestRedraw()
	}

	go func() {
		data, err := fn()

		h.cell.mu.Lock()
		if h.cell.gen != gen {
			h.cell.mu.Unlock()
			return
		}
		if err != nil {
			h.cell.status = AsyncError
			h.cell.err = err
		} else {
			h.cell.status = AsyncSuccess
			h.cell.data = data
		}
		h.cell.mu.Unlock()

		if h.runtime != nil {
			h.runtime.RequestRedraw()
		}
	}()
}

// Status 返回当前异步状态。
func (h *AsyncHandle[T]) Status() AsyncStatus {
	if h == nil || h.cell == nil {
		return AsyncIdle
	}
	h.cell.mu.Lock()
	s := h.cell.status
	h.cell.mu.Unlock()
	return s
}

// Loading 返回是否正在加载。
func (h *AsyncHandle[T]) Loading() bool {
	return h.Status() == AsyncLoading
}

// Data 返回异步操作的结果数据。
func (h *AsyncHandle[T]) Data() T {
	if h == nil || h.cell == nil {
		var zero T
		return zero
	}
	h.cell.mu.Lock()
	d := h.cell.data
	h.cell.mu.Unlock()
	return d
}

// Error 返回异步操作的错误。
func (h *AsyncHandle[T]) Error() error {
	if h == nil || h.cell == nil {
		return nil
	}
	h.cell.mu.Lock()
	e := h.cell.err
	h.cell.mu.Unlock()
	return e
}

// Reset 将异步状态重置为 Idle。
func (h *AsyncHandle[T]) Reset() {
	if h == nil || h.cell == nil {
		return
	}
	h.cell.mu.Lock()
	h.cell.gen++
	h.cell.status = AsyncIdle
	var zero T
	h.cell.data = zero
	h.cell.err = nil
	h.cell.mu.Unlock()

	if h.runtime != nil {
		h.runtime.RequestRedraw()
	}
}
