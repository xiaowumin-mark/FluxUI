# A1.4 旧 API 兼容边界

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 14:59:45 +08:00
- 负责人：Codex
- 关注：Widget、Event
- 输入命令：
  - `git status --short --untracked-files=all`
  - `rg -n "A1\\.4|旧 API|OnClick|OnHover|InputOnChange|ScrollOnChange|Ref" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "func (OnClick|OnHover|InputOnChange|ScrollOnChange|.*AttachRef|New.*Ref)|type .*Ref|OnClick\\(|OnHover\\(|InputOnChange\\(|ScrollOnChange\\(" ui widget docs examples -g "*.go" -g "*.md"`
  - `rg -n "func New.*Ref|func \\(r \\*.*Ref|type .*Ref|func ScrollTo|func \\(r \\*ScrollRef" widget/component_refs.go widget/scroll_ref.go`
  - `rg -n "DispatchClickEvent|DispatchHover|type ClickHandler|type HoverHandler|PreventDefault|DefaultPrevented" event internal widget -g "*.go"`
  - `gopls go_package_api` for `github.com/xiaowumin-mark/FluxUI/ui`、`github.com/xiaowumin-mark/FluxUI/widget`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `docs/event-system-roadmap.md`
  - `docs/guides/event-system.md`
  - `docs/widgets/button.md`
  - `docs/widgets/textfield.md`
  - `docs/widgets/scroll_view.md`
  - `docs/widgets/click_area.md`
  - `docs/widgets/pressable.md`
  - `ui/ui.go`
  - `ui/extended_types.go`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/input.go`
  - `widget/list_grid.go`
  - `widget/component_refs.go`
  - `widget/scroll_ref.go`
  - `event/pointer.go`
  - `event/dispatcher.go`
  - `event/event_test.go`
  - `widget/input_event_test.go`
  - `examples/event_system_testbench/README.md`
- 关联能力：
  - 旧 callback 兼容边界
  - 新事件层兼容桥接
  - Widget Ref 命令边界
  - 后续默认行为和 Ref 生命周期审查输入

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | `M docs/audits/project-audit-baseline.md`；`?? docs/audits/audit-a0.1-workspace-freeze.md`；`?? docs/audits/audit-a0.2-version-deps.md`；`?? docs/audits/audit-a0.3-core-tests.md`；`?? docs/audits/audit-a0.4-example-smoke.md`；`?? docs/audits/audit-a1.1-dependency-graph.md`；`?? docs/audits/audit-a1.2-public-api-ownership.md`；`?? docs/audits/audit-a1.3-escape-hatch.md` |
| 判断 | 当前脏文件均为已完成审计产物或索引文件，允许继续保留；本任务只新增 `audit-a1.4-legacy-api-compatibility.md` 并更新索引。 |

## 旧 API 兼容矩阵

| API 面 | 公开签名 / 入口 | 语义所有者 | 当前桥接行为 | 兼容承诺 |
| --- | --- | --- | --- | --- |
| Button click | `ui.OnClick(fn func(ctx *ui.Context)) ButtonOption`；`widget.OnClick(fn event.ClickHandler) ButtonOption` | `widget` 维护按钮默认行为；`event` 维护 click 分发和 `PreventDefault` 语义；`ui` 只透传 | `widget.Button` 使用 `event.Dispatcher.DispatchClickEvent`，先派发 cancelable `click`，未被取消时调用旧 `OnClick(ctx)` | 不删除、不改签名；后续只能补 payload 或修正 default action，不能要求旧调用方接收事件参数 |
| Button hover | `ui.OnHover(fn func(ctx *ui.Context, hovering bool)) ButtonOption`；`widget.OnHover(fn event.HoverHandler) ButtonOption` | `widget` / `event` | `DispatchHover(ctx, hovering)` 直接调用旧 hover 回调；只表达 hover bool，不公开坐标、按钮、事件阶段 | 不删除、不改签名；不要把旧 `OnHover` 改成 pointer event handler |
| Input change | `ui.InputOnChange(fn func(ctx *ui.Context, value string)) InputOption`；`widget.InputOnChange(fn func(ctx *internal.Context, value string)) InputOption` | `widget.Input` 维护 editor/default 行为；`event` 维护 `beforeinput/input/change/submit` typed event | 用户输入先走 `beforeinput`；未取消时继续分发 `input/change`，最后调用旧 `InputOnChange(ctx, value)`；`InputRef.SetText/Append/Clear` 标记 programmatic source 并继续触发旧回调 | 不删除、不改签名；程序化 ref 变更继续保持当前兼容回调 |
| Scroll change | `ui.ScrollOnChange(fn func(ctx *ui.Context, x, y float32)) ScrollOption`；`widget.ScrollOnChange(fn func(ctx *internal.Context, x, y float32)) ScrollOption` | `widget.ScrollView` | 表示合成滚动位置 `x/y`，不是 wheel delta；wheel 默认行为由 `WheelEvent` / `PreventDefault` 管理，滚动后再报告位置变化 | 不删除、不改签名；`x/y` 继续表示位置，不改成原始 wheel delta |
| Pressable / ClickArea | `ui.PressableElement(child, onClick func(ctx *Context), opts...)`；`widget.Pressable(child, onClick func(ctx *internal.Context), opts...)`；deprecated `ClickArea` | `widget` | `Pressable` 是无视觉样式点击区域；`ClickArea` 作为旧名称兼容；click、keyboard activate、ref click 统一走 click 兼容桥 | `ClickArea` 保留兼容；新代码推荐 `Pressable`，但不得移除旧入口 |
| 派生 click option | `ImageOnClick`、`IconOnClick`、`CardOnClick`、`ListItemOnClick`、`IconButtonOnClick`、`FloatingActionButtonOnClick`、`ChipOnClick` 等 | 各 `widget` 组件 | 公开层仍是 `func(ctx)` 的简单语义回调，部分组件复用 `ButtonRef` 或内部 button-style 分发 | 保持 `func(ctx)` 简单签名；后续高级事件作为增量能力，不替换这些 option |

## Ref 兼容矩阵

| Ref | 创建 / 绑定 | 公开命令 | 当前边界 |
| --- | --- | --- | --- |
| `ButtonRef` | `NewButtonRef()`、`ButtonAttachRef(ref)` | `Click()` | 下一帧由 Button 消费；disabled/loading 时不分发 click；`ui` 以 type alias 暴露同一类型 |
| `ClickAreaRef` / `PressableRef` | `NewClickAreaRef()`、`NewPressableRef()`、`ClickAreaAttachRef(ref)`、`PressableAttachRef(ref)` | `Click()` | `PressableRef` 是 `ClickAreaRef` alias；命令走 click 兼容桥 |
| `InputRef` | `NewInputRef()`、`InputAttachRef(ref)` | `SetText(value string)`、`Append(value string)`、`Clear()`、`Focus()`、`Blur()` | 命令在后续布局帧消费；文本变更继续触发旧 `InputOnChange`，并在 typed input event 中标记 programmatic source |
| `ScrollRef` | `NewScrollRef()`、`ScrollAttachRef(ref)` | `ScrollToStart()`、`ScrollToTop()`、`ScrollToEnd()`、`ScrollToBottom()`、`ScrollToOffset(offset int)`、`ScrollBy(delta float32)` | 命令由 ScrollView 在下一帧消费；`ScrollOnChange` 报告消费后的滚动位置 |
| `CheckboxRef` | `NewCheckboxRef()`、`CheckboxAttachRef(ref)` | `SetChecked(checked bool)`、`Toggle()` | 公开命令式布尔状态入口；后续 A9 需确认是否触发 onChange/default action 的完整矩阵 |
| `SwitchRef` | `NewSwitchRef()`、`SwitchAttachRef(ref)` | `SetChecked(checked bool)`、`Toggle()` | 同 Checkbox 族，当前为稳定公开 Ref |
| `SliderRef` | `NewSliderRef()`、`SliderAttachRef(ref)` | `SetValue(value float32)`、`StepBy(delta float32)` | 公开命令式 value 调整；后续 A9/A10 需复核 range slider、step、clamp 和 onChange 顺序 |
| `RadioGroupRef` | `NewRadioGroupRef()`、`RadioGroupAttachRef(ref)` | `SetValue(value string)` | 公开命令式单选值入口 |
| `SelectRef[T]` | `NewSelectRef[T]()`、`SelectAttachRef[T](ref)` | `SetValue(value T)`、`Open()`、`Close()`、`Toggle()` | 泛型 Ref；控制值与展开状态，后续需结合 overlay/menu 生命周期审查 |
| `TabsRef` | `NewTabsRef()`、`TabsAttachRef(ref)` | `SetActive(key string)` | 公开命令式 tab 切换 |
| `DialogRef` | `NewDialogRef()`、`DialogAttachRef(ref)` | `Open()`、`Close()`、`Toggle()` | 公开命令式 modal 状态入口；后续 A8 审查 portal/overlay 边界 |
| `PopupRef` | `NewPopupRef()`、`PopupAttachRef(ref)` | `Open()`、`Close()`、`Toggle()` | 同 Dialog 族，但内容完全由用户定义 |
| `BottomNavRef` | `NewBottomNavRef()`、`BottomNavAttachRef(ref)` | `SetActive(key string)` | 公开命令式导航项切换 |
| 媒体/卡片 click Ref | `ImageAttachRef(ref *ButtonRef)`、`IconAttachRef(ref *ButtonRef)`、`CardAttachRef(ref *ButtonRef)` | `ButtonRef.Click()` | 复用 `ButtonRef` click 语义，不单独定义媒体 Ref |

## 文档和测试证据

| 证据 | 结论 |
| --- | --- |
| `docs/event-system-roadmap.md` | 明确写入 `OnClick(ctx)`、`OnHover(ctx, bool)`、`InputOnChange`、`ScrollOnChange` 不删除、不改签名；旧 `OnClick` 由 cancelable `click` 事件桥接 |
| `docs/guides/event-system.md` | 将 `OnClick` / `OnHover` / `InputOnChange` / `ScrollOnChange` 定义为简单 API，适用于不需要坐标、按钮、修饰键和事件阶段的场景 |
| `event/event_test.go` | 覆盖 `PreventDefault` 后旧 legacy click handler 不执行的行为 |
| `widget/input_event_test.go` | 覆盖 `beforeinput` 取消和 input/change 行为；作为 `InputOnChange` 兼容链的测试基础 |
| `examples/event_system_testbench/README.md` | 明确 P0 手动入口覆盖旧 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange` 和 `InputRef` |
| `docs/widgets/*.md` | button、textfield、scroll_view、click_area、pressable 等文档列出旧 callback 和 Ref API，构成文档承诺 |

## 风险

- `ui` facade 和 `widget` 实现层同时暴露同名 option；后续修复如果只改一层，可能造成文档签名和实际语义漂移。
- `OnClick` 已有 cancelable 桥接，但其他组件族的 default action 可取消性仍需 A5.3/A5.5 逐项确认，不能从 Button 推断所有组件都已完整一致。
- `OnHover` 是简单状态回调，不是高级 pointer event；如果后续把它改造成事件参数，会破坏旧签名。
- `ScrollOnChange` 的 `x/y` 是位置语义，不是 wheel delta；后续 A6 修复 wheel、父子滚动和横向滚动时必须保持这一点。
- Ref 命令多为下一帧消费，组件卸载后的命令安全性、是否触发 onChange/default action、同帧外部 value 与 Ref command 覆盖顺序尚未在本任务定论，应留给 A9.1/A9.2。
- `ClickArea` 已被文档标记为旧名称/兼容路径；可以推荐 `Pressable`，但删除或改变旧入口会破坏现有示例和文档。

## 事实结论

- 旧 callback API 是正式兼容层，不是待清理临时 API。
- `OnClick(ctx)` 当前是新 `click` 事件层之上的兼容桥：未取消时调用旧回调，被 `PreventDefault` 时跳过旧回调。
- `InputOnChange(ctx, value)` 继续是 TextField/Input 的稳定 value 回写入口；高级 `beforeinput/input/change/submit` 是增量能力。
- `ScrollOnChange(ctx, x, y)` 继续表示滚动位置变化；高级 wheel event 不应改变它的含义。
- `ClickArea` 是旧 API 兼容名称；`Pressable` 是推荐新名称，但两者当前共享 Ref 和 click 语义。
- 各 Ref 是公开命令式 API，`ui` 包通过 type alias 暴露 `widget` 的同一类型；后续修复不得删除或改动已承诺签名。

## 验收

- 已记录 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange` 的兼容矩阵。
- 已记录当前公开 Ref 类型、创建函数、Attach option 和命令方法。
- 已标出旧 callback 与新事件层的桥接关系，以及 `ScrollOnChange` 不等于 wheel delta 的边界。
- 已明确后续修复不得删除或改变已承诺签名。

## 后续依赖

- A5.3 旧事件桥接审查需要基于本矩阵逐项确认 Button、Pressable、ClickArea、Input、ScrollView 是否存在默认行为重复执行或漏执行。
- A5.5 default action 可取消性矩阵需要确认 `PreventDefault` 对各组件旧回调和默认行为的实际影响。
- A6.1/A6.2 需要继续确认 `ScrollOnChange` offset 来源、单位、更新时机、父子滚动和 wheel delta 边界。
- A7.4/A7.5 需要继续确认 `InputRef` 程序化变更、用户输入、paste/delete/undo/redo、IME 和 submit 顺序。
- A9.1/A9.2 需要为每个 Ref 建立命令生命周期、卸载安全性、onChange/default action 触发矩阵。
