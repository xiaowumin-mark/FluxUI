# A1.3 escape hatch 边界审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 14:48:08 +08:00
- 负责人：Codex
- 关注：Runtime、Event
- 输入命令：
  - `git status --short --untracked-files=all`
  - `rg "ctx\\.Gtx|\\.Gtx|FromWidget|Gtx\\.Event|gtx\\.Event|pointer\\.Event|key\\.Event|io/input|gioui\\.org/io/(pointer|key|event|input|transfer|clipboard|system)" internal event ui widget layout style theme system examples -g "*.go"`
  - `rg "ctx\\.Gtx\\.Event|Gtx\\.Event|gtx\\.Event" internal event ui widget layout style theme system examples docs -g "*.go" -g "*.md"`
  - `rg "FromWidget" ui docs examples -g "*.go" -g "*.md"`
- 输入文件：
  - `internal/context.go`
  - `ui/element.go`
  - `widget/pointer_area.go`
  - `widget/input.go`
  - `widget/list_grid.go`
  - `widget/keyboard_scope.go`
  - `widget/drag_source.go`
  - `widget/drop_target.go`
  - `widget/slider.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/utils.go`
  - `event/input.go`
  - `event/keyboard.go`
  - `docs/escape-hatch-strategy.md`
  - `docs/event-system-roadmap.md`
  - `examples/event_system_testbench/README.md`
- 关联能力：
  - Gio context escape hatch
  - Gio 原始事件桥接
  - Widget 与 Element 兼容桥
  - 事件系统迁移边界

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A1.3 执行前工作区干净，没有待解释的代码或文档污染。 |

## escape hatch 使用清单

| 入口 | 使用位置 | 分类 | 边界结论 |
| --- | --- | --- | --- |
| `ui.FromWidget(w Widget) Element` | `ui/element.go`、`ui/extended_types.go`、`ui/*_test.go`、`docs/escape-hatch-strategy.md`、`docs/phase7-legacy-api-plan.md`、`examples/react_workspace`、`examples/docs_browser/runtime_hooks_demo.go` | 正式 API，长期保留 | Widget -> Element bridge。文档明确"不删除、不弱化、不 deprecated"，用于 legacy widget、第三方 widget 和 Element wrapper 渐进迁移。 |
| `ui.RenderElement(element)` | `docs/escape-hatch-strategy.md`、`ui` facade | 正式兼容桥，但不推荐业务频繁手写 | Element -> Widget runtime bridge。主要用于运行时、测试和兼容层；业务层应优先使用 `RunElement` 或 Element wrapper。 |
| `internal.Context.Gtx` / `ui.Context.Gtx` | `internal/context.go` 暴露 `Gtx gioLayout.Context`；大量 `internal/widget/layout` 实现读取 `ctx.Gtx.Ops`、`Constraints`、`Dp`、`Focused`、`Execute` | 正式低层能力暴露 | `Context.Gtx` 是现有 Widget API 的核心 Gio context 通道，不能短期移除。正常组件实现可使用它做 layout、paint、focus command、Gio widget update。 |
| 用户直接 `ctx.Gtx.Event(...)` | `docs/event-system-roadmap.md`、`examples/event_system_testbench` 文案 | 临时兼容 / 高级 escape hatch | 文档明确高级用户可临时直接使用，但推荐迁移到新事件层；直接读取会绕开 FluxUI capture/bubble、default action、diagnostics 和兼容桥接。 |
| 组件内部 `ctx.Gtx.Event(pointer.Filter/key.Filter/transfer.*Filter)` | `widget/pointer_area.go`、`widget/input.go`、`widget/list_grid.go`、`widget/keyboard_scope.go`、`widget/drag_source.go`、`widget/drop_target.go`、`widget/slider.go`、`widget/tabs_dialog_toast.go`、`widget/utils.go` | 正式内部实现细节 | widget 层需要消费 Gio 原始事件并转换为 FluxUI 事件或组件默认行为。该能力不等于推荐用户绕过 `event` 包。 |
| Gio `pointer.Event` / `key.Event` 转换 helpers | `event/input.go`、`event/keyboard.go` | 正式桥接 API | `PointerEventFromGio`、`WheelEventFromGio`、`KeyboardEventFromGio` 等由 `event` 维护 Gio -> FluxUI 字段映射。 |
| Gio `transfer` drag/drop 事件 | `widget/drag_source.go`、`widget/drop_target.go` | 正式内部实现细节 | DragSource/DropTarget 使用 Gio transfer API 作为底层后端；public drag/drop 语义归 `widget` 和 `event`，系统能力探测归 `system`。 |
| Gio `key.FocusCmd` / `ctx.Gtx.Focused` | `widget/input.go`、`internal/clickable.go` | 正式内部实现细节 | 用于当前 Gio focus interop。后续 A7 需确认与 FluxUI FocusManager 的顺序和边界。 |
| Gio `op/clip/paint/event.Op/pointer.PassOp` 直接操作 | `internal/render.go`、`internal/static_subtree.go`、多数 `widget/*.go` | 正式内部实现细节 | 这是 FluxUI 渲染和命中区域实现方式，不是用户 escape hatch；风险在于 hit area/layout area 是否扩大，留给 A3/A5/A6 审查。 |

## 直接 Gio event 使用点

| 文件 | Gio event/filter | 当前用途 | 边界 |
| --- | --- | --- | --- |
| `widget/pointer_area.go` | `ctx.Gtx.Event(pointer.Filter)` | PointerArea 收集 pointer/wheel/enter/leave/click 等原始输入并分发 FluxUI pointer event。 | 正式内部实现。 |
| `widget/input.go` | `editor.Update(ctx.Gtx)`、`ctx.Gtx.Event(pointer.Filter)`、`ctx.Gtx.Execute(key.FocusCmd)`、`ctx.Gtx.Focused(editor)` | Gio Editor 输入、程序化 focus/blur、外部点击失焦。 | 正式内部实现；A7 需继续审查 IME/submit/focus 顺序。 |
| `widget/list_grid.go` | `ctx.Gtx.Event(pointer.Filter)` | ScrollView/ListView/Grid wheel 和 scrollbar drag 处理。 | 正式内部实现；A6 需审查父子滚动和剩余 delta。 |
| `widget/keyboard_scope.go` | `ctx.Gtx.Event(key.Filter{})` | KeyboardScope 原始 key event 捕获并转成 FluxUI keyboard/shortcut 行为。 | 正式内部实现；A7 需审查 scope/focus path。 |
| `widget/drag_source.go` | `ctx.Gtx.Event(transfer.SourceFilter)`、`ctx.Gtx.Execute(transfer.OfferCmd)` | DragSource payload request/offer。 | 正式内部实现；A10.7 需审查 pointer/scroll conflict。 |
| `widget/drop_target.go` | `ctx.Gtx.Event(transfer.TargetFilter)` | DropTarget 接收 transfer payload。 | 正式内部实现；A10.7 继续审查。 |
| `widget/slider.go` | `ctx.Gtx.Event(pointer.Filter{...})` | Slider press/release/move/drag/cancel。 | 正式内部实现；A10.5 继续审查。 |
| `widget/tabs_dialog_toast.go` | `ctx.Gtx.Event(pointer.Filter{...})` | Dialog/Popup outside click 或遮罩点击相关探测。 | 正式内部实现；A8 需审查 outside click 规则。 |
| `widget/utils.go` | `ctx.Gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Press})` | decoration/interactive helper 的 pointer press 探测。 | 正式内部实现；A13 需审查 clickable 冗余。 |
| `widget/*_test.go`、`examples/*_test.go` | `pointer.Event`、`input.Router`、`key.Event` synthetic queue | 测试和 smoke 构造输入。 | 测试专用，不算生产 escape hatch。 |

## `FromWidget` 边界

- `FromWidget` 是长期正式 API，不是临时兼容。
- `FromWidget` 只把 legacy `Widget` 包成 Element host leaf，不把 legacy widget state 迁入 HookSlot。
- `FromWidget` 不改变 Gio 每帧 layout 模型；legacy widget state 仍归 legacy widget、Ref、command queue。
- 新 API 应优先提供 Element wrapper，减少用户手写 `FromWidget`；复杂或未迁移控件可继续通过 `FromWidget` 进入 Element 树。
- `FromWidget(nil)` 在既有文档中被标为安全 no-op 行为，后续兼容矩阵应保留。

## `ctx.Gtx` 边界

- `ctx.Gtx` 字段本身是正式低层能力，因为 `Widget.Layout(ctx *internal.Context)` 需要 Gio `layout.Context` 完成约束、绘制、输入注册和命令执行。
- `ctx.Gtx.Ops`、`ctx.Gtx.Constraints`、`ctx.Gtx.Dp`、`ctx.Gtx.Execute`、`ctx.Gtx.Focused` 在内部 widget/runtime 中属于正式实现方式。
- `ctx.Gtx.Event` 对用户是临时/高级 escape hatch；推荐使用 `event.On*`、`widget.PointerArea`、`KeyboardScope`、组件回调或 Ref API。
- 用户直接消费 `ctx.Gtx.Event` 的风险是绕过 FluxUI event path、capture/bubble、default action、diagnostics、compat bridge 和 hit-test 刷新策略。

## 正式 API / 临时兼容归类

| 能力 | 归类 |
| --- | --- |
| `FromWidget` | 正式 API，长期保留。 |
| `RenderElement` | 正式兼容桥，主要用于 runtime/test/compat，不鼓励业务频繁手写。 |
| `Context.Gtx` 字段 | 正式低层能力暴露。 |
| 内部 widget 使用 Gio event/filter/transfer/key command | 正式内部实现细节。 |
| 用户直接 `ctx.Gtx.Event(...)` | 临时兼容 / 高级 escape hatch。 |
| `event.*FromGio` conversion helpers | 正式桥接 API。 |
| 测试中 synthetic Gio events | 测试工具，不算生产 API。 |

## 风险

- `Context.Gtx` 是公开字段，用户可以直接绕过 FluxUI 事件系统；这会让 A5/A6/A7 中的 event path、default action、diagnostics 和 scroll/focus 规则难以保证。
- widget 内部存在多处直接 Gio event 消费点，后续修复事件系统时不能只改 `event` 包，还要核对组件默认行为是否同步读取 `PreventDefault` 或派发顺序。
- `FromWidget` 长期保留意味着 Element host-state 化不能破坏 legacy widget state、Ref 和 command queue 语义。
- `RenderElement` 作为反向桥若被业务层大量使用，可能形成 Element/Widget 状态所有权混杂；当前文档已将其限制为 runtime/test/compat 主要用途。

## 事实结论

- `FromWidget` 是正式长期 escape hatch，不是临时迁移 hack。
- `ctx.Gtx` 是正式低层 Gio context 暴露；但用户直接 `ctx.Gtx.Event` 只应视为临时/高级兼容入口。
- 生产代码中直接 Gio event 使用点主要集中在 `widget` 包，属于组件实现细节，用于把 Gio 原始输入转为 FluxUI 事件、组件状态和默认行为。
- `event` 包提供正式 Gio -> FluxUI event 字段转换 API，是常规事件桥接的所有权入口。

## 验收

- 已记录 `ctx.Gtx`、`FromWidget`、直接 Gio event 使用点。
- 已标出 `FromWidget` 为正式长期 API。
- 已标出用户直接 `ctx.Gtx.Event` 为临时兼容 / 高级 escape hatch。
- 已区分生产实现、用户 API、测试 synthetic event，避免把 widget 内部 Gio 消费误判为用户推荐路径。

## 后续依赖

- A1.4 旧 API 兼容矩阵应把 `FromWidget`、`Run`/`Widget`、旧 callback 与新事件层的兼容边界合并记录。
- A5 事件系统审查必须逐项核对 `widget` 内部 Gio event 使用点是否正确接入 capture/target/bubble/default action。
- A6/A7/A8 应分别继续审查 scroll、focus/keyboard/text input、overlay outside click 中的直接 Gio event 使用点。
