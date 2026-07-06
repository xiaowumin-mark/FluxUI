# A2.3 Runtime registry 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 17:55:37 +08:00
- 负责人：Codex
- 关注：Runtime、Event
- 输入命令：
  - `git status --short`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "func \\(r \\*Runtime\\) BeginFrame|func \\(r \\*Runtime\\) EndFrame|func \\(r \\*Runtime\\) beginEventFrame|func \\(r \\*Runtime\\) endEventFrame|type runtimeEventState|RegisterEventTarget|RegisterEventListener|RegisterFocusTarget|RegisterShortcut|SetPointerCapture|ReleasePointerCapture|PointerCaptureTarget|DispatchEvent|eventPath|recordEventDispatch|beginPerfFrame|endPerfFrame|RecordViewport|QueuePointerMove|beginInteractionFrame|endInteractionFrame" internal/runtime.go internal/events.go internal/perf.go internal/interaction.go`
  - `rg -n "PointerArea|PointerCaptureOnPress|registerPointerAreaListeners|registerPointerArea\\(|processPointerAreaEvents|pointerAreaDispatchTarget|registerWheelTarget|processWheelEvents|RecordViewport|DispatchWheelEvent|event.Op|pointer.Filter" event/input.go widget/pointer_area.go widget/list_grid.go widget/pointer_area_test.go widget/list_grid_test.go event/event_test.go event/keyboard_test.go`
  - `go test ./event ./widget`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/runtime.go`
  - `internal/events.go`
  - `internal/context.go`
  - `internal/frame.go`
  - `internal/perf.go`
  - `internal/interaction.go`
  - `event/input.go`
  - `event/event_test.go`
  - `event/keyboard_test.go`
  - `widget/pointer_area.go`
  - `widget/pointer_area_test.go`
  - `widget/list_grid.go`
  - `widget/list_grid_test.go`
- 关联能力：
  - Runtime event registry
  - 每帧 target/listener/focus/shortcut 重建
  - 跨帧 focus、pointer capture、keyDown 保留
  - Gio wheel target 与 Flux event target 边界
  - diagnostics current/last/pending 状态

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main` |
| `git status --short` | 无输出 |
| 判断 | A2.3 执行前工作区干净；本任务只新增 `audit-a2.3-runtime-registry.md` 并更新索引，不修改 runtime 代码。 |

## Registry 生命周期图

```text
Runtime.BeginFrame()
  -> beginInteractionFrame()
       frame++，上一帧 pointerMoves 写入 last，activeHover 清零
  -> beginPerfFrame()
       frameSeq++，pending redraw reasons 复制到 current，并清空 pendingReasons
  -> beginRenderCacheFrame()
  -> beginEventFrame()
       初始化缺失 map
       clear(targets)
       clear(listeners)
       clear(focusTargets)
       clear(shortcuts)
       focusOrder = focusOrder[:0]
       nextFocusOrder = 0
       nextSeq = 0
       RegisterEventTarget(root, 0)
  -> HookStore.BeginFrame()
  -> 清空 activeMem、hook count、activeFx、pendingFx、windowDragAreaActive

layout/build 阶段
  -> Context.Child/Scope 注册 event target
  -> event.On / RegisterEventListener 注册本帧 listener
  -> RegisterFocusTarget 注册本帧 focus target 和 tab order
  -> RegisterShortcut 注册本帧 shortcut
  -> PointerArea/ScrollView 通过 Gio event.Op 注册 Gio 输入 tag
  -> Flux pointer/wheel/keyboard/focus dispatch 查询当前 registry

Runtime.EndFrame()
  -> endEventFrame()
       如果当前 focused target 本帧不存在、disabled 或 hidden，触发 blur 并清空 focus
  -> HookStore.EndFrame()
  -> 清理 inactive effects / memory / render cache
  -> endInteractionFrame()
       hoverTarget <- activeHover，必要时请求 hover redraw
  -> endPerfFrame()
       current 写入 last，结束 diagnostics frame
```

## 注册表清单

| Registry | 字段 | 创建 | 每帧更新 | 清理规则 | 跨帧保留 |
| --- | --- | --- | --- | --- | --- |
| event target | `events.targets map[PathID]eventTarget` | `beginEventFrame` 懒初始化 | `Context.Child`、`Context.Scope`、`RegisterEventTarget`、`DispatchEvent` 都会注册目标 | `beginEventFrame` 每帧 `clear`，随后只保留本帧重新注册目标 | 否 |
| event listener | `events.listeners map[PathID][]*eventListener` | `beginEventFrame` 懒初始化 | `RegisterEventListener` append 后按 priority/seq 排序 | `beginEventFrame` 每帧 `clear`；`Once` listener 本帧 dispatch 后标记并 prune | 否 |
| focus target | `events.focusTargets map[PathID]*focusTarget` | `beginEventFrame` 懒初始化 | `RegisterFocusTarget` 写入当前目标、parent、tabIndex、disabled、hidden、activate、context | `beginEventFrame` 每帧 `clear`；`endEventFrame` 对当前 focused target 做本帧有效性检查 | 当前 focused target ID 保留 |
| focus order | `events.focusOrder []PathID`、`nextFocusOrder` | `beginEventFrame` 初始化/复用 slice | `RegisterFocusTarget` 对新 tabStop 目标 append | `beginEventFrame` 置零长度并重置 order | 否 |
| shortcut listener | `events.shortcuts map[PathID][]*shortcutListener` | `beginEventFrame` 懒初始化 | `RegisterShortcut` 以当前 target 记录 listener | `beginEventFrame` 每帧 `clear`；`Once` shortcut 本帧 prune | 否 |
| pointer capture | `events.pointerCaptures map[uint64]PathID` | `beginEventFrame` 懒初始化 | `SetPointerCapture` 写入 pointerID -> target；`ReleasePointerCapture` 删除 | 不在 `beginEventFrame` 清空；依赖 release/cancel 或显式释放 | 是 |
| keyboard down | `events.keyDown map[string]struct{}` | `beginEventFrame` 懒初始化 | `recordKeyState` 在 keydown 写入、keyup 删除，并据此设置 repeat | 不在 `beginEventFrame` 清空 | 是 |
| diagnostics current | `perf.current FrameStats` | `beginPerfFrame` 创建新 frame stats | event dispatch、viewport、cache、section、interaction 持续写入 current | `endPerfFrame` clone 到 `perf.last` 后 currentFrameActive=false | `last` 保留 |
| redraw reasons | `perf.pendingReasons map[string]int` | `RecordRedrawReason` 懒初始化 | frame 外和 frame 内都可追加；frame 内同步写 current reason counts | `beginPerfFrame` 把 pending 复制到 current 后清空 | pending 跨到下一帧 |
| interaction hover | `interact.hoverTarget`、`activeHover` | runtime 初始化零值 | 组件 snapshot 将 hovered target 写入 activeHover | `beginInteractionFrame` 清 activeHover；`endInteractionFrame` 用 activeHover 更新 hoverTarget | hoverTarget 保留 |

## 创建、更新、清理规则

### Event target

`Context.NewContext`、`Context.Child(index)`、`Context.Scope(name)` 都会注册当前路径的 event target；`RegisterEventListener`、`RegisterFocusTarget`、`RegisterShortcut` 和 `DispatchEvent` 也会先确保 target 已注册。事件 parent 优先来自本帧 `events.targets`，没有注册项时回退到 `pathDebug` 中的结构路径，再回退到 root。

结论：target registry 是每帧重建表。上一帧 target 不会因为 map 残留继续存在，但历史 `PathID` 与 `pathDebug` 会保留，用于调试和 parent 回退。

### Event listener

listener 只在当前 frame 生效。`beginEventFrame` 清空 `events.listeners`；组件必须在本帧 layout 时重新调用 `event.On`、`OnPointer`、`OnWheel`、`OnFocus`、`OnKeyboard` 等注册。`Once` listener 只影响当前 frame 内的后续 dispatch，因为下一帧 listener 表本来会重建。

结论：没有上一帧 listener 无条件残留风险。风险来自组件没有重新渲染时 listener 消失，这是符合 frame registry 语义的行为。

### Focus target

`focusTargets` 和 `focusOrder` 每帧清空并重建，但 `focusTarget` 这个“当前焦点 ID”跨帧保留。`endEventFrame` 会检查当前 `focusTarget` 在本帧是否仍存在且 focusable；若目标未注册、disabled 或 hidden，会通过 `changeFocus(..., target, 0)` 分发 blur/focusout 并清空焦点。

结论：focus target 本身不会跨帧残留；当前焦点 ID 可以跨帧保留，但帧尾有有效性清理。

### Shortcut

shortcut listener 与普通 listener 一样每帧重建。`dispatchShortcuts` 会用当前事件 path 判断 focused target 是否在 listener scope 内，并按 path depth、priority、seq 排序。`keyDown` 状态跨帧保留，用于 key repeat 和 keyup 清理。

结论：shortcut registry 没有上一帧 listener 残留；`keyDown` 是有意跨帧输入状态，不属于目标 registry。

### Pointer target 和 capture

Flux runtime 没有独立的“pointer target 命中表”。命中由 Gio `event.Op` 和 `pointer.Filter` 在 widget 层完成，随后 `PointerArea` 把 Gio pointer event 转成 Flux pointer event 并 dispatch 到当前 `PathID`。如果 `pointerCaptures` 中存在 pointerID，`pointerAreaDispatchTarget` 会把 dispatch target 改成 capture owner。

`pointerCaptures` 不在 `beginEventFrame` 清空；`PointerArea`、`PointerEvent` helper、`Slider` 等调用 `SetPointerCapture` 写入，release/cancel 或显式 `ReleasePointerCapture` 删除。

结论：pointer capture 是有意跨帧状态，但当前没有看到帧尾按 `events.targets` 校验 capture owner 是否仍在本帧注册。若 owner 子树在按下后消失且没有收到 release/cancel，capture 可能继续指向历史 `PathID`。

### Scroll target

Runtime 没有独立 `scrollTargets` registry。`ScrollView` 的 wheel target 是 `scrollState.wheelTag`，每帧通过 Gio `event.Op` 注册到当前 ops；`processWheelEvents` 用 `pointer.Filter{Target: wheelTag, Kinds: pointer.Scroll}` 从 Gio router 取事件，再把 wheel dispatch 到当前 `ctx.PathID()`。`PointerArea` 的 wheel 也走同样的 Gio tag 到 Flux wheel dispatch。

结论：scroll target 的创建和清理主要由 Gio router 的 frame ops 决定；Flux runtime 只记录 wheel diagnostics 和 viewport diagnostics。后续审查不要把 scroll wheel tag 误判为 runtime registry。

### Diagnostics registry

`PerfDiagnostics` 不是 target registry，而是 frame stats 状态机：

- `beginPerfFrame` 创建新的 `FrameStats`，把 pending redraw reasons 复制到 current 并清空 pending。
- event dispatch 时 `recordEventDispatch` 记录 dispatch 数、listener 数、耗时、defaultPrevented、propagationStopped、last target/path。
- viewport、virtualized items、render cache、frame section 等在本帧写入 current。
- `endPerfFrame` 把 current clone 到 last，`LastFrameStats()` 返回 last。

结论：diagnostics 的 current 每帧重建，last/pending 按语义跨帧保留，不参与事件命中。

## 现有测试覆盖

| 场景 | 测试 | 覆盖结论 |
| --- | --- | --- |
| event diagnostics 路径、listener、取消状态 | `event.TestEventDiagnosticsRecordPathDurationAndCancellation` | 能记录 path、listener calls、defaultPrevented、propagationStopped |
| event diagnostics 分类 | `event.TestEventDiagnosticsClassifyPointerWheelKeyboardFocus` | pointer、wheel、keyboard、focus 分类计数进入 Events 和 Interaction |
| capture/target/bubble 顺序 | `event.TestDispatchOrderCaptureTargetBubble` | 当前 registry path 下分发顺序稳定 |
| once/passive/stop | `event.TestOnceListenerIsRemovedForCurrentFrame`、`TestPreventDefaultAndPassiveListener` 等 | listener 语义在当前 frame 内有效 |
| portal/boundary path | `event.TestPortalAndBoundaryEventPathRules` | portal parent 和 boundary redirect/stop 使用当前 registry |
| focus 移动和默认键盘行为 | `event.TestFocusMoveDispatchesFocusEventOrder`、`TestKeyDownCanCancelActivationAndFocusMove` | focus target 注册、blur/focus、Tab/Enter 默认动作可被取消 |
| shortcut scope | `event.TestLocalShortcutOnlyFiresInsideFocusedScope` | shortcut 用当前 focus path 判定 scope |
| pointer capture release | `widget.TestPointerAreaDispatchesPointerWheelAndSyntheticEvents` | press 后 capture，release 后释放 |
| wheel target 和 scroll 嵌套 | `widget.TestScrollViewWheelScrollsContent`、`TestNestedScrollPrefersInnerListViewUnderPointer`、`TestNestedScrollPrefersInnerScrollViewUnderPointer` | scroll target 由 Gio router 当前 frame ops 决定，嵌套优先级有回归测试 |

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| pointer capture owner 跨帧消失 | 高 | `pointerCaptures` 跨帧保留，当前没有帧尾按本帧 `events.targets` 清理 stale owner；如果 captured 子树 unmount 且未收到 release/cancel，后续 pointer event 可能 dispatch 到历史 `PathID`。 |
| `DispatchEvent` 对未注册历史 target 自动注册 | 中 | dispatch 会注册 target，并用 `pathDebug` 回退 parent。对程序化 dispatch 有兼容价值，但可能掩盖目标已不在当前 frame 的事实。 |
| `focusContext` 持有上一帧 Context | 中 | 当前 focused target 重新注册时会刷新 context；目标消失时 `endEventFrame` 清焦点。若外部在 frame 外操作 focus，需要确认不会长期保存过期 Context。 |
| `keyDown` 无窗口失焦清理 | 中 | `keyDown` 依赖 keyup 删除；若窗口失焦或系统吞掉 keyup，repeat 状态可能延续。当前 A2.3 未看到统一 window blur 清理入口。 |
| scroll target 不在 runtime registry | 低 | 这是设计边界，但审查和文档容易误判。scroll 的 stale 风险主要落在 Gio router frame ops 和 widget state，而非 `runtimeEventState`。 |
| diagnostics writer frame 内同步输出 | 低 | event log 在 dispatch 后写 writer；若 writer 慢会影响事件处理耗时。当前属于 diagnostics 开启后的可接受成本。 |

## 事实结论

- `Runtime.BeginFrame()` 会调用 `beginEventFrame()`，每帧清空并重建 `events.targets`、`events.listeners`、`events.focusTargets`、`events.shortcuts` 和 `focusOrder`。
- `events.pointerCaptures`、`events.focusTarget`、`events.focusContext`、`events.keyDown` 不在 `beginEventFrame` 清空，是有意跨帧状态。
- focus 的跨帧保留有帧尾校验：当前 focused target 若本帧未注册、disabled 或 hidden，`endEventFrame()` 会 blur 并清空。
- pointer capture 的跨帧保留没有同等帧尾校验；释放依赖 release/cancel 或显式 release。
- Flux runtime 没有独立 scroll target registry；`ScrollView` 和 `PointerArea` 的 wheel target 由 Gio `event.Op` 和 `pointer.Filter` 在当前 frame ops 中注册和命中。
- diagnostics current stats 每帧重建，last stats 和 pending redraw reasons 按复现/日志语义跨帧保留。
- 当前未发现普通 event listener、focus target、shortcut listener 的上一帧无条件残留风险；主要 stale 风险集中在 pointer capture 和程序化 dispatch 历史 target。

## 验收

- 已列出每帧 registry 的创建、更新、清理规则。
- 已区分每帧必须重建字段：event target、event listener、focus target、focus order、shortcut listener、diagnostics current。
- 已区分允许跨帧保留字段：current focus ID/context、pointer capture、keyDown、hoverTarget、diagnostics last/pending reasons。
- 已明确 scroll target 不是 runtime registry，而是 Gio router 当前 frame ops 的输入 tag。
- 已标出上一帧目标残留风险：普通 listener/focus/shortcut 无无条件残留；pointer capture 存在 stale owner 风险，需要后续修复或测试覆盖。
- `go test ./event ./widget` 通过；`git diff --check` 通过，仅保留 `project-audit-baseline.md` 下次 Git touch 的 LF/CRLF 换行提示。

## 后续依赖

- A2.4 redraw 和 invalidation 审查需要沿用 diagnostics pending/current/last 的结论，继续确认 redraw reason 的进入时机和 frame 内外差异。
- A5.x Event 审查需要基于本文件继续细化 dispatch、listener priority、once/passive、portal/boundary path 和程序化 dispatch 历史 target 的边界。
- A6.x Focus/Keyboard 审查需要继续验证窗口失焦、disabled/hidden、TabIndex 和 keyDown 清理策略。
- A7.x Pointer/Drag 审查需要重点验证 pointer capture owner unmount、capture transfer、cancel 丢失和 drag/drop 过程中的 target 稳定性。
- A12.3 diagnostics 能力审查需要基于本文件补齐 event path、target、defaultPrevented、redraw reason 的输出完整性与性能影响。
