<!-- fluxui-doc-meta
{
  "id": "hooks_lifecycle",
  "title": "Hooks 与生命周期",
  "category": "状态与副作用",
  "order": 95,
  "summary": "FluxUI 的半 React 化能力：UseEffect、依赖副作用、挂载/卸载生命周期。",
  "example": { "id": "hooks_lifecycle_basic" },
  "apis": [
    "UseEffect(ctx *Context, effect Effect)",
    "UseEffectWithDeps(ctx *Context, deps []any, effect Effect)",
    "UseMount(ctx *Context, effect Effect)",
    "UseLifecycle(ctx *Context, onMount func(), onUnmount func())"
  ]
}
-->

# Hooks 与生命周期

FluxUI 保持 Gio 的 immediate-mode 渲染模型，同时补充了一套“半 React 化”的组件副作用机制。

## 核心能力

- `UseEffect`：每次渲染后执行一次副作用；在下一次执行前先清理上次的 cleanup。
- `UseEffectWithDeps`：首次渲染执行，之后仅在依赖变化时执行。
- `UseMount`：只在挂载时执行一次；返回的 cleanup 在卸载时执行。
- `UseLifecycle`：快捷绑定 `onMount/onUnmount`。

## 执行时机

- 所有 effect 都在当前 frame 组件树构建完成后统一调度执行。
- effect 不会在 layout 绘制中直接执行，避免布局阶段副作用导致的抖动。

## 依赖比较规则

- 依赖为 `[]any`。
- 框架使用深比较判断依赖是否变化。
- 推荐传入稳定、可预期的值（如基础类型、结构体快照），避免把瞬态对象作为依赖。

## 示例

```go
counter := ui.State[int](ctx)
ui.UseEffectWithDeps(ctx, []any{counter.Value()}, func() func() {
    // 当 counter 变化后执行
    return nil
})

ui.UseMount(ctx, func() func() {
    // mount
    return func() {
        // unmount cleanup
    }
})
```

