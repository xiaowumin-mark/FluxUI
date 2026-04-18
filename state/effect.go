package state

import internal "github.com/xiaowumin-mark/FluxUI/internal"

// Effect 表示渲染后执行的副作用函数，返回 cleanup（可选）。
type Effect func() (cleanup func())

// UseEffect 在每次渲染后执行副作用，并在下一次执行前调用上一次 cleanup。
func UseEffect(ctx *internal.Context, effect Effect) {
	if ctx == nil || effect == nil {
		return
	}
	rt := ctx.Runtime()
	if rt == nil {
		return
	}
	key := ctx.NextKey("effect")
	rt.UseEffect(key, false, nil, internal.EffectSetup(effect))
}

// UseEffectWithDeps 在首次渲染和依赖变化时执行副作用。
func UseEffectWithDeps(ctx *internal.Context, deps []any, effect Effect) {
	if ctx == nil || effect == nil {
		return
	}
	rt := ctx.Runtime()
	if rt == nil {
		return
	}
	key := ctx.NextKey("effect")
	rt.UseEffect(key, true, deps, internal.EffectSetup(effect))
}

// UseMount 仅在组件挂载时执行一次，卸载时执行 cleanup（如果存在）。
func UseMount(ctx *internal.Context, effect Effect) {
	UseEffectWithDeps(ctx, []any{}, effect)
}

// UseLifecycle 为组件提供挂载/卸载钩子。
func UseLifecycle(ctx *internal.Context, onMount func(), onUnmount func()) {
	UseMount(ctx, func() func() {
		if onMount != nil {
			onMount()
		}
		if onUnmount == nil {
			return nil
		}
		return func() {
			onUnmount()
		}
	})
}
