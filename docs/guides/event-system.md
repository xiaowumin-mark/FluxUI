<!-- fluxui-doc-meta
{
  "id": "event_system",
  "title": "Event System 高级事件系统",
  "category": "使用指南",
  "order": 132,
  "summary": "FluxUI Event System 提供浏览器风格的 target/path、capture/target/bubble、pointer、wheel、keyboard、focus、input、drag/drop 和自定义事件能力。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "PointerAreaElement(child Element, opts ...PointerAreaOption) Element",
    "PointerOnDown(fn func(ctx *Context, ev *PointerEvent), opts ...ListenerOption) PointerAreaOption",
    "PointerOnMove(fn func(ctx *Context, ev *PointerEvent), opts ...ListenerOption) PointerAreaOption",
    "PointerOnContextMenu(fn func(ctx *Context, ev *PointerEvent), opts ...ListenerOption) PointerAreaOption",
    "PointerOnWheel(fn func(ctx *Context, ev *WheelEvent), opts ...ListenerOption) PointerAreaOption",
    "KeyboardScopeElement(child Element, opts ...KeyboardScopeOption) Element",
    "ShortcutOn(spec Shortcut, fn func(ctx *Context, ev *KeyboardEvent), opts ...ListenerOption) KeyboardScopeOption",
    "InputOnBeforeInput(fn func(ctx *Context, event *InputEvent)) InputOption",
    "InputOnInputEvent(fn func(ctx *Context, event *InputEvent)) InputOption",
    "InputOnSubmit(fn func(ctx *Context, event *InputEvent)) InputOption",
    "DropTargetElement(child Element, onDrop func(ctx *Context, event DropEvent), opts ...DropTargetOption) Element",
    "DragSourceElement(child Element, opts ...DragSourceOption) Element",
    "OnEvent(ctx *Context, eventType EventType, fn func(ctx *Context, ev *Event), opts ...ListenerOption)",
    "DispatchCustomEvent(ctx *Context, target TargetID, eventType EventType, detail any, opts ...CustomEventOption) bool",
    "Capture() ListenerOption",
    "Once() ListenerOption",
    "Passive() ListenerOption",
    "LogEvents(enabled bool) AppOption"
  ]
}
-->

# Event System 高级事件系统

FluxUI 的高级事件系统以浏览器事件模型为参照：组件在当前 frame 中形成 target tree，事件沿 `ComposedPath()` 经过 capture、target、bubble 三个阶段，监听器可以调用 `StopPropagation`、`StopImmediatePropagation` 和 `PreventDefault`。默认情况下旧 API 仍然可用，新的事件层用于需要更多输入数据和传播控制的场景。

## 什么时候用简单 API

优先使用 `OnClick(ctx)`、`OnHover(ctx, bool)`、`InputOnChange`、`ScrollOnChange`：

- 只关心“按钮被点了”“值变了”“滚动位置变了”。
- 不需要坐标、按钮、修饰键、事件阶段或父容器捕获。
- 不需要取消组件默认行为。
- 业务组件希望保持最小 API 面。

## 什么时候用高级事件

使用 Event System：

- 需要 pointer 坐标、button/buttons、modifiers、pointer id、wheel delta。
- 需要父容器捕获子组件点击，或阻止事件继续冒泡。
- 需要右键菜单、中键点击、双击、pointer capture、拖拽画布。
- 需要局部快捷键、组件树焦点、Escape/Enter/Space 默认行为控制。
- 需要 `beforeinput` 拦截文本输入，例如只允许数字、Enter 提交、IME 组合保护。
- 需要自定义事件在组件间表达“交互语义已经发生”，但不替代状态管理。

## Pointer 和 wheel

`PointerAreaElement` 可包住任意区域：

```go
ui.PointerAreaElement(
    canvas,
    ui.PointerCaptureOnPress(true),
    ui.PointerOnDown(func(ctx *ui.Context, ev *ui.PointerEvent) {
        ev.SetPointerCapture(ctx)
    }),
    ui.PointerOnMove(func(ctx *ui.Context, ev *ui.PointerEvent) {
        x, y := ev.Position.X, ev.Position.Y
        _ = ev.CoalescedSamples()
        _, _ = x, y
    }, ui.Passive()),
    ui.PointerOnContextMenu(func(ctx *ui.Context, ev *ui.PointerEvent) {
        ev.PreventDefault()
    }),
    ui.PointerOnWheel(func(ctx *ui.Context, ev *ui.WheelEvent) {
        ev.PreventDefault()
    }),
)
```

`pointermove` 和 `wheel` 是高频事件。只读日志或指标时可以使用 `Passive()`；需要阻止默认滚动时不要标记 passive。

## Keyboard 和局部快捷键

`KeyboardScopeElement` 建立局部键盘作用域。快捷键只在当前 focus target 位于 scope 内时触发：

```go
ui.KeyboardScopeElement(
    content,
    ui.KeyboardScopeFocusable(true),
    ui.KeyboardScopeAutoFocus(true),
    ui.ShortcutOn(ui.ShortcutKey("k", ui.Modifiers{Ctrl: true}), func(ctx *ui.Context, ev *ui.KeyboardEvent) {
        ev.PreventDefault()
    }),
    ui.KeyOnDown(func(ctx *ui.Context, ev *ui.KeyboardEvent) {
        if ev.Key == "Escape" {
            ev.StopPropagation()
        }
    }),
)
```

全局快捷键属于 system/global shortcut 领域；组件键盘事件只处理当前组件树内的 focus 和 scope。

## Text input

`InputOnBeforeInput` 是可取消入口，`InputOnInputEvent` 描述变更，`InputOnSubmit` 处理提交。当前 Gio 对 mutation-before 的支持不是完整浏览器级别，FluxUI 对部分文本变更使用 best-effort 回滚和推断。

```go
ui.OutlinedTextFieldElement(
    value,
    ui.InputOnBeforeInput(func(ctx *ui.Context, ev *ui.InputEvent) {
        if ev.InputType == ui.InputTypeInsertText && ev.Data == "-" {
            ev.PreventDefault()
        }
    }),
    ui.InputOnSubmit(func(ctx *ui.Context, ev *ui.InputEvent) {
        submit(ev.Value)
    }),
)
```

## Custom events

应用可以用自定义事件表达组件语义：

```go
ui.OnEvent(ctx, ui.EventType("app:item-commit"), func(ctx *ui.Context, ev *ui.Event) {
    // ev.Detail carries application payload.
}, ui.Capture())

allowed := ui.DispatchCustomEvent(
    ctx,
    ctx.PathID(),
    ui.EventType("app:item-commit"),
    map[string]string{"id": "42"},
    ui.CustomCancelable(true),
)
_ = allowed
```

自定义事件适合“已经发生的交互消息”，不适合作为持久状态存储或跨页面状态管理。

## Diagnostics

开发模式可以打开事件诊断：

```go
ui.RunElement(App,
    ui.EnablePerfDiagnostics(true),
    ui.LogRedrawReasons(true),
    ui.LogEvents(true),
)
```

`FrameStats.Events` 会记录 dispatch 数量、listener 调用次数、listener 总耗时、取消状态和最后一次事件摘要。`LogEvents(true)` 会向 diagnostics writer 输出事件类型、target、path、listener 耗时、`default_prevented`、传播停止状态和默认行为是否允许执行。

右侧示例完整展示 pointer/wheel、右键菜单、局部快捷键、富输入拦截、自定义事件和拖放事件流。
