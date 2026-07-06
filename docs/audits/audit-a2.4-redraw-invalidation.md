# A2.4 redraw 和 invalidation 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 18:35:00 +08:00
- 负责人：Codex
- 关注：Runtime、Perf
- 输入命令：
  - `git status --short --branch`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "Invalidate|InvalidateOp|Redraw|redraw|RecordRedrawReason|pendingReasons|animation|Animation|Tick|QueuePointerMove|interaction|Request|request" internal event ui widget layout style theme system examples docs -g "*.go"`
  - `rg -n -C 8 "SetInvalidator|win\\.Invalidate\\(\\)|WindowHandle\\) Invalidate|RequestFrameRedrawReason|RecordRedrawReason" app\\app.go internal\\context.go internal\\perf.go`
  - `rg -n -C 3 'RequestRedrawReason\\("|RecordRedrawReason\\("|RequestFrameRedrawReason\\("|WindowInvalidate|Invalidate\\(\\)' internal state ui widget app examples -g '*.go'`
  - `rg -n "Redraw|redraw|Invalidate|invalidat|animation.running|state.Set|async.start|interval.tick|WindowInvalidate" internal state ui widget examples -g "*_test.go"`
  - `go test ./internal ./state ./ui ./widget ./router`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/runtime.go`
  - `internal/context.go`
  - `internal/perf.go`
  - `internal/interaction.go`
  - `internal/ripple.go`
  - `state/state.go`
  - `state/async.go`
  - `state/interval.go`
  - `ui/hooks.go`
  - `widget/utils.go`
  - `widget/list_grid.go`
  - `widget/progress.go`
  - `widget/tabs_dialog_toast.go`
  - `app/app.go`
  - `router/router.go`
  - `anim/animation.go`
- 关联能力：
  - Runtime invalidator 绑定
  - frame 内 `op.InvalidateCmd{}` 刷新
  - redraw reason pending/current 诊断
  - 动画 tick redraw
  - 状态变更 redraw
  - 输入交互 redraw
  - 程序命令 redraw

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 12 |
| `git status --short` | 无输出 |
| 判断 | A2.4 执行前工作区干净；本任务只新增 `audit-a2.4-redraw-invalidation.md` 并更新索引，不修改 runtime 代码。 |

## Redraw 机制总览

```text
外部/跨 goroutine 触发
  -> Runtime.RequestRedrawReason(reason)
       -> RecordRedrawReason(reason)
       -> Runtime.requestRedraw()
       -> app.Run 绑定的 invalidator
       -> gio app.Window.Invalidate()
       -> 后续 Gio FrameEvent

当前 frame 内触发
  -> Context.RequestFrameRedrawReason(reason)
       -> Gtx.Execute(op.InvalidateCmd{})
       -> Context.RequestRedrawReason(reason)
       -> Runtime.RequestRedrawReason(reason)
       -> 记录 reason 并调用 window invalidator
       -> 当前 frame ops 中也携带 Gio InvalidateCmd

只记录诊断，不主动触发
  -> Runtime.RecordRedrawReason(reason)
       -> 写入 perf pending/current reason
       -> 不调用 invalidator
```

`app.Application.Run` 中创建 `Runtime` 后调用 `rt.SetInvalidator(func(){ w.Invalidate() })`，因此 `Runtime.RequestRedrawReason` 是 Flux 层请求下一次窗口 frame 的主要入口。`Context.RequestFrameRedrawReason` 额外执行 `op.InvalidateCmd{}`，适合动画和 frame 内输入响应；Gio 会把该 command 纳入当前 frame ops。

## 触发来源清单

| 分类 | 入口 | 典型 reason | 是否直接触发 redraw | 诊断进入方式 | 结论 |
| --- | --- | --- | --- | --- | --- |
| 程序命令 | `Runtime.RequestRedraw()`、`Context.RequestRedraw()` | `RequestRedraw` | 是，调用 runtime invalidator | `RequestRedrawReason` 先记 reason 再 invalidator | 适合事件回调或 goroutine 中显式请求窗口刷新 |
| frame 内程序命令 | `Context.RequestFrameRedraw()` | `RequestFrameRedraw` | 是，执行 `op.InvalidateCmd{}` 并调用 invalidator | 当前 frame active 时同时写入 current 和 pending | 仅适合当前 layout/render frame 内使用 |
| 窗口命令 | `Context.WindowInvalidate()` / `ui.WindowInvalidate()` / `WindowHandle.Invalidate()` | `WindowInvalidate` | 是，由 window controller 调用 Gio `Window.Invalidate()` | `WindowInvalidate` 成功后只 `RecordRedrawReason` | 这是显式窗口控制 API，不等同于动画 tick |
| Gio 窗口事件 | `ConfigEvent`、`ViewEvent` | `window.event` | 事件本身来自 Gio；Flux 只记录 reason | `RecordRedrawReason` | 只用于诊断标注窗口配置/视图事件，不主动请求额外 redraw |
| 状态变更 | `state.State.Set` | `state.Set` | 是，调用 runtime invalidator | `RequestRedrawReason` | 业务状态变更会请求新 frame，可从任意 goroutine 调用 |
| 异步状态 | `AsyncHandle.Run`、完成回调、`Reset` | `async.start`、`async.complete`、`async.reset` | 是 | `RequestRedrawReason` | 异步开始和完成都会请求 redraw |
| interval tick | `state.UseInterval` | `interval.tick` | 是 | `RequestRedrawReason` | 定时器 tick 后调用用户函数，再请求 redraw |
| Widget ref / legacy command | `redrawInvalidator(ctx)` 绑定到各 Ref | `widget.redraw` | 是 | `RequestRedrawReason` | 用于旧 API 或组件命令式 ref 的兼容刷新 |
| 输入：hover | `Runtime.endInteractionFrame` | `pointer.hover_changed` | 是 | frame 尾调用 `RequestRedrawReason` | hover target 变化才请求 redraw |
| 输入：press/focus | `Runtime.ObserveInteractionSnapshot` | `pointer.pressed_changed`、`pointer.focus_changed` | 是 | 交互快照变化时 `RequestRedrawReason` | 初始快照不请求 redraw，变化才请求 |
| 输入：scroll wheel | `ScrollView` wheel 处理 | `input.scrollview.wheel` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | wheel 导致 offset 改变才请求 frame redraw |
| 输入：scrollbar drag | `ScrollView` scrollbar 处理 | `input.scrollbar` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | 拖动中会持续请求 frame redraw |
| 动画通用 | `anim.Animation.Value`、`ui.UseAnimatedValue`、`ui.UseAnimatedDecoration` | `animation.running` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | running 时持续 tick；到达目标后停止 |
| Material 动画 | widget motion helper | `animation.running` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | hover/press/focus/颜色/decoration 动画共用 reason |
| Ripple | `internal.drawRipplePress` | `animation.ripple` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | ripple 未结束时请求刷新 |
| Progress | indeterminate progress | `animation.progress` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | 不确定进度条按动画驱动刷新 |
| Toast | toast 进入/退出过程 | `animation.toast` | 是，frame 内 `op.InvalidateCmd{}` | `RequestFrameRedrawReason` | toast 生命周期动画期间请求刷新 |
| Router | 路由导航和过渡 | `router.navigate`、`animation.router` | 是 | 导航可走 `RequestRedrawReason`，过渡走 `RequestFrameRedrawReason` | 属于程序命令和动画混合来源 |

## Redraw reason 记录规则

`Runtime.RecordRedrawReason` 只在 perf diagnostics 启用且 reason 非空时写入统计；它不会调用 invalidator。`Runtime.RequestRedrawReason` 会先调用 `RecordRedrawReason`，再调用 `requestRedraw`。`Context.RequestFrameRedrawReason` 会先向 Gio context 执行 `op.InvalidateCmd{}`，再走 `RequestRedrawReason`。

`beginPerfFrame` 会把上一阶段累积的 `pendingReasons` 复制到当前 frame 的 `ReasonCounts`，随后清空 pending。若当前 frame active 时又记录了 reason，`addReasonLocked` 会同步写入当前 frame 的 `ReasonCounts`，同时保留到 pending，供后续 frame 复现“谁请求了下一帧”。因此 diagnostics 的 current/pending 具备跨 frame 语义，不能把同一 reason 简单理解成只属于当前 frame。

## 用户输入、动画、状态、程序命令区分

| 来源类别 | 判定标准 | 代表 reason | 说明 |
| --- | --- | --- | --- |
| 用户输入 | 输入事件改变交互或滚动状态 | `pointer.hover_changed`、`pointer.pressed_changed`、`pointer.focus_changed`、`input.scrollview.wheel`、`input.scrollbar` | 原始 pointer move 只作为 interaction 统计，未发生 hover/press/focus 或滚动状态变化时不作为业务 redraw reason |
| 动画 | 由 frame 时间推进，未到达目标或生命周期未结束 | `animation.running`、`animation.ripple`、`animation.progress`、`animation.toast`、`animation.router` | 主要走 `RequestFrameRedrawReason` 和 `op.InvalidateCmd{}` |
| 状态变更 | 业务状态、异步状态或定时器更新数据 | `state.Set`、`async.start`、`async.complete`、`async.reset`、`interval.tick` | 主要走 runtime invalidator，适合跨 goroutine |
| 程序命令 | 应用显式要求刷新或导航/窗口 API 改变外部状态 | `RequestRedraw`、`RequestFrameRedraw`、`WindowInvalidate`、`router.navigate`、`widget.redraw`、`window.event` | `window.event` 是诊断记录；其他多数会请求 redraw |

## 现有测试覆盖

| 场景 | 测试 | 覆盖结论 |
| --- | --- | --- |
| runtime invalidator | `internal.TestRequestRedrawCallsInvalidator`、`TestRequestRedrawNilInvalidator`、`TestRequestRedrawConcurrent` | 覆盖显式 runtime redraw、nil invalidator、并发调用 |
| captured context redraw | `internal.TestRequestRedrawFromCapturedFrameContextUsesRuntimeInvalidator` | 覆盖 frame context 被 goroutine 持有后仍可通过 runtime invalidator 请求 redraw |
| frame redraw nil safety | `internal.TestRequestFrameRedrawNilSafety` | 覆盖 nil context 下 frame redraw 不 panic |
| state redraw | `state.TestSetTriggersRedraw`、`TestSetWithoutInvalidator` | 覆盖 `State.Set` 触发 invalidator，且无 invalidator 时不 panic |
| interval redraw | `state.TestUseIntervalMountsOnceAndInvalidates` | 覆盖 interval tick 后 invalidator 被调用 |
| interaction redraw | `internal.TestInteractionPressedChangeRequestsRedraw`、`TestInteractionHoverTargetChangeRequestsRedraw` | 覆盖 pressed/hover 变化触发 redraw，初始快照不触发 pressed redraw |
| ripple invalidation | `internal/ripple_test.go` | 覆盖 ripple active/expired 状态下是否请求 invalidation |
| animation idle | `ui.TestAnimatedHooksDoNotRedrawWhenAlreadyAtTarget`、`examples/docs_browser.TestAnimationDemoSettlesWithoutIdleRedraw`、`examples/component_lab.TestComponentLabSettlesWithoutIdleRedraw`、`examples/material3_showcase.TestMaterial3ShowcaseSettlesWithoutIdleRedraw` | 覆盖动画稳定后不继续 idle redraw |
| perf reason smoke | `examples/component_lab.TestPerfScenario`、`examples/docs_browser.TestPerfScenario` | 覆盖 diagnostics 中能看到 redraw reasons |

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| `RequestFrameRedrawReason` 同时执行 `op.InvalidateCmd{}` 和 runtime invalidator | 中 | 这能兼容 frame 内动画和窗口 invalidator，但在 diagnostics 中可能让同一 reason 同时进入 current 与 pending；后续分析需按 frame active 语义解释。 |
| `RecordRedrawReason` 命名容易被误解为触发 redraw | 中 | `RecordRedrawReason` 只记诊断，不调用 invalidator；`window.event`、`WindowInvalidate` 的记录语义不同，后续文档需明确。 |
| `State.Set` 无值相等短路 | 中 | 即使设置相同值也会请求 redraw；这保持简单语义，但高频状态写入会造成不必要 redraw。 |
| Widget ref invalidator 统一使用 `widget.redraw` | 中 | 能保持旧 API 兼容，但 diagnostics 难以定位具体组件和命令来源。 |
| 原始 pointer move 不作为 redraw reason | 低 | 这是性能设计；若调试者只看 reason，会看不到 pointer move 本身，需要结合 interaction `pointer_moves` 统计。 |
| frame 内动画 reason 粒度偏粗 | 低 | 多数动画共用 `animation.running`，能区分动画类别但难以定位具体 hook/widget。 |
| `WindowHandle.Invalidate` 绕过 `Runtime.RecordRedrawReason` | 低 | 从 handle 直接调用时没有 Flux runtime reason；通过 `Context.WindowInvalidate` 才记录 `WindowInvalidate`。 |

## 事实结论

- FluxUI redraw 有三类核心入口：`Runtime.RequestRedrawReason` 触发 window invalidator，`Context.RequestFrameRedrawReason` 触发 Gio `op.InvalidateCmd{}` 并走 runtime invalidator，`Runtime.RecordRedrawReason` 只记录诊断。
- `app.Application.Run` 把 runtime invalidator 绑定到 Gio `Window.Invalidate()`，因此 runtime/state/async/interval/ref 等跨 goroutine 或程序化刷新最终由 Gio window 产生后续 `FrameEvent`。
- 动画 tick 主要走 `RequestFrameRedrawReason`，包括 `animation.running`、`animation.ripple`、`animation.progress`、`animation.toast`、`animation.router`。
- 用户输入 redraw 只在交互状态改变或滚动状态改变时触发；原始 pointer move 进入 interaction 统计，不直接作为 redraw reason。
- 状态变更 redraw 包括 `state.Set`、`async.start`、`async.complete`、`async.reset`、`interval.tick`，这些入口会调用 runtime invalidator。
- 程序命令 redraw 包括显式 `RequestRedraw`、`RequestFrameRedraw`、`WindowInvalidate`、`widget.redraw`、`router.navigate`；其中 `window.event` 是窗口事件诊断标记，不主动请求 redraw。
- diagnostics 的 reason 同时有 current 和 pending 语义；frame active 期间记录的 reason 会写当前 frame，也会保留给后续 frame 开头复制。

## 验收

- 已列出触发 redraw 的来源清单，并标注是否直接触发 invalidator 或 `op.InvalidateCmd{}`。
- 已区分用户输入、动画、状态变更、程序命令四类 redraw 来源。
- 已明确 `RecordRedrawReason` 与 `RequestRedrawReason` 的差异，避免把诊断记录误判为实际 redraw 触发。
- 已明确 raw pointer move 不作为业务 redraw reason，需要结合 interaction stats 观察。
- 已记录现有测试覆盖，后续修复能判断是否改变既有 redraw 基线。
- `go test ./internal ./state ./ui ./widget ./router` 通过。
- `git diff --check` 通过。

## 后续依赖

- A3.x state/hooks 审查需要沿用本文件对 `state.Set`、async、interval 的 redraw 语义，判断是否需要值相等短路或批处理策略。
- A5.x Event 审查需要结合本文件区分事件 dispatch 与输入状态变化；不是所有事件都会触发 redraw。
- A7.x Pointer/Scroll 审查需要继续验证 pointer move coalescing、scrollbar dragging 和 wheel offset 改变的 redraw 粒度。
- A8.x Animation 审查需要基于 `RequestFrameRedrawReason` 和 `PerfAnimation` 统计继续细化动画 idle 停止条件。
- A12.3 diagnostics 审查需要补齐 reason 来源定位能力，特别是 `widget.redraw` 和 `animation.running` 的具体调用方可观测性。
