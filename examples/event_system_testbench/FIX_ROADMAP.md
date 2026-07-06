# Event System 测试反馈与修复路线图

记录日期：2026-07-06

来源：`examples/event_system_testbench` 手工体验反馈。

目标：先冻结问题清单和修复顺序，再逐项修底层事件系统。修复方向优先放在 runtime / widget 底层，避免要求开发者在每个业务组件里额外绕路。

## 当前问题清单

### P0：兼容 API 与基础滚动

1. 旧 `ScrollOnChange` 的 offset 一直是 `0,0`。
   - 现象：测试区域已经可以滚动，但日志和界面显示的 offset 不更新。
   - 风险：旧 API 兼容承诺不成立，开发者无法依赖 `ScrollOnChange(ctx, x, y)` 判断当前位置。
   - 初步判断：`ScrollView` 内部滚动状态与公开 change callback 使用的 offset 没有同步，或 wheel 默认行为执行后没有把最终滚动位置回调出去。

2. 单行输入框阻挡页面纵向滚动。
   - 现象：鼠标停在单行 `Input` 上时，页面纵向滚轮失效。
   - 风险：表单和长页面混合时滚动体验不可用。
   - 初步判断：`Input` 注册了 Gio pointer/scroll filter 后消费了 wheel，但单行输入框本身没有可滚动的纵向默认行为，也没有把未处理的 wheel 交还给父级 `ScrollView`。

3. 拖拽源组件也会截停竖向滚动。
   - 现象：鼠标停在 `DragSource` 上时，页面纵向滚动被截断。
   - 风险：可拖拽列表或素材面板无法自然滚动。
   - 初步判断：`DragSource` 为 pointer/transfer 注册输入区域时也接管了 scroll/wheel，未区分 drag gesture 与普通 wheel。

### P1：测试台可理解性不足

1. P1 测试区域看不出测试目的。
   - 现象：用户无法直观看懂 capture / target / bubble、once、passive、stop、preventDefault 的含义和操作预期。
   - 风险：即使底层正确，也无法通过手工测试得到有效反馈。
   - 初步判断：示例偏 API 验证，缺少可视化事件路径、预期结果和“点击后应该看到什么”的说明。

### P2：PointerArea 命中区域过大

1. 指针事件测试区域影响整行。
   - 现象：P2 的 pointer 测试区域不只覆盖视觉面板，而是影响到横向整行空间。
   - 风险：任意透明/空白布局区域都可能抢事件，导致旁边控件无法交互或滚动。
   - 初步判断：`PointerAreaElement` 的可命中尺寸来自父布局分配尺寸，而不是视觉 child 的实际绘制/期望尺寸；或者 `ExpandedElement` / row 布局给了过大的 hit rect。

### P3：Focus / Keyboard 全部无效

1. 焦点移动、keydown/keyup、局部快捷键、Enter/Space 默认激活均无效。
   - 现象：P3 区域所有测试项不响应。
   - 风险：Phase 3 的核心目标未达成，组件树内键盘事件系统不可用。
   - 初步判断：
     - `KeyboardScope` 可能没有稳定获得 Gio key input focus。
     - `RegisterFocusTarget` 的 component focus 与 Gio `key.FocusOp` 没有桥接。
     - `processKeyboardScopeEvents` 只在 focused target 位于 scope 内时 drain key event，但事件源可能没有路由到该 scope。
     - pointer 点击请求 component focus 后，没有同步请求底层键盘焦点。

### P4：Text Input 正常

1. 当前 P4 反馈正常。
   - 结论：`beforeinput`、`input`、`change`、`submit`、程序化 `InputRef` 和 composition synthetic 入口暂不作为首要修复对象。
   - 注意：P4 仍依赖 P3 后续补齐真实 IME/key focus 的稳定性。

### P5：默认行为迁移与弹窗

1. 阻止默认事件全部无效。
   - 现象：父容器 capture listener 中调用 `PreventDefault` 后，Button / Pressable / Checkbox / Switch / Radio / Select / ScrollView / Slider 等默认行为仍继续执行。
   - 风险：Phase 5 的“组件默认行为可取消”核心验收不成立。
   - 初步判断：
     - 组件旧回调可能仍直接由 Gio clickable / editor / slider 状态触发，没有等待新事件层 dispatch 的返回值。
     - `click` / `wheel` / `pointerdown` 事件已派发，但默认行为执行点没有统一检查 `DefaultPrevented`。
     - 部分组件可能在子组件内部自己触发旧 `onChange`，父级 capture 的取消结果没有传回默认行为层。

2. 打开 Dialog / Popup 导致程序未响应。
   - 现象：点击打开弹窗后应用卡住或无响应。
   - 风险：overlay / portal / boundary 事件路径可能存在循环、重复布局、阻塞式重绘或事件递归。
   - 初步判断：
     - `Dialog` / `Popup` 的 portal + modal boundary 注册可能形成事件 parent 循环。
     - overlay 内部 pointer outside / focus / keyboard 默认行为可能在同一帧反复触发 openChange。
     - 测试台中 P5 父级 capture listener 与弹窗全局 overlay 组合后可能触发递归关闭/重开。

3. Select 下拉选项上滚轮导致鼠标失焦，鼠标不移动时点击全部不响应。
   - 现象：在可滚动下拉菜单选项上滚动后，只要鼠标位置不变，后续点击无法响应。
   - 风险：这是严重交互缺陷，会影响所有 Select/Menu/DropdownMenu。
   - 初步判断：
     - 下拉菜单内部 scroll/wheel 后，hover/click target 缓存没有刷新。
     - wheel 改变了 overlay 内容位置，但 hit test 或 Gio pointer target 仍停留在旧位置。
     - 鼠标不移动时没有新的 enter/move 事件触发 target 更新，导致点击派发到失效 target。
     - 可滚动菜单需要在 scroll 后主动使 overlay pointer hit state 失效，或保证 click 按当前布局重新 hit test。

### P6 / P7：暂未直接反馈失败，但受前置问题影响

1. P6 自定义事件、boundary、portal 的真实体验会被 Dialog/Popup 卡死问题影响。
2. P7 诊断目前可用于观察，但需要增加更清晰的 GUI 诊断视图，避免只依赖控制台。

## 底层问题归类

### A. Wheel 事件所有权和向父级传递不清晰

涉及：
- 单行 `Input`
- `DragSource`
- `Select` / DropdownMenu 内部滚动
- `ScrollView`
- 旧 `ScrollOnChange`

修复原则：
- 只有真正能处理 wheel 默认行为的组件才消费 wheel。
- 子组件不能处理该方向滚动时，wheel 必须允许父级 `ScrollView` 继续处理。
- 横向滚动只响应横向 delta；纵向滚动只响应纵向 delta；不要用竖向滚轮误触横向滚动。
- overlay 内部滚动后必须刷新 pointer hit / hover / click target 状态。

### B. 事件命中区域使用布局尺寸而不是可视/声明区域

涉及：
- P2 `PointerArea` 影响整行
- 透明空白区域抢事件
- 复杂 row / expanded 布局内的 pointer event 归属

修复原则：
- `PointerArea` 默认 hit rect 应等于 child 实际布局尺寸。
- 如果父布局分配了更大空间，只有显式 fill/passThrough 策略才能扩大 hit rect。
- Pointer 注册、clip、layout dimensions 三者需要一致。

### C. Component focus 与 Gio keyboard focus 未打通

涉及：
- P3 全部无效
- Enter/Space 激活
- Escape 捕获
- 局部 shortcut
- 后续 IME / composition 的真实输入可靠性

修复原则：
- `KeyboardScope` / focus target 请求 component focus 时，同步请求 Gio key focus。
- key event 的 drain target 必须稳定，不依赖鼠标 hover。
- Tab 默认行为、shortcut 和 keydown cancel 顺序要可测试。

补充：tab组件的focus正常，但是仅仅是作为全局的

### D. 默认行为没有统一由事件 dispatch 返回值控制

涉及：
- P5 `PreventDefault` 全部无效
- Button/Pressable/ClickArea
- Checkbox/Switch/Radio/Select
- ScrollView
- Slider

修复原则：
- 所有组件默认行为统一走“先派发 cancelable event，再根据 `Dispatch*` 返回值执行默认行为”。
- 父级 capture listener 的 `PreventDefault` 必须能阻止子组件默认行为。
- 旧 API 只作为默认行为后的兼容回调，不绕过新事件层。

### E. Overlay / Portal / Boundary 生命周期可能存在递归或路径循环

涉及：
- Dialog / Popup 打开后未响应
- Select 下拉菜单滚动后 target 失效
- modal boundary 截断和 popup owner 冒泡

修复原则：
- portal event parent 必须有循环检测。
- overlay open/close 不能在同一帧重复触发互相矛盾的默认行为。
- popup/menu/dialog 内部点击、滚动、空白区域和遮罩点击必须有明确边界。

## 修复路线图

### R1：先修 wheel 传递与 ScrollView 状态同步

优先级：最高。

任务：
- 梳理 `ScrollView` 的 wheel 默认行为、offset 状态、`ScrollOnChange` 触发点。
- 修复旧 `ScrollOnChange` offset 一直为 `0,0`。
- 为 `Input` 单行模式增加 wheel pass-through：单行输入框不消费纵向 wheel。
- 为 `DragSource` 增加 wheel pass-through：未进入拖拽手势时不截停纵向 wheel。
- 为下拉菜单内部 scroll 增加滚动后 hit/hover target 刷新。

验收：
- P0 `ScrollOnChange` 显示真实 y offset。
- 鼠标停在单行输入框上，页面仍能纵向滚动。
- 鼠标停在拖拽源上，页面仍能纵向滚动。
- Select 下拉菜单内滚动后，不移动鼠标也能继续点击选项。
- 横向滚动组件不会被纯竖向滚轮触发。

### R2：修 PointerArea 命中区域

优先级：高。

任务：
- 检查 `PointerArea` 注册 clip rect 使用的 size 是否来自 child 实际尺寸。
- 检查 Row/Expanded 下 child dimensions 与 pointer registration 的关系。
- 增加一个可视边界测试：面板外同一行空白处不应触发 pointer event。

验收：
- P2 指针测试只影响视觉面板区域。
- 面板右侧/同一行空白区域可以正常滚动和点击其他控件。

### R3：修 FocusManager 与 KeyboardScope

优先级：高。

任务：
- 为 `KeyboardScope` 或 focus target 注册 Gio key focus op。
- 点击 focus target 时同时请求 component focus 和 Gio key focus。
- 确认 `processKeyboardScopeEvents` 能收到 keydown/keyup。
- 修正 focus/blur/focusin/focusout 顺序。
- 确认 keydown `PreventDefault` 能阻止 Tab 默认焦点移动和 Enter/Space 默认激活。

验收：
- P3 点击焦点块后，keydown/keyup 日志正常。
- Tab / Shift+Tab 能在可聚焦项之间移动，跳过 disabled/hidden。
- 开启“阻止 Tab 默认”后焦点不移动。
- Ctrl+K 只在局部 scope 内触发。
- Escape 能在 scope 内停止传播。
- Enter/Space 能触发 focused target 的 activation。

### R4：统一组件默认行为的取消语义

优先级：高。

任务：
- 审计 Button / Pressable / ClickArea 的 `click` 桥接，确认旧 `OnClick` 只在 `click` 未取消后调用。
- 审计 Checkbox / Switch / Radio / Select 的点击和键盘激活路径。
- 审计 ScrollView 的 wheel 默认行为，确认父级 capture `PreventDefault` 生效。
- 审计 Slider 的 pointerdown/move/up 默认行为，确认取消后不更新值。
- 补齐跨组件的默认行为测试。

验收：
- P5 开启“阻止 click 默认”后，Button/Pressable/ClickArea/Checkbox/Switch/Radio/Select 不执行旧回调或状态变化。
- P5 开启“阻止 wheel 默认”后，ScrollView 不滚动。
- Slider 在相关 pointer event 被取消时不改变值。

### R5：修 Overlay / Portal / Boundary 的卡死与菜单滚动失焦

优先级：高，需在 R1/R4 后复测。

任务：
- 检查 Dialog/Popup 打开时的 portal parent 和 modal boundary 是否形成循环。
- 检查 outside pointer、mask close、focus trap、keyboard Escape 是否同帧互相触发。
- 对 Select/Menu/DropdownMenu 的滚动菜单做专门修复：滚动后重新计算 hover/click target，或使旧 target 失效。
- 增加 overlay 内部点击、空白点击、组件点击、遮罩点击、滚轮滚动的分场景测试。

验收：
- P5 打开 Dialog/Popup 不再未响应。
- 点击 Dialog/Popup 内部按钮、空白和组件不会误触发外部关闭。
- Select 下拉选项滚动后，鼠标不移动也能点击当前指向的选项。
- Popup 内点击可按规则冒泡到 owner，Dialog/modal 内事件被 boundary 截断。

### R6：重做 P1 测试台表达

优先级：中。

任务：
- 把 P1 改成可视化事件路径：父容器、目标容器、子按钮三层。
- 每次 dispatch 后显示调用顺序，例如 `父 capture -> 目标 capture -> 目标 bubble -> 父 bubble`。
- 为 once/passive/stop/preventDefault 分别提供独立按钮和“预期结果”。
- 避免把多个概念塞进同一个按钮。

验收：
- 不看源码也能理解 P1 在测什么。
- 每个按钮点击后都有明确预期和实际结果。

### R7：补测试与诊断

优先级：中。

任务：
- 为 R1-R5 每个修复补单元测试或最小集成测试。
- 扩展 event diagnostics GUI：显示 event type、target、path、phase、defaultPrevented。
- 为 wheel pass-through、overlay scroll 后点击、keyboard focus bridge 增加回归测试。

验收：
- 手工测试台通过。
- 自动测试覆盖本次所有已反馈问题。
- 控制台和 GUI 诊断能定位后续事件路径问题。

## 推荐修复顺序

1. R1：wheel 传递和 `ScrollOnChange` offset。
2. R2：PointerArea 命中区域。
3. R3：Focus / Keyboard 桥接。
4. R4：默认行为取消语义。
5. R5：Dialog / Popup / Select overlay 边界和滚动失焦。
6. R6：P1 测试台 UX 重做。
7. R7：补自动测试和诊断视图。

原因：R1 是多个体验问题的共同底层；R2 会影响后续所有 pointer 结果可信度；R3 是 P3 的硬阻塞；R4/R5 建立在事件路径、滚轮和焦点稳定之后再修更稳。

## 当前暂不作为首要问题

- P4 文本输入反馈正常，暂不进入首批修复。
- P6/P7 没有独立失败反馈，但需要在 overlay/focus/default 行为修复后重新验收。

## 后续记录格式

每个修复 PR 或提交建议记录：

- 修复项：例如 `R1-Input-wheel-pass-through`
- 触达文件：runtime/widget/ui/example/test
- 修复说明：不超过 5 行
- 手工验收：对应 P0-P7 哪个区域
- 自动测试：新增或更新的测试名称
- 残留风险：例如平台 wheel delta 差异、Gio 后端限制

## R1 修复记录（2026-07-06）

- 修复项：`R1-wheel-propagation-and-scroll-offset`
- 触达文件：`widget/list_grid.go`、`widget/input.go`、`widget/drag_source.go`、`widget/list_grid_test.go`
- 修复说明：
  - `ScrollOnChange` 改为回传真实像素 offset，旧 P0 示例不再因为 `offset/1024` 被格式化成 `0`。
  - 单行 `Input` 的 Gio `Editor` pointer 注册改为 pass-through，让自身仍可编辑/聚焦，同时不独占父级纵向 wheel。
  - `DragSource` 的 drag/transfer 注册改为 pass-through，未进入拖拽手势时不截停父级滚动。
  - `ListView` / `GridView` 在滚动位置变化后主动请求下一帧，用于刷新 Select/Menu/Dropdown 这类可滚动 overlay 的 hit/hover/click 目标。
- 手工验收：对应 P0 单行输入框、拖拽源、ScrollOnChange；对应 P5 下拉选项滚动后不移动鼠标继续点击。
- 自动测试：`TestScrollViewWheelScrollsContent`、`TestScrollViewWheelOverSingleLineInputScrollsOuter`、`TestScrollViewWheelOverDragSourceScrollsOuter`，并复跑横向/嵌套滚动与 Select/Menu 虚拟列表相关测试。
- 残留风险：不同平台 wheel delta 方向和精度可能不同；Select/Menu 的真实 overlay 点击刷新仍建议继续用测试台手工复验。

## R2 修复记录（2026-07-06）

- 修复项：`R2-pointer-area-hit-rect`
- 触达文件：`widget/pointer_area.go`、`widget/pointer_area_test.go`
- 修复说明：
  - `PointerArea` 布局 child 时清除父级 `Min` 约束，只保留父级最大约束，避免 Column/Row/Exact 父布局把普通 child 强行拉成整行后扩大事件命中区。
  - `PointerArea` 的 Gio pointer clip 仍使用 child 实际返回尺寸；显式 `FillWidth` / `Fill` / `Expanded` 仍可主动扩大命中区。
  - 新增 exact 父约束回归测试：父布局给 320x120，child 为 120x80 时，x=200 的同一行空白不触发 pointerdown，x=60 的面板内点击正常触发。
- 手工验收：对应 P2 指针事件测试区域只影响视觉面板区域；面板右侧/同一行空白区应可继续滚动和点击其他控件。
- 自动测试：`TestPointerAreaHitRectUsesRelaxedChildSize`，并复跑 `TestPointerArea*`、`TestScrollViewWheel*`、`go test ./widget -count=1`。
- 残留风险：如果已有调用依赖“PointerArea 包普通 child 时自动吃满父级最小宽度”的隐式行为，需要改为显式 `FillWidth` / `Fill`，这与路线图中“默认 hit rect 等于 child 实际布局尺寸”的目标一致。

## R3 修复记录（2026-07-06）

- 修复项：`R3-focus-keyboard-bridge`
- 触达文件：`internal/events.go`、`event/keyboard.go`、`widget/keyboard_scope.go`、`widget/keyboard_scope_test.go`
- 修复说明：
  - runtime 在 component focus target 注册时，为当前聚焦目标建立 1x1 Gio key-focus anchor，并同步执行 `key.FocusCmd`，避免无边界 `event.Op` 破坏 pointer/wheel 路由。
  - `KeyboardScope` 改为监听当前 component focused target 的 Gio key focus tag，并允许 Ctrl/Command/Shift/Alt/Super 修饰键进入通用 key filter。
  - Gio 字母键转换为浏览器风格 `KeyboardEvent.Key`：未按 Shift 时返回小写，保证 `ShortcutKey("k", Ctrl)` 这类局部快捷键可用。
  - 保持 runtime 原有 focus/blur/focusin/focusout 顺序和 synthetic keyboard dispatch 兼容。
- 手工验收：对应 P3 点击焦点块后 keydown/keyup 日志、Tab/Shift+Tab 焦点移动、阻止 Tab 默认、Ctrl+K 局部快捷键、Enter/Space 默认激活。
- 自动测试：新增 `TestKeyboardScopeReceivesRouterKeyEventsForFocusedTarget`，覆盖真实 Gio router keydown/keyup、局部 Ctrl+K、Enter 激活、Tab `PreventDefault`；并复跑 keyboard/event 与 widget 全量测试。
- 残留风险：component focus 与 Gio focus 在首次请求后可能需要下一帧完成底层同步；这符合 Gio handler 从上一帧 ops 建立的模型，测试台交互路径应不受影响。

## R4 修复记录（2026-07-06）

- 修复项：`R4-cancelable-default-actions`
- 触达文件：`internal/context.go`、`ui/element.go`、`ui/element_event_test.go`、`widget/event_defaults_test.go`、`widget/list_grid_test.go`、`widget/slider.go`、`widget/slider_interaction_test.go`
- 修复说明：
  - 为 Element 渲染输出增加 layout context 绑定，让构建期注册的 `OnEvent(..., Capture())` 进入真实布局事件路径，测试台 P5 父级 capture 不再挂在 build 树孤岛上。
  - 保持 `Button` / `Pressable` / `ClickArea` / `Checkbox` / `Switch` / `Radio` / `Select` 通过 `click` dispatch 返回值决定旧回调或状态默认行为是否执行。
  - `ScrollView` 的 `wheel` 默认行为继续以 `DispatchWheelEvent` 返回值为准，并新增父级 capture `PreventDefault` 回归测试。
  - `Slider` 在 `pointerdown` 被取消时不进入拖动；拖动中的 `pointermove` 被取消时立即结束本次手势并释放 pointer capture。
- 手工验收：对应 P5 打开“阻止 click 默认”后，Button/Pressable/ClickArea/Checkbox/Switch/Radio/Select 不执行旧回调或状态变化；打开“阻止 wheel 默认”后 ScrollView 不滚动；Slider pointer 默认行为可被取消。
- 自动测试：新增 `TestClickDefaultPreventedByParentCapture`、`TestScrollViewWheelPreventDefaultStopsDefaultScroll`、`TestSliderPointerDownPreventDefaultStopsDefaultChange`、`TestElementOnEventCaptureCanPreventChildButtonDefault`；复跑 `go test ./event`、`go test ./widget`、`go test ./ui`。
- 残留风险：R4 只处理默认行为取消语义；Dialog/Popup/Select overlay 生命周期、内部点击误关闭和滚动后失焦仍按 R5 单独修复。

## R5 修复记录（2026-07-06）

- 修复项：`R5-overlay-portal-boundary-and-scroll-hit-refresh`
- 触达文件：`widget/tabs_dialog_toast.go`、`widget/selection.go`、`widget/interactive_layout_test.go`
- 修复说明：
  - Dialog / Popup 的视觉遮罩不再使用全屏 `Pressable` 承担关闭逻辑，改为 pass-through outside-press guard，并用居中面板矩形保护内部按钮、空白和组件点击。
  - Dialog 继续在 portal root 注册 modal boundary；Popup 继续通过 portal 冒泡到 owner，不额外截断内部事件。
  - `SelectQuick(true)` 现在真正跳过下拉动画，避免测试和快速交互中 popup 命中区滞后。
  - Select / DropdownMenu 继续复用 R1 的 ListView 滚动后重绘刷新机制，并补齐“滚动后不移动鼠标直接点击”的回归测试。
- 手工验收：对应 P5 打开 Dialog/Popup 不再未响应；点击弹窗内部按钮、空白、组件不误关闭；点击遮罩仍可关闭；Select / DropdownMenu 选项滚动后不移动鼠标也能点击当前指向项。
- 自动测试：新增 `TestMaskClosableOverlayInternalComponentDoesNotClose`、`TestOpeningMaskClosableOverlayDoesNotCloseFromOpeningClick`、`TestSelectPopupWheelThenClickWithoutMoveUsesUpdatedOption`、`TestDropdownMenuWheelThenClickWithoutMoveUsesUpdatedItem`，并复跑 `TestModalInternalPressDoesNotCloseMaskClosableOverlay`。
- 验证命令：`go test ./widget -count=1`、`go test ./event ./ui -count=1`、`go test ./examples/event_system_testbench ./examples/docs_browser ./examples/component_lab -run ^$`。
- 残留风险：outside dismiss 现在按 pointer press 触发，而不是等待完整 click release；这与 Select/Dropdown 现有 outside press 行为一致，但如果后续需要拖拽穿过遮罩再释放关闭，可以再增加按 release 关闭的配置。
