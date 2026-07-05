# Event System 全量测试台

运行：

```sh
go run ./examples/event_system_testbench
```

这个 example 不接入 docs browser，专门用于手工测试 `event-system-roadmap` 的实现。界面文案为中文，事件 API 名保留英文，窗口启动时会开启：

- `ui.EnablePerfDiagnostics(true)`
- `ui.LogEvents(true)`

因此 GUI 内有业务事件日志，控制台也会输出事件路径、listener 耗时、取消状态和默认行为结果。

覆盖清单：

- P0：旧 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`，以及 `ctx.Gtx.Event` escape hatch 说明。
- P1：`EventTarget`、`TargetID`、`Type`、`Event`、`Phase`、capture/target/bubble、`StopPropagation`、`StopImmediatePropagation`、`PreventDefault`、`DefaultPrevented`、`Capture`、`Once`、`Passive`、synthetic dispatch。
- P2：`PointerEvent`、`WheelEvent`、按钮/修饰键/坐标/pointer id、`pointerdown/up/move/enter/leave/over/out/cancel`、`click`、`dblclick`、`auxclick`、`contextmenu`、`wheel`、pointer capture、coalesced move。
- P3：`FocusManager`、焦点请求/移动/清空、disabled/hidden focus target、`focus/blur/focusin/focusout`、`keydown/keyup`、repeat、key/code/modifiers、局部 `ShortcutOn`、Enter/Space 激活、Escape 停止传播、Tab 默认焦点移动和取消。
- P4：`beforeinput`、`input`、`change`、`submit`、程序化 `InputRef.SetText/Append/Clear/Focus/Blur`、输入来源、best-effort 拦截、`CompositionEvent` 的 synthetic `compositionstart/update/end`。
- P5：Button、Pressable、ClickArea、Checkbox、Switch、Radio、Select、ScrollView、Slider、Dialog、Popup、DragSource、DropTarget 的默认行为迁移与可取消验证。
- P6：自定义事件 `Detail` payload、`DispatchEvent` 到指定 target、`EventBoundary` stop/redirect、`EventPortal` owner path、`ActivationEvent`。
- P7：事件诊断日志、高频 pointer/wheel 观察入口、coalesced move 展示。
