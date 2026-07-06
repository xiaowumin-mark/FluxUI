# A5.3 旧事件桥接审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：事件系统审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A5\\.3|旧事件桥接|callback|Button|Pressable|ClickArea|Input|ScrollView" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type Button|func Button|OnClick|Pressable|ClickArea|InputOnChange|ScrollOnChange|OnHover|DispatchEvent|NewClick|ClickEvent|InputEvent|ScrollEvent" widget event internal ui -g "*.go"`
  - `rg -n --glob "*_test.go" "DispatchClickEvent|PreventDefault|DefaultPrevented|OnClick|ClickArea|Pressable|InputOnChange|OnBeforeInput|ScrollOnChange|WheelEvent|beforeinput|legacy changes" widget event internal`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/dispatcher.go`
  - `event/input.go`
  - `event/text.go`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/input.go`
  - `widget/list_grid.go`
  - `event/event_test.go`
  - `event/text_test.go`
  - `widget/input_event_test.go`
  - `widget/list_grid_test.go`
  - `widget/interactive_layout_test.go`
  - `widget/pointer_area_test.go`
- 关联能力：
  - 新 EventTarget dispatch 到旧 callback 的执行门控
  - `OnClick`、`InputOnChange`、`ScrollOnChange` 兼容语义
  - `preventDefault` 对旧 callback/default action 的影响
  - 程序命令、键盘激活、Gio 原始输入与旧 callback 的顺序

## 旧 callback 桥接顺序

### Button

```text
Button.Layout
  -> event.UseClickable(ctx)
  -> if disabled/loading:
       RegisterFocusTarget(FocusDisabled(true))
     else:
       RegisterFocusTarget(FocusActivate(func { DispatchClickEvent(ctx, nil) }))
  -> ButtonRef commands when enabled:
       DispatchClickEvent(ctx, nil)
  -> clickable.ClickedEvent loop when enabled:
       DispatchClickEvent(ctx, pointerClick)
  -> hover snapshot change:
       DispatchHover(ctx, hovering)          # legacy OnHover only
  -> LayoutButton(clickable.Handle(), ...)

event.Dispatcher.DispatchClickEvent
  -> synthesize PointerEvent when source == nil
  -> DispatchPointerEvent(ctx, ctx.PathID(), source)
  -> if allowed && legacy Click != nil:
       legacy OnClick(ctx)
```

结论：`Button` 的旧 `OnClick` 总是在新 `click`/pointer event dispatch 之后执行；`PreventDefault()` 可以阻止旧 `OnClick`。`disabled` 或 `loading` 时，键盘激活、Ref 命令和 pointer click 都不会进入旧回调。

### Pressable / ClickArea

```text
ClickArea(child, onClick, opts...)
  -> Pressable(child, onClick, opts...)

Pressable.Layout
  -> if child == nil: return
  -> event.UseClickable(ctx)
  -> RegisterFocusTarget(FocusActivate(func { DispatchClickEvent(ctx, nil) }))
  -> PressableRef/ClickAreaRef commands:
       DispatchClickEvent(ctx, nil)
  -> clickable.ClickedEvent loop:
       DispatchClickEvent(ctx, pointerClick)
  -> LayoutClickArea(clickable.Handle(), child)
```

结论：`ClickArea` 是 `Pressable` 的旧别名；二者共享同一桥接路径。与 `Button` 不同，`Pressable` 没有内置 disabled/loading gate，因此 keyboard activation、Ref 命令和 pointer click 只受 `click` 事件取消影响。

### Input / TextField

```text
Input.Layout
  -> registerInputEventListeners(ctx)
       InputOnBeforeInput -> event.OnInput(BeforeInput)
       InputOnInputEvent  -> event.OnInput(Input)
       InputOnSubmit      -> event.OnInput(Submit)
       InputOnChange      -> not registered as EventTarget listener
  -> InputRef commands before Gio editor update:
       commitProgrammaticInput
  -> Gio editor.Update loop:
       ChangeEvent -> commitUserInput
       SubmitEvent -> dispatchSubmit

commitProgrammaticInput
  -> beforeinput(programmatic, cancelable)
  -> if canceled: return
  -> editor.SetText(next)
  -> input
  -> change
  -> legacy InputOnChange(ctx, value)

commitUserInput
  -> Gio editor text has already changed
  -> beforeinput(user, cancelable, best-effort)
  -> if canceled: editor.SetText(previous); return
  -> input
  -> change
  -> legacy InputOnChange(ctx, value)

dispatchSubmit
  -> submit event
```

结论：旧 `InputOnChange` 只在 mutation 被接受后执行，并且在新 `input`、`change` 事件之后执行。`beforeinput.PreventDefault()` 可以阻止程序命令和用户输入继续进入 `input/change/onChange`；用户输入因为 Gio 先变更文本，取消时通过回滚 editor 文本实现。`submit` 不桥接到旧 `InputOnChange`。

### ScrollView

```text
ScrollView.Layout
  -> drain ScrollRef commands
  -> set List axis / auto-to-end
  -> processWheelEvents
       Gio pointer.Scroll
       -> WheelEventFromGio
       -> DispatchWheelEvent(ctx, ctx.PathID(), &wheel)
       -> if allowed: applyWheelDefault(ctx, state, &wheel)
  -> layoutContent / clamp offset / set list position
  -> draw scrollbar
  -> if Position.First or Position.Offset changed:
       legacy ScrollOnChange(ctx, x, y)
```

结论：`ScrollOnChange` 不是 EventTarget listener，也没有对应的新 `scroll` typed event 桥接；它是布局后根据 Gio list position 变化直接触发的旧 callback。wheel 输入会先发新 `wheel` 事件，只有未被 `PreventDefault()` 取消时才执行默认滚动并可能触发 `ScrollOnChange`。Ref 命令和 auto-to-end 造成的位置变化也可能在布局后触发 `ScrollOnChange`。

## 事实结论

1. `event.Dispatcher.DispatchClickEvent` 是 `Button`、`Pressable`、`ClickArea` 旧 `OnClick` 的主要桥接点；旧回调执行前必经 `DispatchPointerEvent`。
2. `DispatchClickEvent` 在 `source == nil` 时构造 synthetic `PointerEvent`：`Type=click`、`Target=ctx.PathID()`、`Bubbles=true`、`Cancelable=true`、`Trusted=true`、`PointerType=other`、`Button=none`。
3. `DispatchClickEvent` 只在 `DispatchPointerEvent` 返回 allowed 时调用旧 `Click` handler，因此 `click` listener 的 `PreventDefault()` 可以阻止旧 `OnClick`。
4. `Button` 对 disabled/loading 有额外保护：注册 disabled focus target，且不处理 Ref click commands、不读取 `clickable.ClickedEvent`，避免禁用态旧 `OnClick` 被触发。
5. `Button` 的 `OnHover` 当前不是新 EventTarget 桥接；它由 `clickable.HoverChangedWithSnapshot` 的变化检测直接调用旧 `Hover` handler。
6. `ClickArea` 仅作为 `Pressable` 的兼容别名保留；`PressableRef` 是 `ClickAreaRef` 的别名，命令桥接完全一致。
7. `Pressable` 没有 disabled/loading 配置；其旧 `onClick` 能否执行主要由 `click` 事件是否被取消决定。
8. `InputOnBeforeInput`、`InputOnInputEvent`、`InputOnSubmit` 会注册为新输入事件 listener；`InputOnChange` 保持旧 callback 形态，不注册到 EventTarget。
9. `inputWidget.dispatchInputEvent` 在正常 runtime 存在时只调用 `fluxevent.DispatchInputEvent` 并返回 allowed，避免同一个新 handler 被注册 listener 和 fallback 路径调用两次。
10. `inputWidget.dispatchInputEvent` 在没有 runtime 的兼容路径下，会直接调用对应的新输入 handler；随后 `finishInputMutation` 仍只调用一次旧 `InputOnChange`。
11. `InputRef.SetText/Append/Clear` 是程序来源 mutation：顺序为 `beforeinput -> SetText -> input -> change -> InputOnChange`；`beforeinput` 取消时不会执行 `SetText` 和旧 `InputOnChange`。
12. 用户编辑是 Gio editor 先产生变更，再由 FluxUI best-effort 分类：顺序为 `Gio ChangeEvent -> beforeinput -> input -> change -> InputOnChange`；`beforeinput` 取消时会回滚到 previous text，且不触发旧 `InputOnChange`。
13. `InputOnChange` 不受 `change` event 的 `preventDefault` 控制，因为 `event/text.go` 中 `Change` 默认不可取消；真正的 mutation gate 是 `BeforeInput`。
14. `ScrollView` 的 `wheel` default action 位于 widget 本地：`DispatchWheelEvent` 返回 allowed 后才调用 `applyWheelDefault` 修改 offset。
15. `ScrollOnChange` 是位置变化观察 callback：在 `layoutContent` 更新 `state.list.Position` 后，如果 `First/Offset` 与上次不同才调用。
16. 现有测试覆盖了点击桥接、beforeinput 取消、旧 input change 触发/不触发、wheel scroll 触发 `ScrollOnChange` 和嵌套滚动优先级。

## 风险

| 风险 | 级别 | 说明 | 建议归属 |
| --- | --- | --- | --- |
| click 桥接依赖调用方统一使用 `event.Dispatcher` | M | `Button`、`Pressable`、`ClickArea` 已走统一桥接；若后续组件绕过 `DispatchClickEvent` 直接调用旧 callback，可能漏掉 `PreventDefault` gate。 | A5.5 default action 可取消性矩阵、Widget test |
| `Pressable` 无 disabled gate | S | 这是当前 API 事实，不是缺陷；但后续若给 Pressable 增加 disabled 语义，需要同时约束 keyboard activation、Ref command、pointer click 三条入口。 | Widget API |
| `OnHover` 仍是旧直接 callback | S | hover 变化没有新 hover event 桥接，当前只保证 change-only 旧回调；若后续引入 `pointerenter/leave` 或 `hoverchange`，需避免重复调用旧 `OnHover`。 | A5 后续 pointer/hover 审查 |
| `InputOnChange` 只能被 `beforeinput` 间接阻止 | M | `change` event 不可取消，因此不能用 `change.PreventDefault()` 阻止旧 `InputOnChange`；文档需明确 mutation gate 是 `beforeinput`。 | A7 text input、Event API 文档 |
| Gio 用户输入取消是 best-effort 回滚 | M | Gio `ChangeEvent` 到达时 editor 已变更，FluxUI 通过 `editor.SetText(previous)` 回滚；复杂 IME/composition 场景仍需要后续专门审查。 | A7 text input |
| `ScrollOnChange` 没有新 `scroll` event 对应物 | M | wheel 可通过 `wheel.PreventDefault()` 阻止默认滚动，但 Ref/auto-to-end/scrollbar/list position 变化没有统一 EventTarget 可取消点。 | A6 scroll、A5.5 default action |
| `ScrollOnChange` 值是位置近似表达 | S | 当前使用 `first + offset/1024` 表示 x/y，适合兼容通知，但不是像素级 absolute scroll offset。 | A6.1 ScrollView/ListView offset |

## 验收

- 已能说明 `Button` 旧 `OnClick` 的桥接顺序：
  - focus activation / Ref command / pointer click
  - -> `DispatchClickEvent`
  - -> new cancelable `click`
  - -> 未取消时旧 `OnClick`
- 已能说明 `Pressable` 和 `ClickArea` 的桥接顺序：
  - `ClickArea` 调用 `Pressable`
  - keyboard activation / Ref command / pointer click
  - -> `DispatchClickEvent`
  - -> 未取消时旧 `onClick`
- 已能说明 `Input` 的桥接顺序：
  - `beforeinput`
  - -> mutation commit 或用户输入回滚
  - -> `input`
  - -> `change`
  - -> 旧 `InputOnChange`
- 已能说明 `ScrollView` 的桥接顺序：
  - Gio wheel
  - -> new `wheel`
  - -> 未取消时默认滚动
  - -> layout 后位置变化触发旧 `ScrollOnChange`
- 已标出默认行为不应执行两次的保护点：
  - `DispatchClickEvent` 只在 allowed 时调用旧 click。
  - `inputWidget.dispatchInputEvent` 在 runtime 存在时不走 fallback 直接 handler。
  - `finishInputMutation` 每次 mutation 只调用一次旧 `InputOnChange`。
  - `ScrollOnChange` 只在 `First/Offset` 变化时调用。
- 已标出回调漏执行边界：
  - `Button` disabled/loading 不触发旧 click 属于预期。
  - `beforeinput` 取消后不触发旧 `InputOnChange` 属于预期。
  - `ScrollOnChange` 无位置变化不触发属于预期。

## 后续依赖

- A5.5 default action 可取消性矩阵应复用本文的 click、beforeinput、wheel allowed gate。
- A6.1 ScrollView 和 ListView offset 审查应继续确认 `ScrollOnChange` 的 offset 表达和 Ref/auto-to-end 触发边界。
- A7 text input 审查应继续验证 Gio `ChangeEvent`、`SubmitEvent`、IME/composition 与旧 `InputOnChange` 的关系。
- A8 Ref 审查应复用本文中 ButtonRef、PressableRef、InputRef、ScrollRef 命令生效顺序。
- A9 文档/API 审查应明确 `OnClick` 可被 `click.PreventDefault()` 阻止，而 `InputOnChange` 的取消入口是 `beforeinput`。
