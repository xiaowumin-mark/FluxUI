# Escape Hatch 策略

## 目标

本文记录 Phase 7.2 的 `FromWidget` escape hatch 策略。目标是让 legacy `Widget` 和 React-style `Element` 可以长期安全共存，而不是要求一次性迁移。

## 长期定位

- `FromWidget` 是长期保留的 Widget -> Element 桥接能力。
- `FromWidget` 不是临时 hack，也不是未来必须删除的内部 API。
- Element wrappers 可以继续使用 `FromWidget` 包装 legacy widget，直到对应控件拥有真正 host-state 实现。
- `RenderElement` 是 Element -> Widget 的运行时桥接，主要用于运行时、测试和兼容层。

## 混用边界

| 场景 | 推荐方式 | 说明 |
| --- | --- | --- |
| 新 Element 页面需要旧控件 | `FromWidget(oldWidget)` | 保留旧控件行为。 |
| 新 Element API 暴露旧控件 | wrapper 内部调用 `FromWidget` | 用户看到 Element 表面，底层可继续 legacy。 |
| 旧 Widget root 逐步引入 Element | `RenderElement(element)` 或专用 wrapper | 主要用于兼容层和过渡期。 |
| 旧项目不迁移 root | 保持 `Run` / `Widget` | 无需强制切到 `RunElement`。 |
| 新项目保留部分旧控件 | `RunElement` + `FromWidget` | 推荐的渐进迁移路径。 |

## 状态和生命周期

- legacy widget state 仍归 legacy widget / ref / command queue 所有。
- component state 仍归 HookSlot 所有。
- host state 未来归 host fiber 所有。
- `FromWidget` 不把 legacy widget state 搬进 HookSlot。
- `FromWidget` 不改变 Gio 每帧 layout 模型。
- 条件渲染、列表、路由等需要稳定复用时，外层继续使用 `Key`。

## 推荐实践

- 新 API 优先提供 Element wrapper，减少用户手写 `FromWidget`。
- 暂未迁移的复杂控件先通过 `FromWidget` 接入 Element 树。
- 当控件完成 host-state 化后，再逐步减少 wrapper 内部 legacy widget 依赖。
- 不删除、弱化或 deprecated `FromWidget`。

## 非目标

- 不删除 `FromWidget`。
- 不强制旧项目迁移到 `RunElement`。
- 不在业务层推荐频繁手写 `RenderElement`。
- 不把 legacy widget state 和 component HookSlot state 合并。
