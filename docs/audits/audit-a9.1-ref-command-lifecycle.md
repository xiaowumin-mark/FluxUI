# A9.1 Ref 命令生命周期审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Widget、Runtime
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph explore "A9.1 Ref command lifecycle ButtonRef InputRef ScrollRef SliderRef DialogRef PopupRef SelectRef component_refs bindInvalidator drainCommands queue command invalidation unmount safety"`
  - `codegraph explore "component_refs.go commandQueue enqueue bindInvalidator drainCommands complete source SelectRef DialogRef PopupRef"`
  - `codegraph explore "ButtonAttachRef drainCommands bindInvalidator button.go pressable click area ref commands execute onClick applyButtonRef"`
  - `codegraph explore "InputAttachRef input.go drainCommands inputCmdSetText inputCmdFocus inputCmdBlur Layout InputRef"`
  - `codegraph explore "ScrollView ScrollAttachRef ScrollRef drainCommands apply scrollCmdToStart scrollCmdToEnd scrollCmdToOffset scrollCmdBy ListView"`
  - `codegraph explore "SliderAttachRef slider.go SliderRef drainCommands sliderCmdSet sliderCmdStep apply set value onChange"`
  - `codegraph explore "SelectAttachRef selection.go SelectRef drainCommands selectCmdSetValue selectCmdOpen selectCmdClose selectCmdToggle DropdownMenu Menu selection commands"`
  - `codegraph explore "DialogAttachRef PopupAttachRef dialogWidget popupWidget drainCommands boolCmdSet boolCmdToggle tabs_dialog_toast.go"`
  - `rg -n "drainCommands|bindInvalidator|InputAttachRef|inputCmd|commitProgrammaticInput|Focus\(|Blur\(" widget/input.go`
  - `rg -n "drainCommands|bindInvalidator|ScrollAttachRef|scrollCmd|applyScroll|ScrollBy|ScrollTo" widget/list_grid.go`
  - `rg -n "drainCommands|bindInvalidator|SelectAttachRef|selectCmd|onOpenChange|open :=|value :=" widget/selection.go`
  - `rg -n "drainCommands|bindInvalidator|DialogAttachRef|PopupAttachRef|boolCmd|open :=|onOpenChange|onOpened|onClosed" widget/tabs_dialog_toast.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/component_refs.go`
  - `widget/scroll_ref.go`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/input.go`
  - `widget/list_grid.go`
  - `widget/slider.go`
  - `widget/selection.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/refs_test.go`
  - `widget/input_event_test.go`
  - `examples/component_lab/main.go`
  - `examples/docs_browser/controls_core_demo.go`
  - `examples/docs_browser/controls_selection_demo.go`
  - `examples/docs_browser/controls_overlay_demo.go`

## 事实结论

1. A9.1 的目标是审查 `ButtonRef`、`InputRef`、`ScrollRef`、`SliderRef`、`DialogRef`、`PopupRef`、`SelectRef` 的命令入队、消费、失效、卸载安全规则；验收要求每个 Ref 何时生效可预测。证据：`docs/project-audit-task-breakdown.md:412`、`docs/project-audit-task-breakdown.md:416`、`docs/project-audit-task-breakdown.md:417`、`docs/project-audit-task-breakdown.md:418`。
2. `component_refs.go` 的通用 `commandQueue[T]` 是多数 Ref 的生命周期核心：`enqueue` 在互斥锁内 append 命令并读取当前 `invalidator`，解锁后触发 invalidation；`bindInvalidator` 覆盖当前 invalidator；`drainCommands` 复制当前命令切片并清空原队列。证据：`widget/component_refs.go:5`、`widget/component_refs.go:11`、`widget/component_refs.go:16`、`widget/component_refs.go:21`、`widget/component_refs.go:27`、`widget/component_refs.go:33`、`widget/component_refs.go:35`。
3. `ScrollRef` 使用与通用队列等价的专用实现，同样用互斥锁保护 `commands` 和 `invalidator`；`ScrollToOffset` 会把负 offset clamp 到 0，`ScrollBy(0)` 不入队。注释明确 ScrollRef 命令在下一帧由 `ScrollView` 消费。证据：`widget/scroll_ref.go:20`、`widget/scroll_ref.go:22`、`widget/scroll_ref.go:54`、`widget/scroll_ref.go:55`、`widget/scroll_ref.go:65`、`widget/scroll_ref.go:75`、`widget/scroll_ref.go:97`。
4. Ref 命令没有全局 runtime registry，也不直接绑定组件实例；组件每次 `Layout` 到达对应节点时调用 `bindInvalidator(redrawInvalidator(ctx))`，随后 `drainCommands`。因此命令生效点是“下一次该 Ref 绑定的组件参与布局并 drain 队列”而不是调用 Ref 方法的瞬间。证据：`widget/button.go:165`、`widget/input.go:405`、`widget/list_grid.go:123`、`widget/slider.go:235`、`widget/selection.go:470`、`widget/tabs_dialog_toast.go:942`、`widget/tabs_dialog_toast.go:1820`。
5. `ButtonRef.Click` 入队空命令；`Button` 布局时先 drain，再在非 disabled 且非 loading 时把每条命令转换为 `DispatchClickEvent(ctx, nil)`。如果按钮当前 disabled/loading，命令会被 drain 但不分发，属于“消费后失效”。证据：`widget/component_refs.go:48`、`widget/component_refs.go:52`、`widget/button.go:165`、`widget/button.go:167`、`widget/button.go:168`、`widget/button.go:170`。
6. `InputRef` 支持 `SetText`、`Append`、`Clear`、`Focus`、`Blur`；文本命令在 `TextField` 布局期通过 `commitProgrammaticInput` 产生 programmatic `beforeinput/input/change` 链路，focus/blur 命令直接执行 Gio `key.FocusCmd`。证据：`widget/component_refs.go:127`、`widget/component_refs.go:137`、`widget/component_refs.go:147`、`widget/component_refs.go:154`、`widget/component_refs.go:161`、`widget/input.go:405`、`widget/input.go:409`、`widget/input.go:416`、`widget/input.go:418`、`widget/input.go:571`。
7. `ScrollRef` 命令在 `ScrollView.Layout` 进入时消费：`ScrollToStart` 重置 `First/Offset` 并标记 before end，`ScrollToEnd` 通过 `forceToEnd` 设置贴底，`ScrollToOffset` 写入主轴 offset，`ScrollBy` 累加为 `pendingScrollBy` 并在布局内容时按 content major 换算。证据：`widget/list_grid.go:123`、`widget/list_grid.go:127`、`widget/list_grid.go:131`、`widget/list_grid.go:134`、`widget/list_grid.go:138`、`widget/list_grid.go:143`、`widget/list_grid.go:157`、`widget/list_grid.go:249`。
8. `SliderRef` 支持 `SetValue` 和 `StepBy`；布局期把命令应用到当前受控值，按 `min/max/step` 归一。普通 Slider 更新 `cfg.value`，RangeSlider 更新 `cfg.valueEnd` 并保证 `valueStart <= valueEnd`。`SliderOnChange` 是否触发仍由后续值差异和 disabled gate 决定。证据：`widget/component_refs.go:293`、`widget/component_refs.go:303`、`widget/slider.go:235`、`widget/slider.go:241`、`widget/slider.go:243`、`widget/slider.go:245`、`widget/slider.go:249`、`widget/slider.go:254`、`widget/slider.go:292`、`widget/slider.go:297`。
9. `SelectRef` 支持 `SetValue`、`Open`、`Close`、`Toggle`；布局期先 drain，再仅在 Select 非 disabled 时执行命令。`SetValue` 会在值变化且存在 `onChange` 时调用旧回调；open/close/toggle 会更新 `state.opened` 并调用 `onOpen`。如果 Select disabled，命令会被 drain 但不生效。证据：`widget/component_refs.go:378`、`widget/component_refs.go:388`、`widget/component_refs.go:395`、`widget/component_refs.go:402`、`widget/selection.go:470`、`widget/selection.go:472`、`widget/selection.go:473`、`widget/selection.go:476`、`widget/selection.go:481`、`widget/selection.go:488`、`widget/selection.go:495`。
10. `DialogRef` 和 `PopupRef` 共享 `boolCommand` 语义：`Open`/`Close` 入队显式 bool，`Toggle` 入队切换命令。布局期从构造参数 `open` 开始顺序应用全部命令，随后按 `state.wasOpen` 与 `open` 的差异触发 `onOpen`、`onClose`、`onOpenChange`，动画完成路径再触发 `onOpened`/`onClosed`。证据：`widget/component_refs.go:452`、`widget/component_refs.go:460`、`widget/component_refs.go:470`、`widget/component_refs.go:480`、`widget/component_refs.go:501`、`widget/component_refs.go:509`、`widget/component_refs.go:519`、`widget/component_refs.go:529`、`widget/tabs_dialog_toast.go:940`、`widget/tabs_dialog_toast.go:944`、`widget/tabs_dialog_toast.go:956`、`widget/tabs_dialog_toast.go:965`、`widget/tabs_dialog_toast.go:1818`、`widget/tabs_dialog_toast.go:1822`、`widget/tabs_dialog_toast.go:1834`、`widget/tabs_dialog_toast.go:1843`。
11. 命令顺序是 FIFO 且一次 drain 内顺序应用；同一帧 drain 前多次调用 Ref 会在下一次布局批量消费。例如 Dialog/Popup 的多条 bool 命令按 drain 顺序不断更新局部 `open`，Select 的多条命令按顺序更新 `currentValue` 和 `state.opened`。证据：`widget/component_refs.go:13`、`widget/component_refs.go:33`、`widget/tabs_dialog_toast.go:944`、`widget/tabs_dialog_toast.go:952`、`widget/selection.go:474`。
12. 未挂载或条件渲染隐藏时，没有对应组件执行 `bindInvalidator`/`drainCommands`，队列会保留命令；如果此前曾绑定过 invalidator，后续调用仍会请求重绘，但命令只有在同一个 Ref 重新绑定到某个组件布局时才被消费。代码中未观察到显式 unbind、队列过期或 target identity 校验。证据：`widget/component_refs.go:21`、`widget/component_refs.go:27`、`widget/scroll_ref.go:88`、`widget/scroll_ref.go:97`、`widget/button.go:165`、`widget/selection.go:470`。
13. nil Ref 方法均安全返回，不会 panic；空值命令有局部过滤：`InputRef.Append("")`、`SliderRef.StepBy(0)`、`ScrollRef.ScrollBy(0)` 不入队。证据：`widget/component_refs.go:49`、`widget/component_refs.go:128`、`widget/component_refs.go:138`、`widget/component_refs.go:294`、`widget/component_refs.go:304`、`widget/scroll_ref.go:66`、`widget/scroll_ref.go:76`。
14. 现有自动测试覆盖了队列基础和 smoke 级 attach：`widget/refs_test.go` 覆盖 Button/Input/Scroll/Dialog/Popup 的 drain 行为和主要组件 attach 不 panic，`widget/input_event_test.go` 覆盖 `InputRef` programmatic input event，`widget/interactive_layout_test.go` 和 `widget/list_grid_test.go` 有部分 Select/Scroll 交互回归。示例入口集中在 component lab 和 docs browser 的 controls demos。证据：`widget/refs_test.go:14`、`widget/refs_test.go:27`、`widget/refs_test.go:67`、`widget/refs_test.go:84`、`widget/refs_test.go:119`、`widget/input_event_test.go:20`、`widget/interactive_layout_test.go:529`、`widget/list_grid_test.go:733`、`examples/component_lab/main.go:69`、`examples/component_lab/main.go:607`、`examples/component_lab/main.go:1189`。

### Ref 生命周期矩阵

| Ref | 入队 API | 消费点 | 生效时机 | 禁用/不可用行为 | 卸载/隐藏行为 |
| --- | --- | --- | --- | --- | --- |
| `ButtonRef` | `Click()` | `buttonWidget.Layout` | 下一次绑定按钮布局时分发 programmatic click | disabled/loading 时 drain 后丢弃，不触发 click | 未布局时保留队列，重新绑定后消费 |
| `InputRef` | `SetText`、`Append`、`Clear`、`Focus`、`Blur` | `inputWidget.Layout` | 下一次输入框布局时执行文本 mutation 或 Gio focus command | disabled/readOnly 下仍会处理 programmatic 文本命令；focus/blur 直接交给 Gio | 未布局时保留队列，重新绑定后消费 |
| `ScrollRef` | `ScrollToStart`、`ScrollToEnd`、`ScrollToOffset`、`ScrollBy` | `scrollWidget.Layout` | 下一次 ScrollView 布局前更新 Gio List position 或 pending delta | 无 disabled gate；offset 负值 clamp，0 delta 不入队 | 未布局时保留队列，重新绑定后消费 |
| `SliderRef` | `SetValue`、`StepBy` | `sliderWidget.Layout` | 下一次 Slider 布局时更新局部受控值，随后可能触发 onChange | 命令仍应用到局部 cfg；disabled 会阻止后续 onChange | 未布局时保留队列，重新绑定后消费 |
| `SelectRef` | `SetValue`、`Open`、`Close`、`Toggle` | `selectWidget.Layout` | 下一次 Select 布局时更新 current value/open state | disabled 时 drain 后不执行命令 | 未布局时保留队列，重新绑定后消费 |
| `DialogRef` | `Open`、`Close`、`Toggle` | `dialogWidget.Layout` | 下一次 Dialog 布局时覆盖本帧 open 值，并触发生命周期回调 | 无 disabled gate；受控 `open` 参数仍是下一帧基线 | 未布局时保留队列，重新绑定后消费 |
| `PopupRef` | `Open`、`Close`、`Toggle` | `popupWidget.Layout` | 下一次 Popup 布局时覆盖本帧 open 值，并触发生命周期回调 | 无 disabled gate；受控 `open` 参数仍是下一帧基线 | 未布局时保留队列，重新绑定后消费 |

## 风险

1. `bindInvalidator` 只有覆盖没有解绑；Ref 若从一个组件移到另一个组件，队列不会记录“命令属于哪个组件实例”。旧组件隐藏后调用 Ref 仍可能触发最后一次绑定的 invalidator，命令会在后续任一同 Ref 绑定组件布局时消费。
2. Button 和 Select 的 disabled 语义是 drain 后丢弃命令，而不是保留到恢复可用后执行；这对“按钮暂时 loading 后再点击”或“Select disabled 期间 Open”是可预测但需要文档化的行为。
3. Dialog/Popup/Select 是受控入参加本帧命令覆盖的模型；如果调用方没有在 `onOpenChange` 中同步外部 state，Ref 命令可能只影响当前帧/动画路径，下一帧又被旧 `open` 入参覆盖。
4. SliderRef 对 RangeSlider 只操作 `valueEnd`，没有单独控制 start thumb 的命令；后续若文档承诺 RangeSlider 双端 Ref，需要扩展 API 或明确当前边界。
5. InputRef programmatic 文本命令会触发 input/change 事件链；如果调用方在回调内再次调用同一个 Ref，命令会进入下一次 drain，需避免无条件反馈循环。
6. 队列没有容量限制或命令去重；长期未挂载的 Ref 被持续调用会累积命令，重新挂载时批量执行，可能造成突然跳转、多次事件或过期 open/close 序列。

## 验收

- 已输出 ButtonRef、InputRef、ScrollRef、SliderRef、DialogRef、PopupRef、SelectRef 的入队 API、消费点、生效时机、禁用/不可用行为和卸载/隐藏行为矩阵。
- 已确认 Ref 命令生效模型是“命令入队后请求 redraw，下一次绑定组件布局时 drain 并消费”，不是立即执行，也不是 runtime 全局调度。
- 已标出 disabled/loading 时 drain 后丢弃的组件：Button、Select；已标出无 disabled gate 的组件：ScrollView、Dialog、Popup；已标出 Slider/Input 的特殊边界。
- 已说明未挂载/条件隐藏时命令不会自动失效，而是保留到后续 drain；当前实现没有显式 unbind、TTL 或 target identity 校验。
- 已确认现有测试覆盖基础队列和 attach smoke，但缺少跨卸载、跨 Ref 复用、disabled 期间命令丢弃、Dialog/Popup 受控 state 反馈的行为测试。

## 后续依赖

- A9.2 controlled value 和内部状态审查应继续确认 Dialog/Popup/Select/Slider 的 Ref 命令如何与受控入参、内部 state、`onChange`/`onOpenChange` 反馈闭环组合。
- A9.3 OnChange 触发条件审查应明确 programmatic Ref 命令是否应触发旧 callback，以及 disabled/loading 时 drain 后丢弃是否属于正式兼容语义。
- A10 组件族修复前应补充 Ref 行为测试：未挂载期间队列保留、重新挂载后消费、disabled 期间命令丢弃、Select/Dialog/Popup 受控 state 未同步时的覆盖风险。
