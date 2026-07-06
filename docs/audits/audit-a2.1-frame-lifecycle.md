# A2.1 Frame 生命周期图

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 15:04:57 +08:00
- 负责人：Codex
- 关注：Runtime
- 输入命令：
  - `git status --short`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "func \\(r \\*Runtime\\) (begin|end).*Frame|func \\(r \\*Runtime\\) Register|func \\(s \\*HookStore\\) (BeginFrame|EndFrame)" internal`
  - `rg -n "BeginFrame\\(|EndFrame\\(|\\.Frame\\(gtx\\)|NewContext\\(" .`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/runtime.go`
  - `internal/frame.go`
  - `internal/context.go`
  - `internal/events.go`
  - `internal/interaction.go`
  - `internal/render_cache.go`
  - `internal/perf.go`
  - `internal/hook_store.go`
  - `internal/path_memory.go`
  - `app/app.go`
- 关联能力：
  - Runtime frame begin/end
  - Context 根作用域和子作用域
  - Event target/listener/focus registry
  - path memory 和持久状态清扫
  - render cache 活跃性清扫
  - redraw reason 和 frame diagnostics

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main` |
| `git status --short` | 无输出 |
| 判断 | 执行 A2.1 前工作区干净；本任务只新增 `audit-a2.1-frame-lifecycle.md` 并更新索引。 |

## Frame 生命周期图

当前真实应用入口在 `app/app.go` 的 `FrameEvent` 分支中组织每一帧：

```text
Gio FrameEvent
  -> gioApp.NewContext(&ops, evt)
  -> Runtime.BeginFrame()
       -> beginInteractionFrame()
       -> beginPerfFrame()
       -> beginRenderCacheFrame()
       -> beginEventFrame()
       -> HookStore.BeginFrame()
       -> clear activeMem / activeFx / pendingFx
       -> swap+clear hookCounts / hookCountIDs
       -> windowDragAreaActive = false
  -> Runtime.Frame(gtx)
       -> NewContext(gtx, runtime)
       -> RegisterEventTarget(rootPathID, 0)
       -> root Context carries Gtx, viewport, theme defaults, pathID=root
  -> build phase
       -> a.Root(buildCtx)
  -> layout + event registration + paint op recording
       -> root.Layout(treeCtx.Child(0))
       -> Context.Child / Scope allocate stable PathID and register event target
       -> Persistent / Memo mark MemoryKey active
       -> UseEffectKey marks effect active and queues post-frame setup
       -> RegisterEventListener / RegisterFocusTarget / RegisterShortcut rebuild registries
       -> widgets record Gio ops for layout, input, paint, cursor/window actions
  -> Runtime.EndFrame()
       -> endEventFrame()
       -> HookStore.EndFrame()
       -> validate hook count consistency
       -> cleanup unmounted runtime effects
       -> run pending runtime effects
       -> sweepInactiveMemory()
       -> endRenderCacheFrame()
       -> trackingMem = false
       -> endInteractionFrame()
       -> endPerfFrame()
  -> window drag native action router update
  -> evt.Frame(gtx.Ops)
```

注意：FluxUI 的 `Runtime.EndFrame()` 发生在 Gio `evt.Frame(gtx.Ops)` 之前。也就是说，FluxUI 的 frame end 是运行时登记、清扫和 effect 执行边界；真正把已记录的 Gio ops 交给后端绘制是在 `evt.Frame(gtx.Ops)`。

## 每帧重建状态

| 状态 | 重建位置 | 当前规则 |
| --- | --- | --- |
| 根 `Context` | `Runtime.Frame` / `NewContext` | 每个 Gio frame 创建新根上下文，包含本帧 `Gtx`、viewport、theme 默认前景色、root `PathID`。 |
| 子 `Context` scope | `Context.Child` / `Context.Scope` | 每次 layout 期间复制上下文并重置 `hookIndex`，根据 parent+index/name 获取稳定 `PathID`。 |
| event target registry | `beginEventFrame` | `targets` 每帧清空后重建，root target 在 `beginEventFrame` 和 `NewContext` 都会确保注册。 |
| event listener registry | `beginEventFrame` | `listeners` 每帧清空，组件 layout 期间用 `RegisterEventListener` 重新登记。 |
| focus target registry / tab order | `beginEventFrame` | `focusTargets`、`focusOrder`、`nextFocusOrder` 每帧重建；当前 focused target 值不在 begin 阶段清空。 |
| shortcut registry | `beginEventFrame` | `shortcuts` 每帧清空，局部快捷键由当前渲染树重新登记。 |
| hook active set | `HookStore.BeginFrame` | `active` 每帧清空，实际实例在组件渲染时重新标记 active。 |
| runtime memory active set | `BeginFrame` | `activeMem` 清空，`trackingMem=true` 后由 `PersistentKey` / `memoryValueKey` 重新标记。 |
| effect active set / pending effect queue | `BeginFrame` | `activeFx` 清空，`pendingFx` 截断；本帧 `UseEffectKey` 重新标记并按 deps 决定是否排队。 |
| render cache active sets | `beginRenderCacheFrame` | `activeText`、`activePaint`、`activeTree` 每帧清空，命中的 cache key 本帧重新标记。 |
| interaction frame stats | `beginInteractionFrame` | frame 序号递增，上一轮 coalesced pointer move 数进入 `last`，`activeHover`/`hasPointer` 重置。 |
| perf current frame | `beginPerfFrame` | diagnostics 开启时新建 `current FrameStats`，搬运 pending redraw reasons。 |
| window drag area flag | `BeginFrame` | `windowDragAreaActive=false`，本帧有窗口拖动区域时重新登记。 |

## 跨帧保留状态

| 状态 | 保留位置 | 当前规则 |
| --- | --- | --- |
| `Runtime` 本体 | `NewRuntime` 后跨帧持有 | theme、material theme、window controller、invalidator、path table、memory/effects/cache/diagnostics 子系统都挂在 Runtime。 |
| PathID table | `pathIDs` / `pathDebug` / `nextPathID` | 不在 `BeginFrame` 清空；相同 parent+child index 或 parent+scope name 复用同一 `PathID`。 |
| 持久 memory | `memory` | 跨帧保留；只有在 `EndFrame.sweepInactiveMemory()` 中删除本帧未被标记 active 的 key。 |
| runtime effects | `effects` | 跨帧保留；本帧未 active 的 effect 在 `EndFrame` cleanup 并删除，active effect 按 deps 决定是否重跑。 |
| HookStore instances | `hookStore.instances` | 跨帧保留；`EndFrame` 删除本帧未 active 的 component instance 并执行 hook cleanup。 |
| render cache entries | `render.text/staticPaint/staticTree` | 跨帧保留；`endRenderCacheFrame` 删除本帧未命中的 cache entry。 |
| focused target | `events.focusTarget` / `focusContext` | 不在 begin 阶段清空；`endEventFrame` 检查当前 focused target 是否仍注册且 focusable，否则触发 blur 到 0。 |
| pointer captures | `events.pointerCaptures` | 不在 `beginEventFrame` 清空；capture 通过显式 release/cancel 清除，A2.3 需要继续审查 stale target 风险。 |
| key down set | `events.keyDown` | 不在 `beginEventFrame` 清空；用于识别 repeat，keyup 时删除。 |
| hover target | `interact.hoverTarget` | 跨帧保留用于与本帧 `activeHover` 对比；变化时请求 redraw。 |
| redraw reasons | `perf.pendingReasons` | frame 外请求 redraw 时累积，`beginPerfFrame` 搬运到 current stats；frame 内新增也会写入 current。 |
| Gio `op.Ops` 缓冲 | `app` 层 `ops` | Gio `NewContext` 会准备本帧 ops；FluxUI layout/paint 期间写入，`evt.Frame(gtx.Ops)` 提交。 |

## Frame 阶段事实

| 阶段 | 主要入口 | 事实结论 |
| --- | --- | --- |
| frame begin | `Runtime.BeginFrame` | Runtime 子系统的活跃集合和本帧统计先重置，跨帧 map 本体通常保留，帧尾按 active 标记清扫。 |
| layout | `root.Layout(treeCtx.Child(0))` | layout 是组件树真正执行阶段，Context scope、PathID、memory active、event target/listener/focus/shortcut 和 paint ops 都在这里生成。 |
| event registration | `RegisterEventTarget` / `RegisterEventListener` / `RegisterFocusTarget` / `RegisterShortcut` | target/listener/focus/shortcut 是每帧重建模型；只有 focus 当前值、keyDown、pointer capture 等输入状态跨帧保留。 |
| paint | widget layout 内 Gio ops | 当前没有独立 `Runtime.PaintFrame`；绘制操作由 widget 在 layout 中写入 `gtx.Ops`，FluxUI `EndFrame` 后由 Gio `evt.Frame(gtx.Ops)` 提交。 |
| frame end | `Runtime.EndFrame` | 先修正失效 focus，再清理 hook/effect/memory/render cache，最后更新 interaction/perf stats；这是 runtime 的清扫和一致性检查边界。 |

## 风险

- `Runtime.EndFrame()` 在 `evt.Frame(gtx.Ops)` 之前执行；如果后续 effect 或 cleanup 直接依赖“画面已经提交”，需要重新定义边界，否则会误判 effect 时机。
- `events.pointerCaptures` 不随每帧清空，且本任务未看到 `endEventFrame` 对 capture owner 是否仍 active 的统一校验；A2.3 需要重点审查 stale capture 是否可能残留。
- `events.keyDown` 跨帧保留依赖 keyup 清理；窗口失焦、隐藏或平台丢 keyup 时是否清空，需要 A7 键盘审查确认。
- `PathID` 由 parent+index/name 稳定复用，但未在本任务深入列表重排、条件渲染、portal 路径规则；这会影响 memory 和 event path 正确性，应由 A2.2 继续审查。
- layout、event registration、paint op recording 目前混在 widget `Layout` 执行过程中；后续修复命中区域或 paint 顺序时，不能假设存在独立的 paint 阶段钩子。
- `NewContext` 和 `beginEventFrame` 都会注册 root target，当前行为幂等；如果未来改动注册逻辑，需保留 root target 一定存在的约束。

## 事实结论

- 每个 Gio `FrameEvent` 的主路径是 `BeginFrame -> Frame/NewContext -> build -> Layout -> EndFrame -> evt.Frame`。
- `Runtime.Frame` 本身只把 Gio `layout.Context` 包装为 FluxUI 根 `Context`，不负责清扫或提交；清扫由 `BeginFrame` / `EndFrame` 完成。
- FluxUI 事件目标、监听器、focus target、shortcut 是每帧重建模型；当前 focus target、pointer capture、keyDown 是跨帧输入状态。
- `Persistent` / `Memo` / runtime effects / HookStore / render cache 都是跨帧保留、本帧 active 标记、帧尾清扫的模型。
- `PathID` 表跨帧保留，是 memory、hook count、event path、debug path 的基础身份来源。
- paint 没有独立 runtime 阶段；widget layout 时记录 Gio ops，Gio `evt.Frame` 在 FluxUI `EndFrame` 后提交这些 ops。

## 验收

- 已记录 frame begin、layout、event registration、paint、frame end 的生命周期图。
- 已明确每帧重建的状态：Context、event registries、focus targets/tab order、shortcuts、active memory/effect/cache 标记、hook active 标记、interaction/perf current frame、window drag flag。
- 已明确跨帧保留的状态：Runtime 本体、PathID table、memory/effects、HookStore instances、render cache entries、focused target、pointer captures、keyDown、hover target、pending redraw reasons。
- 已标出 `pointerCaptures`、`keyDown`、PathID/list reorder、effect 时机和 layout/paint 混合边界作为后续风险。

## 后续依赖

- A2.2 PathID 和状态保存审查需要基于本任务的 PathID 保留规则，继续确认列表重排、条件渲染、portal/overlay 下状态是否错配。
- A2.3 Runtime registry 审查需要继续检查 event listener、focus target、pointer target、scroll target、pointer capture 的创建、更新、清理规则。
- A2.4 redraw 和 invalidation 审查需要继续展开 `RequestRedrawReason`、`RequestFrameRedrawReason`、interaction redraw 和 animation tick 的来源。
- A5.2 EventTarget 分发顺序审查需要基于每帧 event target/listener 重建事实，继续确认 capture/target/bubble/default action 顺序。
- A6.4 滚动后命中刷新审查需要确认滚动 offset 更新后，下一帧 target registry 和 Gio hit-test 是否能覆盖“不移动鼠标也能点击新目标”的验收。
