<!-- fluxui-doc-meta
{
  "id": "hooks_lifecycle",
  "title": "Hooks 与生命周期",
  "category": "状态与副作用",
  "order": 95,
  "summary": "FluxUI React-style runtime 的 Hook 与生命周期能力。",
  "example": { "id": "hooks_lifecycle_basic" },
  "apis": [
    "RunElement(root Component, opts ...RunOption) error",
    "type Component func(ctx *Context) Element",
    "UseState[T](ctx *Context, initial T) *State[T]",
    "UseEffect(ctx *Context, effect Effect)",
    "UseEffectWithDeps(ctx *Context, deps []any, effect Effect)",
    "UseMount(ctx *Context, effect Effect)",
    "UseLifecycle(ctx *Context, onMount func(), onUnmount func())",
    "FromWidget(w Widget) Element"
  ]
}
-->

# Hooks 与生命周期

FluxUI 保持 Gio 的 immediate-mode 渲染模型，同时提供 React-style `Element` / `Component` runtime。`RunElement` root 会通过 reconciler 管理 component instance、HookSlot、effect cleanup、provider scope 和 keyed identity。

## React-style runtime 状态

- React-style runtime 已完成并可通过 `RunElement` 作为 root 入口使用。
- `Element` / `Component` / HookSlot / effect cleanup / provider / keyed reconciler 已接入 runtime。
- Legacy `Run` / `Widget` 继续保留，旧项目无需迁移；新旧混用通过 `FromWidget` 连接。

## 核心能力

- `RunElement`：React-style root 入口，适合新组件和新示例。
- `Component`：函数组件签名，返回 `Element`。
- `UseState`：在 component instance 上优先使用 HookSlot 保存状态；没有 component instance 时回退到 legacy path state。
- `UseEffect`：每次渲染后执行一次副作用；在下一次执行前先清理上次的 cleanup。
- `UseEffectWithDeps`：首次渲染执行，之后仅在依赖变化时执行。
- `UseMount`：只在挂载时执行一次；返回的 cleanup 在卸载时执行。
- `UseLifecycle`：快捷绑定 `onMount/onUnmount`。
- `FromWidget`：长期保留的 Widget -> Element bridge，用于在 React-style tree 中复用 legacy widget。

## Legacy state 与 HookSlot

- `ui.State[T](ctx)` 是 legacy path-state API，继续稳定可用。
- `ui.UseState[T](ctx, initial)` 是 React-style Hook API，在 `RunElement` / component instance 中优先归属于 HookSlot。
- Hook 调用顺序仍必须稳定；条件分支中增减 Hook 调用会破坏状态对应关系。
- Legacy widget state、ref command queue 和 component HookSlot 是不同层级，不应混为同一个状态所有者。

## 执行时机

- 所有 effect 都在当前 frame 组件树构建完成后统一调度执行。
- effect 不会在 layout 绘制中直接执行，避免布局阶段副作用导致的抖动。

## 依赖比较规则

- 依赖为 `[]any`。
- 框架使用深比较判断依赖是否变化。
- 推荐传入稳定、可预期的值（如基础类型、结构体快照），避免把瞬态对象作为依赖。

## 示例

```go
func Counter(ctx *ui.Context) ui.Element {
    count := ui.UseState(ctx, 0)

    ui.UseEffectWithDeps(ctx, []any{count.Value()}, func() func() {
        // count 变化后执行
        return nil
    })

    ui.UseMount(ctx, func() func() {
        // mount
        return func() {
            // unmount cleanup
        }
    })

    return ui.TextElement(fmt.Sprintf("count = %d", count.Value()))
}
```
