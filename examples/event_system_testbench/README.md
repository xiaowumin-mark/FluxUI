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

## P0-P7 手工验收

| 阶段 | 操作 | 预期结果 |
| --- | --- | --- |
| P0 | 点击旧 `OnClick` 按钮，悬停旧 hover 按钮，编辑输入框，滚动 P0 区域。 | GUI 日志显示旧回调触发，`ScrollOnChange` 上报真实 offset，escape hatch 说明不影响其它区域。 |
| P1 | 在 EventTarget 面板分别打开 PreventDefault、StopPropagation、Once、Passive，再点击 synthetic dispatch。 | 日志顺序能区分 capture/target/bubble；PreventDefault 只取消默认行为；StopPropagation 停止后续传播；Passive 取消失败可见。 |
| P2 | 在蓝色 pointer 面板移动、拖动、右键、双击和滚轮。 | 事件只来自面板视觉区域，capture/coalesced/wheel 坐标和取消状态可从日志判断。 |
| P3 | 聚焦测试块，按 Tab、Shift+Tab、Ctrl+K、Escape、Enter、Space，并切换阻止 Tab 默认。 | 局部 shortcut 只在 scope 内触发；Tab 默认可取消；Enter/Space 激活当前焦点块；Escape 不污染其它区域。 |
| P4 | 输入普通字符和数字，按 Enter，使用 InputRef SetText/Append/Clear/Focus/Blur，并触发 synthetic composition。 | beforeinput/input/change/submit/source 日志按顺序出现；被拦截输入不触发后续 change；composition 日志只代表 synthetic 链路。 |
| P5 | 关闭阻止开关逐项点击控件，再打开阻止 click/wheel 默认重复操作；打开 Dialog/Popup 并点击内部和遮罩。 | 未阻止时旧回调或状态变化执行；阻止后对应默认行为不执行；overlay 内部点击不误关闭，遮罩按规则关闭。 |
| P6 | 触发自定义事件、boundary stop/redirect、portal owner path 和 activation。 | detail payload 到达目标；普通冒泡、截断、重定向和 portal owner path 在日志中可区分。 |
| P7 | 在 P1/P2/P5 区域制造事件后查看 GUI 日志和控制台，再点击诊断标记。 | 能看到 event type、target/path、listener 耗时、default/cancel 结果，以及高频 pointer/wheel 观察入口。 |
