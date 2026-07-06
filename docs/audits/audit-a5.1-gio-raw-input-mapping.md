# A5.1 Gio 原始输入映射表

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：事件系统审查。

- 状态：Done
- 日期：2026-07-06 21:18:00 +08:00
- 负责人：Codex
- 关注：Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A5\\.1|Gio 原始输入|pointer|clipboard|drag" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "PointerEventFromGio|WheelEventFromGio|KeyboardEventFromGio|pointer\\.Event|key\\.Event|clipboard|Clipboard|EditEvent|editor|Transfer|drag|drop|Drag" event internal widget ui examples -g "*.go"`
  - `rg -n "KeyboardEventFromGio|key\\.Filter|key\\.Event|DispatchKeyboardEvent|OnKeyDown|OnShortcut|FocusTarget" internal widget ui event -g "*.go"`
  - `rg -n "ReadClipboard|WriteClipboard|clipboard|transfer\\.|OfferCmd|TargetFilter|SourceFilter" system internal widget event ui -g "*.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/input.go`
  - `event/keyboard.go`
  - `event/text.go`
  - `event/drag.go`
  - `event/event.go`
  - `internal/events.go`
  - `widget/pointer_area.go`
  - `widget/keyboard_scope.go`
  - `widget/input.go`
  - `widget/drop_target.go`
  - `widget/drag_source.go`
  - `widget/slider.go`
  - `widget/utils.go`
  - `system/clipboard.go`
  - `system/drag_drop.go`
- 关联能力：
  - Gio pointer 到 FluxUI PointerEvent/WheelEvent 字段映射
  - Gio key 到 FluxUI KeyboardEvent 字段映射
  - Gio Editor Change/Submit 到 FluxUI InputEvent best-effort 映射
  - Gio transfer 到 FluxUI DragEvent/DropEvent 映射
  - clipboard system API 与 event system 边界识别
  - 不可用字段、best-effort 字段、后端依赖字段标注

## Gio 到 FluxUI 输入映射总表

| Gio 来源 | 桥接入口 | FluxUI 事件 | 字段映射 | 状态 |
| --- | --- | --- | --- | --- |
| `pointer.Event.Kind` | `event.TypeFromGioPointerKind`，`widget/pointer_area.go` 的 `processPointerAreaEvents` | `pointerdown`、`pointerup`、`pointermove`、`pointerenter`、`pointerleave`、`pointercancel`、`wheel`；PointerArea 额外合成 `pointerover`、`pointerout`、`click`、`dblclick`、`auxclick`、`contextmenu` | `Press -> pointerdown`，`Release -> pointerup`，`Move/Drag -> pointermove`，`Enter/Leave/Cancel -> pointerenter/pointerleave/pointercancel`，`Scroll -> wheel` | 可用；`pointerover/out/click-like` 为 FluxUI 合成事件 |
| `pointer.Event.PointerID` | `event.PointerEventFromGio` | `PointerEvent.PointerID`，`PointerSample.PointerID` | `uint64(ev.PointerID)` | 可用；受 Gio pointer ID 语义约束 |
| `pointer.Event.Source` | `event.PointerTypeFromGio` | `PointerEvent.PointerType` | `pointer.Mouse -> mouse`，`pointer.Touch -> touch`，其他 -> `other` | 可用但不完整；Gio 0.9 未提供 pen 类型，FluxUI 的 `pen` 目前不可由 Gio 映射产生 |
| `pointer.Event.Buttons` | `event.ButtonFromGio` / `ButtonsFromGio` | `PointerEvent.Button` / `PointerEvent.Buttons` | primary/secondary/tertiary/quaternary/quinary 映射到 DOM 风格 primary/secondary/auxiliary/back/forward | 可用；`Button` 来自当前按钮集合推断，不是 Gio 单独 changed button 字段 |
| `pointer.Event.Position` | `event.PointerEventFromGio`、`WheelEventFromGio` | `PointerEvent.Position`、`WheelEvent.Position`、coalesced sample position | 原样复制 Gio 局部坐标；outside-dismiss 会按 viewport 坐标重写 position | 可用；坐标系由 Gio tag/transform 决定 |
| `pointer.Event.Scroll` | `event.WheelEventFromGio` | `WheelEvent.DeltaX`、`DeltaY` | `Scroll.X -> DeltaX`，`Scroll.Y -> DeltaY`，`DeltaMode = WheelDeltaPixel`，`DeltaZ = 0` | 可用；delta mode 固定为 pixel，Z 轴不可用 |
| `pointer.Event.Modifiers` | `event.ModifiersFromGio` | `Modifiers` | Ctrl/Shift/Alt/Meta/Shortcut 逐项映射；Meta 来自 Command 或 Super | 可用；Shortcut 取 Gio `ModShortcut` |
| `pointer.Event.Time` | `event.PointerEventFromGio`、`WheelEventFromGio` | `TimeOffset` | 原样复制 Gio 相对时间；FluxUI `Event.Time` 使用 `ctx.Now()` 或 `time.Now()` | 可用；绝对时间不是 Gio 原始字段 |
| `pointer.Event.Priority` | 无 | 无 | 未写入 FluxUI event | 当前不可用 |
| Gio pressure/tilt/twist/width/height | 无 Gio 来源 | `PointerEvent.Width/Height/Pressure/TangentialPressure/TiltX/TiltY/Twist` | 字段保留为零值 | 当前不可用；属于未来后端或 Gio 能力扩展字段 |
| `key.Event.State` | `event.KeyboardEventFromGio` | `KeyboardEvent.Type` | `key.Release -> keyup`，其他 -> `keydown` | 可用；keyup 后端覆盖受 Gio 平台支持约束 |
| `key.Event.Name` | `event.KeyboardEventFromGio` / `keyNameForEvent` | `KeyboardEvent.Key`、`KeyboardEvent.Code` | Return/Enter/Space/Escape/Tab 规范化为常见字符串；其他 `Key` 和 `Code` 均为 `string(ev.Name)` | 可用但简化；没有 DOM `code` 的物理键位语义 |
| `key.Event.Modifiers` | `event.ModifiersFromGio` | `KeyboardEvent.Modifiers` | 同 pointer modifiers | 可用 |
| key repeat | `internal.Runtime.recordKeyState` | `KeyboardEvent.Repeat` | runtime 用 `Code` 或 `Key` 维护跨帧按下集合，重复 keydown 标记为 repeat | best-effort；依赖 keyup 能否到达 |
| key location / composing | 无 Gio 来源 | `KeyboardEvent.Location`、`IsComposing` | 保持零值/false | 当前不可用 |
| `widget.Editor.ChangeEvent` | `widget/input.go` 的 `editor.Update` -> `commitUserInput` | `beforeinput`、`input`、`change` | 用编辑前后文本 diff 填 `Data`、`InputType`、`Source`、`Value`、`PreviousValue`，`Native` 保存 `ChangeEvent` | best-effort；Gio 多数编辑事件为变更后通知，`beforeinput` 取消通过回滚实现 |
| `widget.Editor.SubmitEvent` | `dispatchSubmit` | `submit` | `Value = ev.Text`，`InputType = insertLineBreak`，`Source = user`，`Native` 保存 `SubmitEvent` | 可用；是否产生 submit 取决于 `editor.Submit` 配置 |
| clipboard paste/copy | `system/clipboard.go` | 无 FluxUI event | `ReadClipboardText/WriteClipboardText` 等为系统能力 API，不进入 `event` 包 | 后端依赖；当前不是 Gio 原始输入事件映射 |
| `transfer.InitiateEvent` target | `widget/drop_target.go` | `dragenter`、`dragover` | Active=true，Types 来自 DropTarget 接受类型，Native 保存 Gio event | 可用；payload 尚不可读 |
| `transfer.CancelEvent` target | `widget/drop_target.go` | `dragleave` | Active=false，Native 保存 Gio event | 可用 |
| `transfer.DataEvent` target | `dropEventFromTransfer` | `drop` | `Type -> MIMEType/DropEvent.Type`，`Open()` 读取 Data/Text，`text/uri-list` 或绝对路径解析为 Paths，Operation 来自组件配置，Native 保存 Gio event | 可用；数据只在事件帧可读，大小受 `DropTargetMaxBytes` 限制 |
| `transfer.InitiateEvent` source | `widget/drag_source.go` | `dragstart` | DragSource active 状态开始，Operation 来自配置 | 可用；同时依赖 Gio transfer backend |
| `transfer.RequestEvent` source | `widget/drag_source.go` | `drag` | `RequestEvent.Type -> MIMEType`，payload bytes -> Data/Text，执行 `transfer.OfferCmd` | 可用；payload 由应用配置提供 |
| `transfer.CancelEvent` source / pointer release | `widget/drag_source.go` | `dragend` | active 结束或取消，Operation 来自配置 | 可用；完成/取消语义由 Gio 请求和本地 pointer gesture 共同决定 |

## 关键桥接路径

| 路径 | 证据 | 审查结论 |
| --- | --- | --- |
| 低层 pointer 统一转换 | `event/input.go:229` 定义 `PointerEventFromGio`，`widget/pointer_area.go:215` 处理 Gio pointer events，`widget/pointer_area.go:307` 分发 FluxUI pointer event | PointerArea 是通用 pointer/wheel 桥接入口；Slider 等控件可直接使用同一转换函数。 |
| move/drag coalescing | `widget/pointer_area.go:219` drain events，`widget/pointer_area.go:231` 将连续 move/drag 合并成 `Coalesced` | `CoalescedSamples()` 对外可读的是 FluxUI 合成后的 raw sample 列表，不是 Gio 单独的 coalesced API。 |
| pointer capture target | `widget/pointer_area.go:319` 和 `widget/slider.go` 的 `sliderPointerDispatchTarget` 均优先使用 runtime capture target | Gio 原始事件仍由当前 tag 获取，但 FluxUI 分发目标可被 runtime pointer capture 改写。 |
| wheel 映射 | `event/input.go:250` 定义 `WheelEventFromGio`，`widget/pointer_area.go:276` 分发 wheel | Wheel 只记录 X/Y pixel delta、position、modifiers、time offset；未记录惯性、phase、device 等字段。 |
| keyboard raw dispatch | `widget/keyboard_scope.go:186` 只在当前 focus target 位于 scope path 内消费 `key.Filter{}`，`widget/keyboard_scope.go:200` 调用 `KeyboardEventFromGio` | 键盘事件以 component-tree focus 为目标，不是按 Gio focus tag 逐控件直接分发。 |
| input/editor dispatch | `widget/input.go:427` 消费 `editor.Update`，`widget/input.go:588` 处理用户输入，`widget/input.go:625` 创建 `InputEvent` | Text input 不是直接映射 Gio `key.Event`，而是基于 Gio editor mutation 生成 browser-style input 事件。 |
| drop target dispatch | `widget/drop_target.go:173` 使用 `transfer.TargetFilter`，`widget/drop_target.go:215` 转成 `DragEvent` | DropTarget 的事件系统表面是 `DragEvent`；旧 `DropEvent` widget 回调仍保留。 |
| drag source dispatch | `widget/drag_source.go:243` 使用 `transfer.SourceFilter`，`widget/drag_source.go:258` 执行 `OfferCmd`，`widget/drag_source.go:323` 转成 `DragEvent` | DragSource 对 Gio transfer 请求提供 payload，并把生命周期投射为 `dragstart/drag/dragend`。 |
| clipboard 边界 | `system/clipboard.go:29`、`system/clipboard.go:46` 等系统 API；未发现 `event`/`widget` 对 clipboard raw event 的转换入口 | Clipboard 当前是 `system` 能力，不是 Event 子系统的原始输入映射。 |

## 事实结论

- A5.1 输入范围内，FluxUI 已有显式 Gio 原始输入转换函数：`PointerEventFromGio`、`WheelEventFromGio`、`KeyboardEventFromGio`。这些函数是 pointer/wheel/key 字段进入 FluxUI typed event 的权威映射层。
- Pointer 事件的 FluxUI target 不一定等于当前 Gio tag 对应的组件路径：`PointerArea` 和 `Slider` 都会先查询 runtime pointer capture；存在 capture 时事件分发到 capture owner。
- Pointer `move` 与 `drag` 都映射为 `pointermove`。`PointerArea` 会把连续同一 pointer 的 move/drag 聚合成 `Coalesced` samples，并以最后一个 Gio event 作为主事件。
- `pointerenter`/`pointerleave` 不冒泡且不可取消；`pointercancel` 不可取消；其他 pointer/click-like 事件通常冒泡且可取消。这由 `event/input.go` 的 `applyPointerDefaults` 规则决定。
- `click`、`dblclick`、`auxclick`、`contextmenu` 不是 Gio 原始事件，而是 `PointerArea` 在 press/release 状态、按钮和时间/距离阈值基础上合成的 FluxUI 事件。
- Keyboard 事件从 `KeyboardScope` 拉取 `key.Filter{}`，再分发到 runtime focused target；没有 focus target 或 focused target 不在该 scope 的 event path 内时，该 scope 不消费 raw key event。
- Keyboard `Repeat` 不来自 Gio 字段，而由 runtime `keyDown` 集合在 `keydown`/`keyup` 分发期间维护。因此 repeat 依赖 key identity 稳定性和 keyup 可达性。
- Text input 事件不是 `key.Event` 的简单映射。`Input` 控件读取 Gio `widget.Editor` 的 `ChangeEvent` / `SubmitEvent`，用文本 diff 和历史索引推断 `InputType` 与 `Source`；`BestEffort=true` 明确标注该推断性质。
- `beforeinput` 对用户编辑的取消是 best-effort：因为 Gio editor 多数情况下已经完成 mutation，FluxUI 在取消时把 editor 文本回滚到 `PreviousValue`。
- Clipboard 相关能力位于 `system` 包，包含文本、文件列表、PNG 图像读写；本次未发现 clipboard raw event 被转为 FluxUI `event.Event` 的路径。
- Drag/drop 使用 Gio `transfer` API：DropTarget 的 `Initiate/Cancel/Data` 转为 `dragenter/dragover/dragleave/drop`，DragSource 的 `Initiate/Request/Cancel` 和本地 pointer gesture 转为 `dragstart/drag/dragend`。
- Drag/drop payload 的可用性依赖 Gio transfer backend 和当前平台能力；`system.ProbeDragAndDrop` 提供能力探测，但 widget 事件映射本身仍以 Gio `transfer` event 为直接输入。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| Pointer changed button 语义不完整 | 中 | `ButtonFromGio` 从当前 pressed buttons 集合推断 changed button；release 时 `PointerArea` 用 press 时保存的 button 覆盖，普通转换函数本身无法表达 Gio 未提供的独立 changed-button 字段。 |
| Pointer pen/pressure/tilt 等字段不可用 | 中 | FluxUI `PointerEvent` 预留了 DOM 风格字段，但 Gio 0.9 raw pointer event 未提供对应数据；当前保持零值，后续文档和 API 不应承诺其可用。 |
| Keyboard `Code` 不是 DOM physical code | 中 | 当前 `Code = string(ev.Name)`，与浏览器 `KeyboardEvent.code` 的物理键位语义不同；快捷键如果依赖键盘布局无关的 physical key，需要后续能力补足。 |
| Keyboard repeat 依赖 keyup | 中 | runtime 用 keyDown map 推断 Repeat；若某平台或焦点路径导致 keyup 丢失，repeat 状态可能偏保守或残留。Gio 注释显示 release 事件有平台支持范围。 |
| Text input Source/InputType 为 best-effort | 中 | paste/delete/undo/redo 由 diff、文本长度和历史序列推断；复杂 IME、替换、批量编辑无法完全等同浏览器 beforeinput/inputType。 |
| Composition event 未接入 | 中 | `CompositionEvent` API 已存在，但本次未发现 Gio editor/IME composition lifecycle 到 FluxUI composition event 的桥接。 |
| Clipboard 不在 event system 内 | 低 | `system` clipboard API 可用但没有 paste/copy/cut typed event；后续如要实现编辑器 paste default action，需要明确跨 `system` 和 `event` 的所有权。 |
| Drag/drop operation negotiation 有限 | 中 | DragSource/DropTarget 的 Operation 多来自组件配置；Gio transfer API 当前未暴露完整 OS operation negotiation，文档应标注为应用层语义。 |
| Drop payload 读取受 frame 生命周期限制 | 中 | `transfer.DataEvent.Open` 只在收到事件的 frame 内有效，FluxUI 立即读取并复制；大 payload 受 max bytes 限制，超限进入 Err。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出 Gio 字段到 FluxUI event 字段映射表 | 通过 | 本文“Gio 到 FluxUI 输入映射总表”覆盖 pointer、key、editor、clipboard、drag/drop。 |
| 标出当前不可用字段 | 通过 | 已标注 pointer pressure/tilt/twist/width/height、key location/composing、clipboard raw event、composition bridge 等不可用或未接入项。 |
| 标出 best-effort 字段 | 通过 | 已标注 InputEvent `InputType`/`Source`、`beforeinput` rollback、keyboard Repeat、Pointer click-like/coalescing 等推断或合成语义。 |
| 标出后端依赖字段 | 通过 | 已标注 keyup 平台覆盖、clipboard system driver、drag/drop Gio transfer backend、drop payload frame 生命周期。 |
| 没有修改运行时代码 | 通过 | 本任务只新增审计子文件并更新审计索引，不修改 `event`、`widget`、`internal`、`system` 源码。 |
| Go workspace 安全检查 | 记录 | `gopls go_vulncheck ./...` 完成，输出多项 Go 标准库与 x/image/x/sys 相关 findings；本任务未改依赖，作为环境风险保留。 |

## 后续依赖

- A5.2 pointer event flow 审查应继续沿 `PointerArea`、`Slider`、ClickArea/Pressable 和 overlay outside press 路径确认 target、capture、hit area、defaultPrevented 的传播效果。
- A5.3 keyboard/focus/default action 审查应细化 `KeyboardScope`、runtime focus target、shortcut、Tab/Enter/Space 默认行为，以及 Gio focus 与 FluxUI focus 的边界。
- A5.4 text input/default action 或 A7 文本输入审查应覆盖 composition lifecycle、paste/cut/copy default action、beforeinput 取消后 editor selection/caret 的一致性。
- A6 scroll 审查应继续确认 `WheelEvent` 与 ScrollView/ListView 的 wheel consumption、nested scroll 和 axis clamp 规则。
- A10 perf 审查应关注 high-frequency pointer move coalescing 是否足够降低分配与 dispatch 开销，并补充 pointer/wheel/key benchmark。
