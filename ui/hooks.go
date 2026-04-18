package ui

import "fluxui/state"

// Effect 表示渲染后执行的副作用函数，返回 cleanup（可选）。
type Effect = state.Effect

// UseEffect 在每次渲染后执行副作用，并在下一次执行前清理。
func UseEffect(ctx *Context, effect Effect) {
	state.UseEffect(ctx, effect)
}

// UseEffectWithDeps 在首次渲染和依赖变化时执行副作用。
func UseEffectWithDeps(ctx *Context, deps []any, effect Effect) {
	state.UseEffectWithDeps(ctx, deps, effect)
}

// UseMount 在组件挂载时执行一次，卸载时执行 cleanup（如果存在）。
func UseMount(ctx *Context, effect Effect) {
	state.UseMount(ctx, effect)
}

// UseLifecycle 绑定组件挂载/卸载生命周期。
func UseLifecycle(ctx *Context, onMount func(), onUnmount func()) {
	state.UseLifecycle(ctx, onMount, onUnmount)
}
