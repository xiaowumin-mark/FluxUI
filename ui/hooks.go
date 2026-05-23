package ui

import (
	"fmt"
	"math"
	"time"

	anim "github.com/xiaowumin-mark/FluxUI/anim"
	internal "github.com/xiaowumin-mark/FluxUI/internal"
	state "github.com/xiaowumin-mark/FluxUI/state"
)

// Effect 表示渲染后执行的副作用函数，返回 cleanup（可选）。
type Effect = state.Effect

// UseEffect 在每次渲染后执行副作用，并在下一次执行前清理。
func UseEffect(ctx *Context, effect Effect) {
	useEffect(ctx, false, nil, effect)
}

// UseEffectWithDeps 在首次渲染和依赖变化时执行副作用。
func UseEffectWithDeps(ctx *Context, deps []any, effect Effect) {
	useEffect(ctx, true, deps, effect)
}

// UseMount 在组件挂载时执行一次，卸载时执行 cleanup（如果存在）。
func UseMount(ctx *Context, effect Effect) {
	UseEffectWithDeps(ctx, []any{}, effect)
}

// UseLifecycle 绑定组件挂载/卸载生命周期。
func UseLifecycle(ctx *Context, onMount func(), onUnmount func()) {
	state.UseLifecycle(ctx, onMount, onUnmount)
}

// UseInterval 在组件挂载期间按固定间隔执行 fn，并在卸载时停止。
func UseInterval(ctx *Context, interval time.Duration, fn func()) {
	state.UseInterval(ctx, interval, fn)
}

// UseMemo memoizes a value until one of deps changes.
func UseMemo[T any](ctx *Context, deps []any, factory func() T) T {
	if factory == nil {
		var zero T
		return zero
	}
	if ctx == nil {
		return factory()
	}
	if hook := ctx.NextHookSlot(internal.HookMemo); hook != nil {
		return useHookMemo(ctx, hook, deps, factory)
	}
	return useLegacyMemo(ctx, deps, factory)
}

// Ref stores a mutable value that persists for a component instance lifetime.
type Ref[T any] struct {
	Current T
}

// UseRef returns a stable mutable ref initialized on first render.
func UseRef[T any](ctx *Context, initial T) *Ref[T] {
	if ctx == nil {
		return &Ref[T]{Current: initial}
	}
	if hook := ctx.NextHookSlot(internal.HookRef); hook != nil {
		return useHookRef(ctx, hook, initial)
	}
	return useLegacyRef(ctx, initial)
}

// UseCallback memoizes a callback or function-like value until deps changes.
func UseCallback[T any](ctx *Context, deps []any, fn T) T {
	return UseMemo(ctx, deps, func() T { return fn })
}

// AsyncStatus 表示异步操作的状态。
type AsyncStatus = state.AsyncStatus

// AsyncHandle 是异步操作的句柄。
type AsyncHandle[T any] = state.AsyncHandle[T]

const (
	AsyncIdle    = state.AsyncIdle
	AsyncLoading = state.AsyncLoading
	AsyncSuccess = state.AsyncSuccess
	AsyncError   = state.AsyncError
)

// UseAsync 创建或读取当前作用域下的异步状态。
func UseAsync[T any](ctx *Context) *AsyncHandle[T] {
	return state.UseAsync[T](ctx)
}

type memoSlot[T any] struct {
	initialized bool
	deps        []any
	value       T
}

func useHookMemo[T any](ctx *Context, hook *internal.HookSlot, deps []any, factory func() T) T {
	if hook.Initialized {
		if value, ok := hook.Value.(T); ok && internal.DepsEqual(hook.Deps, deps) {
			return value
		}
		if _, ok := hook.Value.(T); !ok && hook.Value != nil {
			panic(hookTypeMismatch[T](ctx, "memo"))
		}
	}
	value := factory()
	hook.Value = value
	hook.Initialized = true
	hook.HasDeps = true
	hook.Deps = internal.CloneDeps(deps)
	return value
}

func useLegacyMemo[T any](ctx *Context, deps []any, factory func() T) T {
	key := ctx.NextKey("memo")
	value := ctx.Persistent(key, func() any {
		return &memoSlot[T]{}
	})
	slot, ok := value.(*memoSlot[T])
	if !ok {
		panic(hookTypeMismatch[T](ctx, "memo"))
	}
	if !slot.initialized || !internal.DepsEqual(slot.deps, deps) {
		slot.value = factory()
		slot.deps = internal.CloneDeps(deps)
		slot.initialized = true
	}
	return slot.value
}

func useHookRef[T any](ctx *Context, hook *internal.HookSlot, initial T) *Ref[T] {
	if hook.Initialized {
		ref, ok := hook.Value.(*Ref[T])
		if !ok {
			panic(hookTypeMismatch[T](ctx, "ref"))
		}
		return ref
	}
	ref := &Ref[T]{Current: initial}
	hook.Value = ref
	hook.Initialized = true
	return ref
}

func useLegacyRef[T any](ctx *Context, initial T) *Ref[T] {
	key := ctx.NextKey("ref")
	value := ctx.Persistent(key, func() any {
		return &Ref[T]{Current: initial}
	})
	ref, ok := value.(*Ref[T])
	if !ok {
		panic(hookTypeMismatch[T](ctx, "ref"))
	}
	return ref
}

func hookTypeMismatch[T any](ctx *Context, kind string) string {
	path := ""
	if ctx != nil {
		path = ctx.TreePath()
	}
	return fmt.Sprintf("github.com/xiaowumin-mark/FluxUI/ui: %s hook type mismatch at %q for %T", kind, path, *new(T))
}

type animNum interface {
	~float32 | ~int | ~float64
}

type animValueState[T animNum] struct {
	startedAt time.Time
	from      T
	to        T
	duration  time.Duration
	easing    anim.Easing
}

func animNumLerp[T animNum](from, to T, t float32) T {
	t = animClamp01(t)
	return T(float32(from) + (float32(to)-float32(from))*t)
}

func animClamp01(v float32) float32 {
	if math.IsNaN(float64(v)) {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func animProgress(now, startedAt time.Time, duration time.Duration) (float32, bool) {
	if duration <= 0 {
		return 1, false
	}
	elapsed := now.Sub(startedAt)
	if elapsed <= 0 {
		return 0, true
	}
	if elapsed >= duration {
		return 1, false
	}
	return animClamp01(float32(elapsed) / float32(duration)), true
}

func animEasedProgress(easing anim.Easing, progress float32) float32 {
	if easing == nil {
		easing = anim.Linear
	}
	v := easing(animClamp01(progress))
	if math.IsNaN(float64(v)) {
		return 0
	}
	if math.IsInf(float64(v), 1) {
		return 1
	}
	if math.IsInf(float64(v), -1) {
		return 0
	}
	return v
}

func UseAnimatedValue[T animNum](ctx *Context, target T, duration time.Duration, easing anim.Easing) T {
	if easing == nil {
		easing = anim.Linear
	}
	if ctx == nil {
		return target
	}

	if hook := ctx.NextHookSlot(internal.HookAnimValue); hook != nil {
		return useHookAnimValue(ctx, hook, target, duration, easing)
	}
	return useLegacyAnimValue(ctx, target, duration, easing)
}

func useHookAnimValue[T animNum](ctx *Context, hook *internal.HookSlot, target T, duration time.Duration, easing anim.Easing) T {
	if hook.Initialized {
		state, ok := hook.Value.(*animValueState[T])
		if !ok {
			panic(hookTypeMismatch[T](ctx, "anim_value"))
		}
		if duration <= 0 {
			state.snap(ctx.Now(), target, duration, easing)
			return target
		}
		if state.to == target && state.duration == duration {
			p, running := animProgress(ctx.Now(), state.startedAt, duration)
			state.easing = easing
			if !running {
				return target
			}
			ctx.RequestFrameRedraw()
			return animNumLerp(state.from, state.to, animEasedProgress(easing, p))
		}

		now := ctx.Now()
		var currentValue T
		oldP, running := animProgress(now, state.startedAt, state.duration)
		if !running {
			currentValue = state.to
		} else {
			currentValue = animNumLerp(state.from, state.to, animEasedProgress(state.easing, oldP))
		}
		state.startedAt = now
		state.from = currentValue
		state.to = target
		state.duration = duration
		state.easing = easing
		ctx.RequestFrameRedraw()
		return state.from
	}

	state := &animValueState[T]{
		startedAt: ctx.Now(),
		from:      target,
		to:        target,
		duration:  duration,
		easing:    easing,
	}
	hook.Value = state
	hook.Initialized = true
	return target
}

func (state *animValueState[T]) snap(now time.Time, target T, duration time.Duration, easing anim.Easing) {
	state.startedAt = now
	state.from = target
	state.to = target
	state.duration = duration
	state.easing = easing
}

func useLegacyAnimValue[T animNum](ctx *Context, target T, duration time.Duration, easing anim.Easing) T {
	key := ctx.NextKey("anim_value")
	value := ctx.Persistent(key, func() any {
		return &animValueState[T]{startedAt: ctx.Now(), from: target, to: target, duration: duration, easing: easing}
	})
	state, ok := value.(*animValueState[T])
	if !ok {
		panic(hookTypeMismatch[T](ctx, "anim_value"))
	}
	if duration <= 0 {
		state.snap(ctx.Now(), target, duration, easing)
		return target
	}

	if state.to == target && state.duration == duration {
		p, running := animProgress(ctx.Now(), state.startedAt, duration)
		state.easing = easing
		if !running {
			return target
		}
		ctx.RequestFrameRedraw()
		return animNumLerp(state.from, state.to, animEasedProgress(easing, p))
	}

	now := ctx.Now()
	oldP, running := animProgress(now, state.startedAt, state.duration)
	if !running {
		state.from = state.to
	} else {
		state.from = animNumLerp(state.from, state.to, animEasedProgress(state.easing, oldP))
	}
	state.startedAt = now
	state.to = target
	state.duration = duration
	state.easing = easing
	ctx.RequestFrameRedraw()
	return animNumLerp(state.from, state.to, 0)
}

func UseAnimatedDecoration(ctx *Context, target Decoration, duration time.Duration, easing anim.Easing) Decoration {
	if easing == nil {
		easing = anim.Linear
	}
	if ctx == nil {
		return target
	}

	if hook := ctx.NextHookSlot(internal.HookAnimDeco); hook != nil {
		return useHookAnimDeco(ctx, hook, target, duration, easing)
	}
	return useLegacyAnimDeco(ctx, target, duration, easing)
}

type animDecoState struct {
	startedAt time.Time
	from      Decoration
	current   Decoration
	to        Decoration
	duration  time.Duration
	easing    anim.Easing
}

func useHookAnimDeco(ctx *Context, hook *internal.HookSlot, target Decoration, duration time.Duration, easing anim.Easing) Decoration {
	if hook.Initialized {
		state, ok := hook.Value.(*animDecoState)
		if !ok {
			panic(hookTypeMismatch[Decoration](ctx, "anim_deco"))
		}
		if duration <= 0 {
			state.snap(ctx.Now(), target, duration, easing)
			return target
		}
		if anim.DecorationEqual(state.to, target) && state.duration == duration {
			p, running := animProgress(ctx.Now(), state.startedAt, duration)
			state.easing = easing
			if !running {
				return target
			}
			ctx.RequestFrameRedraw()
			return anim.LerpDecoration(state.from, state.to, animEasedProgress(easing, p))
		}

		now := ctx.Now()
		p, running := animProgress(now, state.startedAt, state.duration)
		if !running {
			state.current = state.to
		} else {
			state.current = anim.LerpDecoration(state.from, state.to, animEasedProgress(state.easing, p))
		}
		state.startedAt = now
		state.from = state.current
		state.to = target
		state.duration = duration
		state.easing = easing
		ctx.RequestFrameRedraw()
		return state.from
	}

	state := &animDecoState{
		startedAt: ctx.Now(),
		from:      target,
		current:   target,
		to:        target,
		duration:  duration,
		easing:    easing,
	}
	hook.Value = state
	hook.Initialized = true
	return target
}

func (state *animDecoState) snap(now time.Time, target Decoration, duration time.Duration, easing anim.Easing) {
	state.startedAt = now
	state.from = target
	state.current = target
	state.to = target
	state.duration = duration
	state.easing = easing
}

func useLegacyAnimDeco(ctx *Context, target Decoration, duration time.Duration, easing anim.Easing) Decoration {
	key := ctx.NextKey("anim_deco")
	value := ctx.Persistent(key, func() any {
		return &animDecoState{
			startedAt: ctx.Now(),
			from:      target,
			current:   target,
			to:        target,
			duration:  duration,
			easing:    easing,
		}
	})
	state, ok := value.(*animDecoState)
	if !ok {
		panic(hookTypeMismatch[Decoration](ctx, "anim_deco"))
	}
	if duration <= 0 {
		state.snap(ctx.Now(), target, duration, easing)
		return target
	}

	if anim.DecorationEqual(state.to, target) && state.duration == duration {
		p, running := animProgress(ctx.Now(), state.startedAt, duration)
		state.easing = easing
		if !running {
			return target
		}
		ctx.RequestFrameRedraw()
		return anim.LerpDecoration(state.from, state.to, animEasedProgress(easing, p))
	}

	now := ctx.Now()
	p, running := animProgress(now, state.startedAt, state.duration)
	if !running {
		state.current = state.to
	} else {
		state.current = anim.LerpDecoration(state.from, state.to, animEasedProgress(state.easing, p))
	}

	state.startedAt = now
	state.from = state.current
	state.to = target
	state.duration = duration
	state.easing = easing
	ctx.RequestFrameRedraw()
	return state.from
}
