<!-- fluxui-doc-meta
{
  "id": "event_system_roadmap",
  "title": "高级事件系统路线图",
  "category": "使用指南",
  "order": 123,
  "summary": "以浏览器 DOM / UI / Pointer Events 为参照，规划 FluxUI 从组件语义回调走向高级事件系统的演进路线。",
  "example": { "id": "pressable_basic" },
  "apis": [
    "event.Event",
    "event.PointerEvent",
    "event.KeyboardEvent",
    "event.InputEvent",
    "event.WheelEvent",
    "event.FocusEvent",
    "event.ListenerOptions",
    "event.Dispatch(ctx *Context, target TargetID, event Event) bool",
    "On(type event.Type, fn event.Handler, opts ...event.ListenerOption) WidgetOption",
    "PointerArea(child Widget, opts ...PointerAreaOption) Widget",
    "KeyboardScope(child Widget, opts ...KeyboardScopeOption) Widget"
  ]
}
-->

# 高级事件系统路线图

本文档记录 FluxUI 后续事件系统的设计方向。目标不是把浏览器 API 逐字搬到 Go 里，而是以浏览器事件系统的成熟模型为标准，补齐 FluxUI 现在对开发者暴露不足的问题：原始输入数据少、事件类型少、没有传播模型、无法取消默认行为、键盘与焦点能力薄弱，自定义交互只能绕回 Gio 底层。

参考标准：

- WHATWG DOM Standard: `EventTarget`、`addEventListener`、事件路径、捕获/目标/冒泡、`stopPropagation`、`preventDefault`、`dispatchEvent`。
- W3C UI Events: `UIEvent`、`MouseEvent`、`KeyboardEvent`、`WheelEvent`、`FocusEvent`、`InputEvent`、`CompositionEvent`。
- W3C Pointer Events Level 3: 统一鼠标、触摸、手写笔输入，包含 `pointerId`、`pointerType`、按钮状态、压力、倾斜、pointer capture、coalesced/predicted events。
- MDN Event reference: 常用事件类型和开发者使用习惯，用作 API 易用性参考。

截至 2026-07-04，FluxUI 的事件能力更接近“组件语义回调”，还不是完整的应用级事件系统。本文档作为后续实现和评审的路线图。

## 当前状态

### 已经可用的能力

- `event.UseClickable(ctx)` 基于 Gio `widget.Clickable` 提供点击状态，公开能力主要是 `Clicked(ctx)`、`Hovered()`、`Pressed()`、`Focused(ctx)` 和 `Snapshot(ctx, includeFocus)`。
- `event.Dispatcher` 当前只包含 `Click` 和 `Hover` 两类处理器。
- `Button` / `Pressable` / `ClickArea` / `ContainerDecorationOnClick` 等组件提供 `OnClick(ctx)`、`OnHover(ctx, hovering)` 这种高层语义回调。
- `Input` 基于 Gio `widget.Editor`，公开 `InputOnChange(ctx, value)`、`InputOnFocus(ctx, focused)` 和 `InputRef.SetText/Append/Clear/Focus/Blur`。
- `ScrollView` 暴露 `ScrollOnChange(ctx, x, y)`，给出滚动位置变化。
- `Slider` 内部已经消费 Gio 原始 pointer 事件，包括 press、move、drag、release、enter、leave、cancel、pointer id、buttons、position，但对外主要只暴露值变化。
- `DragSource` / `DropTarget` 的数据传输模型相对完整，已经有 `DragSourceEvent`、`DropEvent`、MIME type、payload、文件路径和 operation。
- `system.RegisterGlobalShortcut` 提供系统级快捷键事件，但它是 OS-level 能力，不是组件树内的键盘事件传播。
- 高级开发者可以直接使用 `ctx.Gtx.Event(pointer.Filter{...})` 或 `key.Filter` 处理 Gio 原始事件，因为 `Context` 暴露了 `Gtx`。

### 主要缺口

- 没有浏览器式 `EventTarget` 树，也没有捕获、目标、冒泡阶段。
- 没有 `stopPropagation`、`stopImmediatePropagation`、`preventDefault` 和 `defaultPrevented` 语义。
- 没有 listener options，例如 `capture`、`once`、`passive`、可取消订阅信号、监听器优先级。
- 点击回调拿不到坐标、按钮、修饰键、点击次数、pointer id、输入来源、时间戳。
- 鼠标右键、中键、双击、context menu、aux click、wheel delta 没有统一公开。
- 键盘事件没有组件级公开 API，缺少 `keydown`、`keyup`、`repeat`、`key`、`code`、`modifiers`、`isComposing`。
- 文本输入只有值变化，没有 `beforeinput`、`input`、`compositionstart/update/end`、`submit`、选择区变化等事件。
- 焦点事件只有局部状态或组件回调，没有统一 `focus`、`blur`、`focusin`、`focusout`、`relatedTarget` 和可遍历焦点模型。
- 滚动暴露的是位置变化，不是原始 wheel/touchpad delta，也不能通过 passive/non-passive 语义表达是否阻止默认滚动。
- 拖放虽然数据模型较完整，但没有和通用事件传播模型整合成 `dragstart/dragenter/dragover/drop/dragend`。
- 运行时内部的 pointer move 合并、交互统计、redraw reason 更偏诊断和性能，不是稳定的开发者事件数据。

结论：现有事件对普通表单、按钮、列表交互够用，但对画布、游戏、快捷键编辑器、富文本编辑器、设计器、右键菜单、复杂拖拽、自定义手势和跨组件事件治理不够。

## 设计目标

1. 以浏览器事件模型为参照，提供稳定、可组合、可取消、可传播的事件系统。
2. 保持 Gio-native，不强行伪装成 DOM；组件树路径、布局命中、输入注册和 frame 生命周期仍按 FluxUI/Gio 的约束实现。
3. 不破坏现有 `OnClick(func(ctx))`、`InputOnChange`、`ScrollOnChange` 等简单 API；高级事件作为增量能力引入。
4. 让事件对象携带足够数据，开发者不用回到底层 Gio 才能知道鼠标坐标、按钮、修饰键、按键、滚轮、pointer id。
5. 默认行为必须显式建模，例如点击触发、焦点移动、文本输入、滚动、slider 拖动、drag/drop；可取消事件才能调用 `PreventDefault` 生效。
6. 高频事件需要可控性能策略，包括 passive listener、move coalescing、raw/update 分层和事件对象复用策略。
7. 文档和测试必须把事件顺序、传播路径、取消语义写清楚，避免每个组件各自解释。

## 浏览器模型到 FluxUI 的映射

| 浏览器概念 | FluxUI 目标映射 |
| --- | --- |
| `EventTarget` | 每个可监听组件实例拥有稳定 `TargetID`，由 `PathID`、显式 key 或内部 target 生成 |
| DOM tree path | FluxUI 组件树路径和 overlay/portal 路径组成事件路径 |
| capturing phase | root 到 target parent 的捕获阶段 |
| at target | target 上 capture 和 bubble listener 均可按规则触发 |
| bubbling phase | target parent 到 root 的冒泡阶段 |
| `currentTarget` | 当前正在调用 listener 的 FluxUI target |
| `target` | 初始命中的 FluxUI target；必要时允许 retarget |
| `composedPath()` | 返回 FluxUI target path，overlay/portal 需要定义跨边界行为 |
| `preventDefault()` | 取消 FluxUI 默认行为，例如滚动、文本插入、点击激活、focus move |
| `stopPropagation()` | 阻止继续沿事件路径传播 |
| `stopImmediatePropagation()` | 阻止同一 target 后续 listener 和后续传播 |
| `passive` | 声明 listener 不会取消事件，允许滚动和 pointer move 优化 |
| `once` | listener 触发一次后自动移除 |
| synthetic event | 应用代码构造并通过 FluxUI 事件系统分发的自定义事件 |
| trusted event | 来自系统/Gio 的真实用户输入；应用分发的是 untrusted |

FluxUI 不需要实现浏览器 Shadow DOM，但需要为 overlay、popup、dialog、portal 预留类似 `composed` 的边界策略：事件能否穿过浮层边界、是否 retarget 到拥有者、是否停止在 modal root。

## 目标事件数据

### Base Event

建议新增统一基础事件类型：

```go
type Type string

type Phase int

const (
    PhaseNone Phase = iota
    PhaseCapture
    PhaseTarget
    PhaseBubble
)

type Event struct {
    Type          Type
    Target        TargetID
    CurrentTarget TargetID
    Phase         Phase
    Time          time.Time
    Bubbles       bool
    Cancelable    bool
    DefaultPrevented bool
    Trusted       bool
    Detail        any
}

func (e *Event) StopPropagation()
func (e *Event) StopImmediatePropagation()
func (e *Event) PreventDefault() bool
func (e *Event) PropagationStopped() bool
func (e *Event) ImmediatePropagationStopped() bool
func (e *Event) ComposedPath() []TargetID
```

### Pointer Event

Pointer 事件应该优先作为统一低层指针模型，而不是只提供 MouseEvent。

```go
type PointerType string

const (
    PointerMouse PointerType = "mouse"
    PointerPen   PointerType = "pen"
    PointerTouch PointerType = "touch"
    PointerOther PointerType = "other"
)

type PointerEvent struct {
    Event
    PointerID    uint64
    PointerType  PointerType
    IsPrimary    bool
    Button       Button
    Buttons      Buttons
    Position     f32.Point
    LocalPosition f32.Point
    Movement     f32.Point
    Scroll       f32.Point
    Modifiers    Modifiers
    Pressure     float32
    TangentialPressure float32
    TiltX        float32
    TiltY        float32
    Twist        float32
    Width        float32
    Height       float32
    ClickCount   int
}
```

第一阶段必须保证从 Gio 已有数据里可稳定提供 `PointerID`、`Button/Buttons`、`Position`、`Scroll`、`Modifiers`、`Time`、`Source`。压力、倾斜、接触面积等字段可以先存在但标记为 backend dependent，当前后端没有数据时给默认值。

建议事件类型：

- `pointerover`
- `pointerenter`
- `pointerdown`
- `pointermove`
- `pointerrawupdate`
- `pointerup`
- `pointercancel`
- `pointerout`
- `pointerleave`
- `gotpointercapture`
- `lostpointercapture`
- `click`
- `dblclick`
- `auxclick`
- `contextmenu`

### Wheel Event

`WheelEvent` 不应该只映射为滚动位置变化，而应暴露原始 delta：

```go
type DeltaMode int

const (
    DeltaPixel DeltaMode = iota
    DeltaLine
    DeltaPage
)

type WheelEvent struct {
    Event
    DeltaX    float32
    DeltaY    float32
    DeltaZ    float32
    DeltaMode DeltaMode
    Position  f32.Point
    Modifiers Modifiers
}
```

`ScrollView` 继续保留 `ScrollOnChange`，但新增 wheel 事件后，开发者可以选择在 `wheel` 上 `PreventDefault` 来阻止默认滚动。

### Keyboard Event

组件级键盘事件需要和系统全局快捷键区分开。

```go
type KeyboardEvent struct {
    Event
    Key         string
    Code        string
    Location    KeyLocation
    Repeat      bool
    Modifiers   Modifiers
    IsComposing bool
}
```

建议事件类型：

- `keydown`
- `keyup`
- `shortcut`，FluxUI 自定义的局部快捷键语义事件，可由 `keydown` 派生

`keydown` 的默认行为包括文本输入、焦点移动、按钮激活、快捷键触发等。取消 `keydown` 应阻止后续可取消默认行为，但不能吞掉对应的 `keyup`。

### Input and Composition Event

文本编辑需要区分按键和文本输入。键盘不等于字符输入，IME、粘贴、语音输入、虚拟键盘都可能产生文本变化。

```go
type InputEvent struct {
    Event
    Data        string
    InputType   string
    IsComposing bool
}

type CompositionEvent struct {
    Event
    Data string
}
```

建议事件类型：

- `beforeinput`
- `input`
- `change`
- `submit`
- `compositionstart`
- `compositionupdate`
- `compositionend`
- `selectionchange`

`beforeinput` 是高优先级目标，因为它是富文本、过滤输入、快捷键编辑器和自定义编辑器需要的取消点。若 Gio `widget.Editor` 无法在变更前提供完整拦截，需要先文档化限制，再按后端能力逐步补齐。

### Focus Event

```go
type FocusEvent struct {
    Event
    RelatedTarget TargetID
}
```

建议事件类型：

- `focus`
- `blur`
- `focusin`
- `focusout`

`focus` / `blur` 不冒泡，`focusin` / `focusout` 冒泡。需要建立统一 FocusManager，负责当前 focus target、tab order、programmatic focus、disabled/hidden target 处理和 `relatedTarget`。

### Drag, Drop and Clipboard Event

现有 `DragSourceEvent` / `DropEvent` 可以作为数据层基础，后续纳入通用传播：

- `dragstart`
- `drag`
- `dragenter`
- `dragover`
- `dragleave`
- `drop`
- `dragend`
- `copy`
- `cut`
- `paste`

拖放事件需要支持 `PreventDefault` 语义：例如 `dragover` 被允许后才视为可 drop target。Clipboard 事件应和 `system` clipboard 能力解耦，组件内事件只表达用户意图和数据交换。

## 监听器 API 方向

底层建议在 `event` 包提供通用监听器注册模型：

```go
type ListenerOptions struct {
    Capture bool
    Once    bool
    Passive bool
    Priority int
}

type Handler func(ctx *internal.Context, ev EventLike)
```

组件和 Element 层可以暴露更符合 FluxUI 风格的选项：

```go
ui.On(event.PointerDown, func(ctx *ui.Context, ev *event.PointerEvent) {
    // ev.Position, ev.Buttons, ev.Modifiers
}, ui.EventCapture())

ui.On(event.KeyDown, func(ctx *ui.Context, ev *event.KeyboardEvent) {
    if ev.Key == "Escape" {
        ev.StopPropagation()
    }
})

ui.PointerAreaElement(
    child,
    ui.PointerOnMove(func(ctx *ui.Context, ev *event.PointerEvent) {}),
    ui.PointerCaptureOnPress(true),
)

ui.KeyboardScopeElement(
    child,
    ui.KeyOnDown(func(ctx *ui.Context, ev *event.KeyboardEvent) {}),
)
```

同时保留简单语义 API：

```go
ui.ButtonElement(
    ui.TextElement("Save"),
    ui.OnClick(func(ctx *ui.Context) {
        save()
    }),
)
```

实现上可以让 `OnClick(func(ctx))` 订阅新的 `click` 事件，并丢弃事件参数。这样旧 API 不破坏，新 API 能拿到完整数据。

## 事件分发算法

目标算法：

1. 布局期间注册可命中 target：target id、父子关系、clip/bounds、listener 列表、默认行为。
2. Gio 输入事件到达后，由 FluxUI adapter 转成 `PointerEvent`、`KeyboardEvent`、`WheelEvent` 等。
3. 对 pointer/wheel 事件执行 hit test，找到 target；对 keyboard/input/focus 事件使用 FocusManager 的 focused target。
4. 构造 event path：target 到 root，再考虑 overlay、portal、dialog modal boundary。
5. 捕获阶段从 root 到 target parent 调用 capture listener。
6. 目标阶段调用 target listener。
7. 若 `Bubbles` 为 true，冒泡阶段从 target parent 到 root 调用 bubble listener。
8. 调用 listener 时设置 `CurrentTarget` 和 `Phase`。
9. `StopPropagation` 阻止后续 target 的传播，`StopImmediatePropagation` 同时阻止当前 target 的后续 listener。
10. `Once` listener 调用后移除。
11. `Passive` listener 内调用 `PreventDefault` 不生效，并可在开发模式记录 warning。
12. 若事件可取消且没有 `DefaultPrevented`，执行默认行为。
13. 记录调试信息：事件类型、target path、listener 数量、是否取消、是否触发默认行为、耗时。

这一层必须明确 Gio 的限制：Gio 事件依赖每帧注册的 `event.Op`、clip 和 tag。FluxUI 不能假设有浏览器 DOM 那样长期存在的真实节点；事件 target 注册要和 frame 生命周期绑定，跨 frame 的稳定性由 `TargetID` 和组件路径维护。

## 默认行为清单

| 事件 | 默认行为 |
| --- | --- |
| `pointerdown` | 可请求焦点、开始 press state、开始 pointer capture |
| `pointermove` | 更新 hover、drag、slider 拖动、tooltip 位置 |
| `pointerup` | 结束 press/drag，可能合成 `click` |
| `click` | 触发 Button/Pressable 激活、checkbox/switch 切换 |
| `contextmenu` | 打开组件菜单；如果被取消则不显示默认菜单 |
| `wheel` | ScrollView 滚动 |
| `keydown` | 文本输入、焦点移动、Enter/Space 激活、局部 shortcut |
| `beforeinput` | 应用文本编辑变更；被取消则不修改文本 |
| `focus` / `blur` | 更新 focus state、样式、无障碍语义 |
| `dragover` | 被取消或显式接受后，目标可接收 drop |
| `drop` | 交付 payload 给 DropTarget |

默认行为要由组件注册到事件系统中，不要散落在各组件内部隐式执行。这样开发者才能通过事件取消机制稳定接管行为。

## 分阶段路线图

### Phase 0: 盘点和兼容边界

目标：冻结当前事件 API 的事实清单，避免后续重构误伤。

状态：已完成，记录日期 2026-07-04。P0 只更新本文档，不修改运行时代码和组件行为。

任务：

- [x] 为 `event`、`widget.Button`、`ClickArea`、`Input`、`ScrollView`、`Slider`、`DragSource`、`DropTarget` 建立事件能力矩阵。
- [x] 明确旧 API 保持兼容：`OnClick(ctx)`、`OnHover(ctx, bool)`、`InputOnChange`、`ScrollOnChange` 不删除。
- [x] 梳理 Gio 可提供的原始数据字段，并标注哪些字段当前后端不可用。
- [x] 增加文档中的“escape hatch”：高级用户可临时直接用 `ctx.Gtx.Event`，但推荐迁移到新事件层。

验收：

- [x] 文档列出每个组件当前事件和计划事件。
- [x] 没有运行时代码行为变化。

#### P0 兼容结论

以下 API 属于兼容边界，Phase 1+ 不得删除或改变签名；后续只能把它们桥接到新事件系统：

- `event.ClickHandler func(ctx *internal.Context)`
- `event.HoverHandler func(ctx *internal.Context, hovering bool)`
- `widget.OnClick(fn event.ClickHandler) ButtonOption`
- `widget.OnHover(fn event.HoverHandler) ButtonOption`
- `widget.ClickArea(child Widget, onClick func(ctx *internal.Context), opts ...ClickAreaOption) Widget`
- `widget.Pressable(child Widget, onClick func(ctx *internal.Context), opts ...PressableOption) Widget`
- `widget.InputOnChange(fn func(ctx *internal.Context, value string)) InputOption`
- `widget.InputOnFocus(fn func(ctx *internal.Context, focused bool)) InputOption`
- `widget.ScrollOnChange(fn func(ctx *internal.Context, x, y float32)) ScrollOption`
- `widget.SliderOnChange(fn func(ctx *internal.Context, value float32)) SliderOption`
- `widget.SliderOnRangeChange(fn func(ctx *internal.Context, start, end float32)) SliderOption`
- `widget.DragSourceOnEvent(fn func(ctx *internal.Context, event DragSourceEvent)) DragSourceOption`
- `widget.DragSourceOnRequest(fn func(ctx *internal.Context, event DragSourceEvent)) DragSourceOption`
- `widget.DropTarget(child Widget, onDrop func(ctx *internal.Context, event DropEvent), opts ...DropTargetOption) Widget`
- `widget.DropTargetOnActiveChange(fn func(ctx *internal.Context, event DropTargetStateEvent)) DropTargetOption`
- `widget.DropTargetOnError(fn func(ctx *internal.Context, event DropEvent)) DropTargetOption`

`ClickArea` 当前已标记为 deprecated，并作为 `Pressable` 的兼容别名保留。高级事件系统可以优先完善 `Pressable` / `PointerArea`，但不能移除 `ClickArea` 入口。

#### P0 组件事件能力矩阵

| 范围 | 当前公开事件 API | 当前公开数据 | 当前内部/底层数据 | Phase 1+ 计划事件 | 兼容要求 |
| --- | --- | --- | --- | --- | --- |
| `event` 包 | `UseClickable(ctx)`、`Clicked(ctx)`、`Hovered()`、`Pressed()`、`Focused(ctx)`、`Snapshot(ctx, includeFocus)`、`Dispatcher{Click, Hover}` | `ctx`、hovering bool、hovered/pressed/focused snapshot | `internal.ClickableState` 包装 Gio `widget.Clickable`；`History()` 内部可拿 Gio `Press` 历史，但没有作为公开事件对象暴露 | `Event`、`PointerEvent`、`KeyboardEvent`、`ListenerOptions`、capture/target/bubble 分发 | 保留 `ClickHandler`、`HoverHandler` 和现有 `Clickable` 状态 API |
| `widget.Button` | `OnClick(fn)`、`OnHover(fn)`、`ButtonAttachRef(ref)` | click 只传 `ctx`；hover 只传 `ctx` 和 bool | 使用 `event.UseClickable`；内部有 hovered/pressed/focused 用于样式；不公开坐标、按钮、修饰键、点击次数 | `click`、`pointerdown/up/enter/leave`、`focus/blur/focusin/focusout`、Enter/Space 键盘激活 | `OnClick`、`OnHover`、`ButtonRef.Click()` 行为不变，后续作为新事件的语义桥接 |
| `widget.ClickArea` / `Pressable` | `Pressable(child, onClick)`、deprecated `ClickArea(child, onClick)`、`PressableAttachRef` / `ClickAreaAttachRef` | click 只传 `ctx` | 使用 `event.UseClickable` 和 `ctx.LayoutClickArea`；不公开 hover、pressed、坐标、按钮 | `click`、`pointerdown/up/move/enter/leave`；可作为 `PointerArea` 的简单语义版本 | `ClickArea` 继续作为旧代码兼容别名；`Pressable` 保持无默认视觉样式 |
| `widget.Input` | `InputOnChange(ctx, value)`、`InputOnFocus(ctx, focused)`、`InputAttachRef` 的 `SetText/Append/Clear/Focus/Blur` | value string、focused bool | Gio `widget.Editor.Update` 产生 `ChangeEvent`、`SubmitEvent`、`SelectEvent`；`ctx.Gtx.Focused(editor)` 可查焦点；外部点击失焦使用 `pointer.Press` 和位置命中 | `beforeinput`、`input`、`change`、`submit`、`compositionstart/update/end`、`selectionchange`、`keydown/keyup`、focus events | `InputOnChange` 和 `InputOnFocus` 不删除；程序化 ref 修改继续触发当前兼容回调 |
| `ScrollView` | `ScrollOnChange(ctx, x, y)`、`ScrollAttachRef` | 合成滚动位置 `x/y`，由 list `First/Offset` 转换 | Gio list position 可得到 `First`、`Offset`、`BeforeEnd`；当前不公开原始 wheel/touchpad delta | `wheel` 暴露 delta、position、modifiers；`scroll` 暴露滚动状态；`PreventDefault` 可阻止默认滚动 | `ScrollOnChange` 继续表示滚动位置变化，不改成 wheel delta |
| `Slider` | `SliderOnChange(ctx, value)`、`SliderOnRangeChange(ctx, start, end)`、`SliderAttachRef` | value 或 range start/end | 内部直接消费 `pointer.Press/Release/Move/Drag/Enter/Leave/Cancel`；保存 `PointerID`、pressed、active handle；使用 `Buttons` 和 `Position` 更新值 | 通过 `pointerdown/move/up/cancel` 和 pointer capture 实现；对外补 `input/change` 风格值事件 | 现有 value/range 回调和 ref 命令保持不变 |
| `DragSource` | `DragSourceOnEvent`、`DragSourceOnRequest` | `DragSourceEvent{Kind, Type, Data, Operation, Err}` | Gio `transfer.SourceFilter` 和 `gesture.Drag`；生命周期已有 started/requested/completed/cancelled | `dragstart`、`drag`、`dragend`，并和通用事件路径整合 | 现有 payload、preview、operation、onEvent/onRequest API 保持不变 |
| `DropTarget` | `onDrop(ctx, DropEvent)`、`DropTargetOnActiveChange`、`DropTargetOnError` | `DropEvent{Type, Data, Text, Paths, Operation, Err}`；`DropTargetStateEvent{Active, Types}` | Gio `transfer.TargetFilter`；支持 MIME type、max bytes、URI list/text/path 解析 | `dragenter`、`dragover`、`dragleave`、`drop`；`dragover`/`drop` 支持取消默认行为和接收协商 | `DropEvent` 和 active/error 回调保持不变，后续作为 drag/drop 事件数据层 |

#### P0 Gio 原始数据字段矩阵

| Gio 来源 | 当前可取得字段 | 当前 FluxUI 使用情况 | 当前不可用或未稳定暴露字段 | Phase 1+ 处理 |
| --- | --- | --- | --- | --- |
| `pointer.Event` | `Kind`、`Source`、`PointerID`、`Priority`、`Time`、`Buttons`、`Position`、`Scroll`、`Modifiers` | `Slider` 使用 press/move/drag/enter/leave/cancel；`Input` 用 press 做外部点击失焦；drag/drop 测试和部分工具使用 pointer filter | 浏览器 Pointer Events 的 `pressure`、`tangentialPressure`、`tiltX/Y`、`twist`、`width/height`、`isPrimary`、coalesced/predicted events 当前没有直接稳定来源 | 先公开 Gio 已有字段；高级硬件字段保留为 backend dependent，默认零值 |
| `pointer.Kind` | `Cancel`、`Press`、`Release`、`Move`、`Drag`、`Enter`、`Leave`、`Scroll` | 只有部分组件内部消费；没有统一事件类型 | `pointerover/out`、`click/dblclick/contextmenu/auxclick` 需要 FluxUI 合成 | Phase 2 建立 pointer event adapter 和 click/contextmenu 合成规则 |
| `pointer.Buttons` | primary、secondary、tertiary、quaternary、quinary bitset | `Slider` 只判断 primary 或 drag；Button/Pressable 不公开按钮 | 当前公开 API 不区分左键、右键、中键 | Phase 2 公开 `Button` / `Buttons`，右键进入 `contextmenu` |
| `pointer.Scroll` | `f32.Point` delta，可由 pointer scroll event 提供 | `ScrollView` 当前没有作为公开 wheel delta 暴露，而是公开滚动位置变化 | delta mode、精确触摸板语义、默认滚动取消语义未建模 | Phase 2 新增 `WheelEvent`，`ScrollOnChange` 继续表示位置 |
| `key.Event` | `Name`、`Modifiers`、`State` | 当前没有组件级公开键盘事件；主要由 Gio Editor 内部消费 | 浏览器 `code`、`location`、`repeat`、`isComposing` 当前没有直接等价字段 | Phase 3 新增 `KeyboardEvent`，可用字段先稳定，缺失字段标记 backend dependent |
| `key.FocusEvent` | `Focus bool` | `Input` 和 `Clickable` 通过 `ctx.Gtx.Focused(tag)` 查询焦点状态 | `relatedTarget`、focus path、focusin/focusout 冒泡不存在 | Phase 3 建 FocusManager 补齐 focus event 顺序和 related target |
| `key.EditEvent` | `Range`、`Text` | Gio Editor 内部处理；FluxUI Input 当前只收到最终 `ChangeEvent` 后的文本值 | `beforeinput` 的可取消变更点、`inputType`、composition 状态未公开 | Phase 4 在 Input 层提供 best-effort `beforeinput/input`，无法后验推导的字段文档化 |
| `widget.EditorEvent` | `ChangeEvent`、`SubmitEvent{Text}`、`SelectEvent` | `Input` 目前只公开 `ChangeEvent` 对应的 `InputOnChange`；`SubmitEvent` 和 `SelectEvent` 没有公开 | selection range、selected text、submit、composition 生命周期未公开 | Phase 4 暴露 submit、selectionchange 和 composition 相关事件 |
| `transfer.DataEvent` / transfer filters | MIME type、payload reader、target/source tag | `DropTarget` / `DragSource` 已封装为 `DropEvent` / `DragSourceEvent` | 浏览器式 dragenter/dragover/drop 传播和 `PreventDefault` 接收协商未建模 | Phase 5 把现有数据层接入通用 drag/drop 事件 |

#### P0 Escape Hatch

在高级事件层完成前，高级用户可以直接使用 Gio input routing 处理原始事件：

```go
type rawPointerArea struct {
    tag   any
    child ui.Widget
}

func (w *rawPointerArea) Layout(ctx *ui.Context) layout.Dimensions {
    if w.tag == nil {
        w.tag = new(struct{})
    }

    for {
        ev, ok := ctx.Gtx.Event(pointer.Filter{
            Target: w.tag,
            Kinds:  pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Scroll,
        })
        if !ok {
            break
        }
        pe := ev.(pointer.Event)
        _ = pe.Position
        _ = pe.Buttons
        _ = pe.Modifiers
    }

    dims := w.child.Layout(ctx.Child(0))
    clipStack := clip.Rect(image.Rectangle{Max: dims.Size}).Push(ctx.Gtx.Ops)
    gioevent.Op(ctx.Gtx.Ops, w.tag)
    clipStack.Pop()
    return dims
}
```

这条 escape hatch 有明确代价：

- 调用方必须自己维护稳定 tag、clip/hit area、坐标空间、pointer grab、焦点和每帧注册。
- 直接读取 `ctx.Gtx.Event` 会绕开后续 FluxUI 的 capture/bubble、default action、diagnostics 和兼容桥接。
- 新事件层完成后，应用代码应迁移到 `PointerArea`、`KeyboardScope` 或 `On(event.Type, ...)`，只在框架尚未覆盖的后端能力上继续使用底层 Gio。

### Phase 1: EventTarget 和基础分发

目标：先建立事件系统骨架，不急着覆盖全部输入类型。

状态：已完成，记录日期 2026-07-04。P1 已落地基础事件对象、frame 级 target/listener 注册表、三阶段分发和 `click` 兼容桥接。

任务：

- [x] 新增 `event.TargetID`、`event.Type`、`event.Event`、`event.Phase`。
- [x] 在 runtime 中维护当前 frame 的 target tree、listener registry、event path。
- [x] 支持 capture/target/bubble 三阶段。
- [x] 支持 `StopPropagation`、`StopImmediatePropagation`、`PreventDefault`、`DefaultPrevented`。
- [x] 支持 `ListenerOptions{Capture, Once, Passive}`。
- [x] 支持 synthetic dispatch，用于测试和应用自定义事件。
- [x] 让现有 `OnClick` 可以由新 `click` 事件桥接，但保持原签名。

验收：

- [x] 单元测试覆盖 listener 调用顺序、停止传播、once 移除、passive 禁止取消、synthetic dispatch 返回值。
- [x] 旧按钮点击测试继续通过。

#### P1 实现记录

新增公开事件 API：

- `event.TargetID`
- `event.Type`
- `event.Click`
- `event.Phase`、`event.PhaseNone`、`event.PhaseCapture`、`event.PhaseTarget`、`event.PhaseBubble`
- `event.Event`
- `event.ListenerOptions`
- `event.Handler`
- `event.Capture()`、`event.Once()`、`event.Passive()`、`event.Priority(priority)`
- `event.On(ctx, type, handler, opts...)`
- `event.Dispatch(ctx, target, event)`
- `event.DispatchEvent(ctx, target, *event)`

`event.Event` 当前支持：

- `Type`、`Target`、`CurrentTarget`、`Phase`、`Time`
- `Bubbles`、`Cancelable`、`DefaultPrevented`、`Trusted`、`Detail`
- `StopPropagation()`
- `StopImmediatePropagation()`
- `PreventDefault() bool`
- `PropagationStopped() bool`
- `ImmediatePropagationStopped() bool`
- `ComposedPath() []TargetID`

Runtime 侧新增 `runtimeEventState`，每帧由 `BeginFrame` 重置：

- `RegisterEventTarget(target, parent)` 记录当前 frame 的 target tree。
- `RegisterEventListener(ctx, type, handler, options)` 记录当前 frame listener。
- `DispatchEvent(ctx, target, *Event)` 按 target path 执行 capture、target、bubble。
- `Context.NewContext`、`Context.Child`、`Context.Scope` 会把 root/child/scope target 注册进当前 frame。

兼容桥接：

- `event.Dispatcher.DispatchClick(ctx)` 现在会先分发 `event.Click` synthetic event。
- 该 click event 默认 `Bubbles=true`、`Cancelable=true`、`Trusted=true`。
- 如果 listener 调用 `PreventDefault()`，旧的 `OnClick(ctx)` 回调不会执行。
- 如果没有被取消，旧的 `OnClick(ctx)` 继续按原签名执行。

P1 暂不做的内容：

- 不实现 pointer、wheel、keyboard、focus 的专用事件结构。
- 不做 hit test adapter；P1 的 dispatch 只支持 synthetic dispatch 或现有 clickable 语义桥接。
- `Once` 当前按 frame 内注册表移除；声明式组件每帧重新注册 listener 的长期 once 语义留到后续 API 稳定后再定义。

验证命令：

```sh
go test ./event ./internal ./widget -count=1
```

### Phase 2: Pointer 和 Wheel

目标：解决鼠标/触摸/笔输入数据不透明的问题。

状态：已完成，记录日期 2026-07-04。P2 已落地浏览器式 pointer/wheel 事件对象、PointerArea/PointerAreaElement、右键/中键/双击合成、基础 pointer capture 和 move coalescing；Slider 已开始复用公开 pointer event 转换层。

任务：

- [x] 新增 `PointerEvent`、`WheelEvent`、`Button`、`Buttons`、`Modifiers`。
- [x] 新增 `PointerArea` / `PointerAreaElement`，允许开发者在任意区域监听原始 pointer 事件。
- [x] 支持 `pointerdown/up/move/enter/leave/over/out/cancel`。
- [x] 支持 `click`、`dblclick`、`auxclick`、`contextmenu` 合成。
- [x] 支持 `wheel`，暴露 delta 和 position。
- [x] 支持基础 pointer capture：按下后继续把同一 pointer id 的 move/up 发给 capture target。
- [x] 在高频 move 上提供 coalesced 读取策略，至少保证最后一次 move 不丢。

验收：

- [x] 开发者能在回调里拿到坐标、按钮、修饰键、pointer id、滚轮 delta。
- [x] 右键菜单和中键点击不再只能通过 Gio escape hatch 实现。
- [x] Slider 可以逐步迁移到公开 pointer event 层，而不丢现有行为。

#### P2 实现记录

新增公开事件 API：

- `event.PointerEvent`、`event.WheelEvent`、`event.PointerSample`
- `event.PointerType`、`event.Button`、`event.Buttons`、`event.Modifiers`
- `event.PointerDown`、`event.PointerUp`、`event.PointerMove`、`event.PointerEnter`、`event.PointerLeave`、`event.PointerOver`、`event.PointerOut`、`event.PointerCancel`
- `event.DblClick`、`event.AuxClick`、`event.ContextMenu`、`event.Wheel`
- `event.OnPointer`、`event.OnWheel`、`event.DispatchPointerEvent`、`event.DispatchWheelEvent`
- Gio adapter：`event.PointerEventFromGio`、`event.WheelEventFromGio`、`event.TypeFromGioPointerKind`、`event.ButtonsFromGio`、`event.ModifiersFromGio`

新增组件 API：

- `widget.PointerArea(child, opts...)`
- `widget.PointerAreaDisabled`、`widget.PointerAreaPassThrough`、`widget.PointerCaptureOnPress`
- `widget.PointerOnDown/Up/Move/Enter/Leave/Over/Out/Cancel`
- `widget.PointerOnClick`、`widget.PointerOnDoubleClick`、`widget.PointerOnAuxClick`、`widget.PointerOnContextMenu`、`widget.PointerOnWheel`
- `ui.PointerArea` / `ui.PointerAreaElement` 以及同名 option re-export

Runtime 变化：

- `internal.Runtime` 保留 pointer capture 表，`SetPointerCapture` / `ReleasePointerCapture` / `PointerCaptureTarget` / `HasPointerCapture` 可被事件层使用。
- capture target 使用当前 frame 的 target tree 和 listener registry 分发；若捕获目标当帧没有注册 listener，则不会额外伪造回调。

Slider 迁移记录：

- `widget.Slider` 的内部 pointer 处理已经先调用 `event.TypeFromGioPointerKind` 和 `event.PointerEventFromGio`，再读取公开的 `PointerEvent.Position`、`PointerEvent.Buttons`、`PointerEvent.PointerID` 更新值。
- 旧 `SliderOnChange` / `SliderOnRangeChange` 行为保持不变。

边界：

- 当前 Gio 后端稳定提供 `PointerID`、`PointerType(mouse/touch)`、`Buttons`、`Position`、`Scroll`、`Modifiers` 和事件时间偏移。
- `PointerPen`、pressure、tilt、twist、width/height、predicted events 暂无稳定后端数据，字段保留但默认零值。
- `WheelEvent.DeltaMode` 当前为 pixel；精确触摸板语义、默认滚动取消和 ScrollView 默认行为迁移留到 Phase 5。

### Phase 3: Focus、Keyboard 和局部快捷键

目标：建立组件树内键盘事件系统，而不是只依赖 Input 或系统全局快捷键。

状态：已完成，记录日期 2026-07-04。P3 已落地组件树内 FocusManager、focus/keyboard typed event、KeyboardScope/KeyboardScopeElement、局部 shortcut registry，以及 Button/Pressable 的 Enter/Space 默认激活桥接。

任务：

- [x] 新增 FocusManager：当前 focus target、tab order、focus request、blur、disabled/hidden target 处理。
- [x] 新增 `KeyboardEvent`，支持 `keydown`、`keyup`、repeat、key/code、location、modifiers、isComposing。
- [x] 新增 `KeyboardScope` / `KeyboardScopeElement`。
- [x] 新增局部 shortcut API，限定在当前 focus scope 或显式 scope 内。
- [x] Enter/Space 激活 Button/Pressable；Escape 可被 Dialog/Menu/Popup 捕获。
- [x] 明确 global shortcut 与 component keyboard event 的边界。

验收：

- [x] 焦点移动产生正确 `focus/blur/focusin/focusout` 顺序。
- [x] `keydown` 可取消默认按钮激活或焦点移动。
- [x] 局部快捷键只在目标 scope 内触发，不污染系统快捷键。

#### P3 实现记录

新增公开事件 API：

- `event.FocusEvent`、`event.KeyboardEvent`、`event.KeyLocation`、`event.FocusDirection`
- `event.Focus`、`event.Blur`、`event.FocusIn`、`event.FocusOut`
- `event.KeyDown`、`event.KeyUp`
- `event.FocusManagerFor(ctx)`、`event.RequestFocus(ctx)`、`event.BlurFocus(ctx)`、`event.Focused(ctx)`、`event.FocusedTarget(ctx)`、`event.MoveFocus(ctx, direction)`
- `event.RegisterFocusTarget(ctx, opts...)`
- `event.FocusTabIndex(index)`、`event.FocusDisabled(disabled)`、`event.FocusHidden(hidden)`、`event.FocusActivate(fn)`
- `event.OnFocus`、`event.OnKeyboard`、`event.OnKeyDown`、`event.OnKeyUp`
- `event.Shortcut`、`event.ShortcutKey`、`event.ShortcutCode`、`event.ShortcutExactModifiers`、`event.ShortcutScope`、`event.OnShortcut`
- Gio adapter：`event.KeyboardEventFromGio`

新增组件和 `ui` API：

- `widget.KeyboardScope(child, opts...)` / `ui.KeyboardScope(child, opts...)`
- `widget.KeyboardScopeDisabled`、`KeyboardScopeFocusable`、`KeyboardScopeTabIndex`、`KeyboardScopeAutoFocus`
- `widget.KeyOn/KeyOnDown/KeyOnUp`
- `widget.FocusOn/FocusOnFocus/FocusOnBlur/FocusOnIn/FocusOnOut`
- `widget.ShortcutOn`
- `ui.KeyboardScopeElement`
- `ui` 同名 re-export；焦点查询在 `ui` 中命名为 `IsFocused(ctx)`，避免和既有装饰 API `Focused(decoration)` 冲突。

Runtime 变化：

- `internal.Runtime` 在事件状态中维护当前 focus target、当前 frame 的 focus target 表、tab order、keydown repeat 状态和局部 shortcut registry。
- `RegisterFocusTarget` 使用当前 `PathID` 作为 focus target；disabled/hidden target 不进入 tab order，若当前焦点目标在 frame 结束时不可用，会自动 blur。
- `RequestFocus` 按 `blur` -> `focusout` -> `focus` -> `focusin` 顺序分发；`focus` / `blur` 不冒泡，`focusin` / `focusout` 冒泡，并填充 `RelatedTarget`。
- `DispatchKeyboardEvent` 先按现有 capture/target/bubble 分发 `keydown` / `keyup`，再在未 `PreventDefault` 时执行局部 shortcut 和默认行为。
- `Tab` 默认移动焦点；`Enter` / `Space` 默认激活当前 focus target。

兼容桥接：

- `Button` 每帧注册为 focus target；Enter/Space 默认激活调用现有 `event.Dispatcher.DispatchClick(ctx)`。
- `Pressable` / `ClickArea` 的 pointer click、ref click 和 keyboard 默认激活统一走 `click` 兼容桥；旧 `onClick(ctx)` 签名不变。
- `keydown.preventDefault()` 会阻止默认按钮激活和默认焦点移动。

边界：

- Gio `key.Event` 当前稳定提供 `Name`、`Modifiers`、`State`；`KeyboardEvent.Key` 归一化常见键名，`Code` 先使用 Gio `Name` 的字符串形式。
- `Repeat` 由 FluxUI runtime 根据连续 keydown 状态推导；`location`、`isComposing` 暂无 Gio 等价后端数据，字段保留但默认零值。
- `KeyboardScope` 处理组件树内 keyboard event；`system.RegisterGlobalShortcut` 仍是 OS-level 全局快捷键，两者不共享 registry，也不会互相触发。
- Escape 捕获通过 `KeyboardScope` / `KeyOnDown` 完成；Dialog/Menu/Popup 可在迁移阶段使用该层拦截 Escape 并 `StopPropagation` 或 `PreventDefault`。

验证命令：

```sh
go test ./event ./internal ./widget ./ui ./examples/docs_browser -count=1
```

### Phase 4: Text Input、IME 和编辑事件

状态：已完成，记录日期 2026-07-04。P4 已落地文本输入 typed event、composition typed event、Input 事件 option、`beforeinput` best-effort 取消回滚、`InputRef` 程序化变更来源标记，以及 Gio `SubmitEvent` 到 `submit` 的桥接。

目标：把文本编辑从“值变化回调”升级为可拦截、可组合、可解释的输入事件。

任务：

- [x] 新增 `beforeinput`、`input`、`change`、`submit`。
- [x] 新增 `CompositionEvent` 和 `compositionstart/update/end`。
- [x] 为 `Input` 暴露 `InputOnBeforeInput`、`InputOnInputEvent`、`InputOnSubmit`。
- [x] 区分用户输入、程序调用 `InputRef.SetText`、粘贴、删除、撤销重做。
- [x] 若 Gio 无法完整支持 mutation-before 拦截，先提供 best-effort 实现并在文档中标注。

验收：

- [x] 能实现“只允许数字输入”“Enter 提交”“IME 组合中不提前提交”等常见需求。
- [x] 富文本/代码编辑器方向有明确可扩展入口。

#### P4 实现记录

公开事件 API：

- `event.InputEvent`、`event.CompositionEvent`、`event.InputSource`。
- 事件类型常量：`BeforeInput`、`Input`、`Change`、`Submit`、`CompositionStart`、`CompositionUpdate`、`CompositionEnd`。
- 来源和编辑类型常量：`InputSourceUser`、`InputSourceProgrammatic`、`InputSourcePaste`、`InputSourceDelete`、`InputSourceUndo`、`InputSourceRedo`，以及 `insertText`、`insertFromPaste`、`deleteContentBackward`、`historyUndo`、`historyRedo`、`insertLineBreak`、`programmaticSetText` 等 input type。
- typed helper：`event.OnInput`、`event.OnComposition`、`event.DispatchInputEvent`、`event.DispatchCompositionEvent`。

`Input` / `ui` 层 API：

- `widget.InputOnBeforeInput`、`widget.InputOnInputEvent`、`widget.InputOnSubmit`。
- `ui.InputOnBeforeInput`、`ui.InputOnInputEvent`、`ui.InputOnSubmit`。
- `ui` 重新导出 `InputEvent`、`CompositionEvent`、`InputSource` 和相关常量，方便 React-style Element 与旧 Widget API 共用。

`Input` 桥接行为：

- Gio `ChangeEvent` 先分发 cancelable `beforeinput`。若被 `PreventDefault`，回滚 `gioWidget.Editor` 文本到上一帧同步值，不继续分发 `input/change/InputOnChange`。
- 若 `beforeinput` 未取消，则依次分发 `input`、`change`，最后继续调用旧 `InputOnChange(ctx, value)`，保持兼容。
- `InputRef.SetText/Append/Clear` 标记为 `InputSourceProgrammatic`，input type 分别为 `programmaticSetText`、`programmaticAppend`、`programmaticClear`；旧 ref 变更仍会触发 `InputOnChange`。
- 用户输入的 `Source` / `InputType` 通过前后文本 diff 和本地历史 best-effort 推断：常见插入、粘贴、删除、撤销、重做都会给出可用分类。
- `InputOnSubmit` 会启用 Gio `Editor.Submit`，并把 `SubmitEvent{Text}` 映射为 cancelable `submit`，`Value` 为提交文本，`InputType` 为 `insertLineBreak`。

边界和后续扩展入口：

- Gio 0.9 的 `widget.Editor` 当前主要暴露 mutation-after 的 `ChangeEvent` / `SubmitEvent` / `SelectEvent`；FluxUI 的 `beforeinput` 是回滚型 best-effort，推断字段会设置 `BestEffort=true`。
- 粘贴、删除方向、撤销/重做的精确原始原因不能完全从 Gio `ChangeEvent` 还原；当前实现覆盖常见场景，富文本/代码编辑器后续应在 `key.EditEvent` 级别补一个真正 mutation-before、可取消的 adapter。
- `CompositionEvent` API 和 synthetic dispatch 已可用，但 Gio `Editor` 没有向 FluxUI 暴露稳定的 composition 生命周期，所以真实 IME composition 事件暂不自动生成，`InputEvent.IsComposing` 默认 `false`。
- “IME 组合中不提前提交”当前含义是 submit 只响应 Gio `SubmitEvent`；当后端暴露 composing 状态后，应让 `IsComposing=true` 的 Enter/submit 可取消或延迟。

验证命令：

```sh
go test ./event ./internal ./widget ./ui ./examples/docs_browser -count=1
```

### Phase 5: 组件默认行为统一迁移

状态：已完成，记录日期 2026-07-04。P5 已把核心组件的默认交互迁移到可传播、可取消的事件层，并保留旧回调签名。

目标：让核心组件的交互行为都通过事件系统表达。

任务：

- [x] Button/Pressable/ClickArea 迁移到 `click` 默认行为和新事件参数。
- [x] Checkbox/Switch/Radio/Select 通过可取消的 `click` 或 `keydown` 激活。
- [x] ScrollView 使用 `wheel` 默认行为，并支持 `PreventDefault`。
- [x] Slider 使用 pointer capture 和 pointer events。
- [x] Menu/Dialog/Popup 使用 focus、keyboard、pointer outside 事件。
- [x] DragSource/DropTarget 接入 `drag*` / `drop` 事件流。

验收：

- [x] 开发者能在父容器捕获子组件点击，或阻止事件冒泡。
- [x] 可在特定区域拦截滚轮、右键、键盘事件而不破坏其他组件。
- [x] 旧 API 仍可用，且行为与迁移前一致。

#### P5 实现记录

公开事件 API：

- 新增 `event.DragEvent`、`event.DragHandler`、`event.OnDrag`、`event.DispatchDragEvent`。
- 新增 drag/drop 事件类型：`dragstart`、`drag`、`dragenter`、`dragover`、`dragleave`、`drop`、`dragend`。
- `ui` 重新导出 `Event`、`EventType`、`EventPhase`、`DragEvent`、`DragHandler`、`OnEvent`、`DispatchEvent`、`DispatchEventPtr`、`OnDrag`、`DispatchDragEvent`、`Capture`、`Once`、`Passive`、`EventPriority` 和 `Event*` 常量。

默认行为迁移：

- `Button`、`Pressable`、`ClickArea` 的旧 `OnClick(ctx)` 现在由 cancelable `click` 事件桥接；事件 payload 是 `PointerEvent`，包含 position、button、modifiers、click count 等可用数据。
- `Checkbox`、`Switch`、`RadioGroup`、`Select`、`Menu`、`DropdownMenu` 的点击激活先派发 `click`，未被 `PreventDefault` 时才执行旧的 onChange/onSelect/onOpenChange 默认行为。
- `ScrollView` 为滚轮注册 `wheel` 目标，先派发 cancelable `WheelEvent`；未取消时再执行默认滚动，并继续保留 `ScrollOnChange(ctx, x, y)`。
- `Slider` / `RangeSlider` 在 pointer down/move/up/cancel 上派发 `PointerEvent`，按下后使用 pointer capture，事件被取消时不更新值；旧 `SliderOnChange` / `SliderOnRangeChange` 继续保留。
- `DragSource` 把开始、请求、完成、取消映射到 `dragstart` / `drag` / `dragend`；`DropTarget` 把 Gio transfer 生命周期映射到 `dragenter` / `dragover` / `drop` / `dragleave`，并把 MIME type、payload、text、paths、operation、types、error 放入 `DragEvent`。
- `Select`、`DropdownMenu` 的外部点击关闭会先派发 cancelable `pointerdown`；调用方可通过 `PreventDefault` 阻止关闭。Dialog/Popup 的 Escape 仍通过 P3 的 `KeyboardScope` / `KeyOnDown` 捕获，后续 Phase 6 继续定义 modal/portal 边界规则。

兼容边界：

- `OnClick(ctx)`、`OnHover(ctx, bool)`、`InputOnChange`、`ScrollOnChange` 不删除、不改签名。
- Gio transfer 的 OS 级拖放协商仍由 Gio 后端负责；FluxUI 的 `PreventDefault` 作用于组件默认行为和旧回调桥接，不声明已经替代平台级 drag accept/drop negotiation。
- `WheelEvent.DeltaMode` 当前仍按 pixel 处理；精确触摸板/平台滚动语义后续继续细化。

新增测试覆盖：

- `event`：drag/drop typed dispatch 默认值、传播顺序、`PreventDefault`、`dragleave/dragend` 不可取消。
- `ui`：通用事件导出和 `DispatchDragEvent` synthetic dispatch。
- 既有目标包测试覆盖旧按钮点击、pointer/wheel、keyboard/focus、text input 以及核心 widget 编译路径。

验证命令：

```sh
go test ./event ./internal ./widget ./ui ./examples/docs_browser -count=1
```

### Phase 6: 自定义事件、Portal 和边界策略

目标：让事件系统能支撑复杂应用架构。

任务：

- 支持应用自定义事件类型和 `Detail` payload。
- 支持 `DispatchEvent` 到指定 target。
- 定义 overlay、portal、dialog、modal 的 event path 和 retarget 规则。
- 支持事件边界组件，例如 `EventBoundary`，用于截断或重定向传播。
- 为无障碍语义预留 activation event 和 keyboard equivalent。

验收：

- Popup 内点击可以按规则冒泡到 owner 或被 modal boundary 截断。
- 应用可发自定义事件做组件间通信，但文档明确不替代状态管理。

状态：已完成。

完成记录：

- `event`：
  - 新增 `NewCustomEvent`、`DispatchCustomEvent`、`CustomBubbles`、`CustomCancelable`、`CustomTrusted`。
  - 自定义事件直接使用 `Event.Type` 字符串和 `Event.Detail` payload；`Detail` 不做框架解释，适合少量组件间交互语义，不替代状态管理。
  - 新增 `Activate` / `ActivationEvent` / `OnActivate` / `DispatchActivationEvent`，预留无障碍 activation event 与 `KeyboardEquivalent`。
  - 新增 `BoundaryMode`、`BoundaryPolicy`、`RegisterBoundary`、`RegisterPortal`、`BoundaryStopPropagation`、`BoundaryRedirectTo`。
- `internal`：
  - `Runtime` 的 target tree 现在支持 target 级 boundary policy。
  - `EventBoundaryStop`：`ComposedPath()` 截断在 boundary target，capture/bubble 不再到达外层 owner/root。
  - `EventBoundaryRedirect`：事件到达 boundary 后跳转到指定 target，再继续沿该 target 的逻辑父链传播。
  - `RegisterEventPortal`：overlay/portal root 的逻辑事件父级可指向 owner target，而不依赖视觉 layout parent。
- `widget`：
  - 新增 `EventBoundary`、`EventBoundaryDisabled`、`EventBoundaryStopPropagation`、`EventBoundaryRedirectTo`。
  - 新增 `EventPortal(child, owner)`，用于把 visually mounted elsewhere 的子树挂回 owner 事件路径。
  - `Dialog` panel 注册为 portal + modal stop boundary：内部事件停止在 dialog portal root。
  - `Popup` panel 注册为普通 portal：内部事件按逻辑路径冒泡到 popup owner。
- `ui`：
  - 导出 custom event、activation event、boundary/portal API。
  - 新增 `EventBoundaryElement`、`EventPortalElement`。

边界规则：

- `Event.Target` 仍表示原始 dispatch target；本阶段不实现 Shadow DOM 式逐 listener 动态 retarget。
- `ComposedPath()` 是 FluxUI 的逻辑 target path：portal 会把路径接回 owner；modal boundary 会截断路径。
- `EventPortal` 只改变事件路径，不改变 Gio input hit test、绘制层级或焦点模型。
- 自定义事件适合表达“用户/组件交互语义已经发生”，不用于同步持久状态；跨组件状态仍应使用 state/provider/store。

验证命令：

```sh
go test ./event ./internal ./widget ./ui -count=1
```

### Phase 7: 性能、诊断和文档完善

目标：让高级事件系统可观测、可调优、可教学。

任务：

- 开发模式输出事件路径、listener 耗时、取消状态和默认行为。
- 扩展现有 interaction diagnostics，覆盖 pointer/keyboard/focus/wheel。
- 为高频 pointer/wheel 建 benchmark，避免每帧大量分配。
- 为 docs browser 新增 Event System 指南和示例。
- 新增示例：画布拖拽、右键菜单、局部快捷键、富输入拦截、复杂拖放。

验收：

- 高频 pointer move 示例无明显分配热点。
- 文档能说明“什么时候用简单 OnClick，什么时候用高级事件”。

P7 完成记录：

- 运行时诊断：
  - `internal.PerfDiagnostics` 新增 `LogEvents`。
  - `ui.LogEvents(enabled)` / `app.LogEvents(enabled)` 可在开发模式输出事件诊断。
  - `FrameStats.Events` 记录 dispatch 数量、listener 调用次数、listener 总耗时、取消状态、传播停止状态、分类计数和最后一次事件摘要。
  - 事件日志输出 `type`、`target`、`path`、`listeners`、`listener_duration`、`default_prevented`、`propagation_stopped`、`immediate_stopped`、`default_allowed`。
- interaction diagnostics：
  - `InteractionFrameStats` 扩展 `PointerEvents`、`WheelEvents`、`KeyboardEvents`、`FocusEvents`。
  - pointer / wheel / keyboard / focus 事件在 perf diagnostics 开启时进入统一统计。
- 性能：
  - `widget/pointer_area_bench_test.go` 新增 `BenchmarkPointerAreaHighFrequencyMove` 和 `BenchmarkPointerAreaHighFrequencyWheel`。
  - `PointerArea` 复用 Gio pointer event drain buffer 和 coalesced sample buffer，listener dispatch 不再为匹配列表额外分配。
- docs browser：
  - 新增 `docs/guides/event-system.md` 指南，明确简单 `OnClick` / `OnHover` / `InputOnChange` / `ScrollOnChange` 与高级事件层的使用边界。
  - 新增 `examples/docs_browser/event_system_demo.go`，完整展示 pointer capture、coalesced move、wheel、右键菜单、局部快捷键、focus、`beforeinput`、`input`、`submit`、custom event、capture/bubble、drag/drop。
  - `event_system_basic` 已接入 docs browser demo registry、presentation 和文档元数据回归测试。

验证命令：

```sh
go test ./event ./internal ./widget ./ui ./app ./examples/docs_browser -count=1
go test ./widget -bench "BenchmarkPointerAreaHighFrequency(Move|Wheel)$" -benchmem -run ^$ -benchtime=100x
```

## 能力矩阵

| 能力 | 当前状态 | 目标状态 |
| --- | --- | --- |
| 点击 | 有 `OnClick(ctx)` | `click` 事件含 target、phase、position、button、modifiers、click count |
| Hover | 有 `OnHover(ctx, bool)` | `pointerover/out` 冒泡，`pointerenter/leave` 不冒泡 |
| Pressed | 组件内部状态 | `pointerdown/up/cancel` 公开 |
| 鼠标坐标 | 大多不公开 | pointer/wheel/click 事件统一公开 |
| 鼠标按钮 | 大多不公开 | primary/secondary/auxiliary/buttons bitset |
| 右键菜单 | 缺少统一 API | `contextmenu` 可取消默认菜单 |
| 双击 | 缺少统一 API | `dblclick` 或 `ClickCount` |
| 滚轮 | 只有滚动位置变化 | `wheel` delta + 默认滚动 |
| 键盘 | 只有 Input 和 global shortcut | 组件级 `keydown/keyup` + focus scope |
| 文本输入 | `InputOnChange` | `beforeinput/input/change/composition/submit` |
| 焦点 | 局部组件能力 | 统一 FocusManager + focus events |
| 拖放 | 数据模型较完整 | 纳入通用 `drag*` / `drop` 传播 |
| 传播 | 无 | capture/target/bubble |
| 取消默认行为 | 无 | cancelable event + default action registry |
| 自定义事件 | 无 | synthetic/custom event dispatch |
| Listener options | 无 | capture/once/passive/priority |

## API 命名建议

为了兼容现有风格，建议分三层：

1. 简单组件语义 API：继续保留 `OnClick`、`InputOnChange`、`ScrollOnChange`。
2. 类型化事件 API：提供 `OnPointerDown`、`OnKeyDown`、`OnWheel`、`OnFocusIn` 等常用选项。
3. 通用事件 API：提供 `On(event.Type, handler, opts...)` 和 `DispatchEvent`，用于高级场景和自定义事件。

示例：

```go
ui.ContainerElement(
    child,
    ui.On(event.PointerDown, func(ctx *ui.Context, ev *event.PointerEvent) {
        if ev.Button == event.ButtonSecondary {
            ev.PreventDefault()
            openContextMenu(ev.Position)
        }
    }),
)
```

```go
ui.KeyboardScopeElement(
    editor,
    ui.OnKeyDown(func(ctx *ui.Context, ev *event.KeyboardEvent) {
        if ev.Modifiers.Control() && ev.Key == "s" {
            ev.PreventDefault()
            save()
        }
    }),
)
```

## 非目标

- 不实现完整浏览器 DOM、Shadow DOM、CSSOM 或 Web API。
- 不承诺所有平台都能提供压力、倾斜、预测事件等高级硬件数据。
- 不把事件系统变成状态管理系统；自定义事件只用于输入/交互语义和少量组件通信。
- 不一次性替换所有组件；迁移必须分阶段、兼容旧 API。
- 不让开发者必须理解 Gio 才能使用高级事件，但文档应说明 Gio frame/input routing 对实现的约束。

## 测试策略

- 分发顺序测试：capture -> target -> bubble。
- 停止传播测试：`StopPropagation` 与 `StopImmediatePropagation`。
- 取消测试：cancelable、non-cancelable、passive listener。
- listener 生命周期测试：once、移除、同一 target 多 listener 顺序。
- pointer 测试：press/move/up、enter/leave、right click、wheel、capture。
- keyboard 测试：keydown/keyup 顺序、repeat、modifiers、focus target、局部 shortcut。
- input 测试：beforeinput 取消、change、submit、composition 状态。
- 组件兼容测试：旧 `OnClick`、`InputOnChange`、`ScrollOnChange` 行为不变。
- 性能测试：高频 pointer move、wheel、拖拽场景不产生不可控分配。

## 完成定义

高级事件系统达到可用状态时，应满足：

- 开发者可以在任意区域监听 pointer、wheel、keyboard、focus、input、drag/drop 事件。
- 事件对象包含足够的原始数据，不需要常规场景回退到 `ctx.Gtx.Event`。
- 事件能沿组件树捕获和冒泡，父组件能治理子组件事件。
- 可取消事件有明确默认行为，`PreventDefault` 结果稳定可测。
- 现有简单 API 完全兼容，并可逐步由新事件系统驱动。
- 文档、示例、测试覆盖常见高级场景：右键菜单、画布拖拽、局部快捷键、输入拦截、复杂拖放。
