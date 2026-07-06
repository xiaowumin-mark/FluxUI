# A6.2 wheel 分发和父子滚动审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：滚动、手势和嵌套交互审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Layout
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A6\\.2|wheel|Wheel|Scroll|DragSource|Slider|Select|docs content" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "processWheelEvents|registerWheelTarget|applyWheelDefault|scrollViewClampOffset|ScrollRange|TestNestedScroll|TestVerticalWheelOverHorizontal|TestHorizontalScrollViewRequires|TestListViewWheel|TestScrollViewWheelScrolls" widget/list_grid.go widget/list_grid_test.go`
  - `rg -n "PointerOnWheel|pointer\\.Scroll|DispatchWheelEvent|PointerAreaPassThrough|registerPointerArea|drainPointerAreaEvents" widget/pointer_area.go examples/docs_browser/event_system_demo.go`
  - `rg -n "drag\\.Add|gesture\\.Both|pointer\\.Scroll|func \\(s \\*dragSourceState\\) update|DragSourceElement" widget/drag_source.go examples/docs_browser/event_system_demo.go examples/docs_browser/drag_drop_demo.go examples/component_lab/main.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/input.go`
  - `widget/list_grid.go`
  - `widget/list_grid_test.go`
  - `widget/pointer_area.go`
  - `widget/input.go`
  - `widget/slider.go`
  - `widget/drag_source.go`
  - `widget/selection.go`
  - `widget/material3_components.go`
  - `examples/docs_browser/docs_browser_app.go`
  - `examples/docs_browser/right_panel.go`
  - `examples/docs_browser/demo_presentation.go`
  - `examples/docs_browser/markdown_renderer.go`
  - `examples/docs_browser/event_system_demo.go`
  - `examples/component_lab/main.go`
- 关联能力：
  - wheel target 注册区域和轴向过滤
  - ScrollView wheel default action 和 `PreventDefault` gate
  - 父子滚动命中优先级和横纵轴传递
  - Select/Menu/ListView 选项滚动入口
  - Input、Slider、DragSource 对 wheel 的影响边界
  - docs content scroll、代码块和表格横向滚动入口

## 事实结论

1. FluxUI 的 typed `WheelEvent` 来自 Gio `pointer.Scroll`，`WheelEventFromGio` 直接映射 `Scroll.X`、`Scroll.Y` 为 `DeltaX`、`DeltaY`，`DeltaMode` 固定为 pixel，事件默认 `Bubbles=true`、`Cancelable=true`。证据：`event/input.go:215`、`event/input.go:249`。
2. `ScrollView.Layout` 在内容布局前调用 `processWheelEvents`，再执行 `layoutContent`；因此 wheel 默认行为先更新 `state.list.Position.Offset`，随后本帧布局会按新 offset 重新计算和绘制内容。证据：`widget/list_grid.go:162`、`widget/list_grid.go:256`。
3. `ScrollView` 的 wheel target 注册在 viewport 尺寸矩形内，并包在 `pointer.PassOp` 下；它不应阻塞同一区域的 child click 或其他 pointer target。证据：`widget/list_grid.go:284`、`widget/list_grid_test.go:116`。
4. `ScrollView.processWheelEvents` 对横向和纵向使用不同 Gio `pointer.Filter`：横向只接受 `ScrollX`，纵向只接受 `ScrollY`。这不是剩余 delta 算法，而是输入路由层的轴向过滤。证据：`widget/list_grid.go:295`、`widget/list_grid.go:305`。
5. wheel 事件先经过 `DispatchWheelEvent`；若 listener 调用 `PreventDefault` 使 dispatch 返回 false，`ScrollView` 不执行 `applyWheelDefault`。证据：`widget/list_grid.go:320`。
6. 默认滚动只消费当前主轴 delta：纵向取 `DeltaY`，横向取 `DeltaX`；非主轴 delta 不会由该 `ScrollView` 转换为 offset。证据：`widget/list_grid.go:328`。
7. `applyWheelDefault` 会累计小数残量并把整数像素加到 `Position.Offset`，但它不计算“本次实际消费像素”和“未消费剩余 delta”；触顶或触底后的 delta 只会在后续 `layoutContent` 里 clamp 掉。证据：`widget/list_grid.go:338`、`widget/list_grid.go:358`。
8. 现有嵌套滚动行为依赖 Gio 命中路由和 `pointer.PassOp`，不是 FluxUI 自己维护父子 scroll chain。内层命中时，符合内层轴向 filter 的 wheel 由内层处理；不符合内层轴向 filter 的 wheel 可被外层相应轴向容器收到。
9. 已有测试覆盖纵向 wheel 落在横向内层 ScrollView 上时，外层纵向 ScrollView 会滚动，内层横向 ScrollView 不滚动。证据：`widget/list_grid_test.go:343`。
10. 已有测试覆盖内层 `ScrollView` 和内层 virtualized `ListView` 在指针下优先处理同轴 wheel，外层 `ScrollView` 不同时处理。证据：`widget/list_grid_test.go:534`、`widget/list_grid_test.go:601`。
11. `ListView` 和 `GridView` 使用 Gio `layout.List` 自身的滚动输入；FluxUI 没有为它们注册 `WheelEvent`、`PreventDefault` gate 或 `ScrollOnChange`。证据：`widget/list_grid.go:711`、`widget/list_grid.go:888`。
12. `Select` 打开菜单后使用 `ListView` 构建选项列表，并根据 `SelectMaxHeight` 限制弹出层高度；因此大量选项的 wheel 行为来自 Gio `layout.List`，不是 `ScrollView` 的 wheel default action。证据：`widget/selection.go:654`、`widget/selection.go:392`。
13. `Menu` 在估算高度超过 `MenuMaxHeight` 时用 `FixedHeight(maxHeight, ListView(...))` 包裹选项；`DropdownMenu` 也把 placement 的 `MaxHeightPx` 回写为菜单最大高度。证据：`widget/material3_components.go:492`、`widget/material3_components.go:875`。
14. `PointerArea` 默认 `passThrough=true`，会在 child layout 尺寸内注册全轴 `pointer.Scroll`，并把 wheel 分发到 FluxUI event system；它只派发事件，不执行滚动默认行为。证据：`widget/pointer_area.go:49`、`widget/pointer_area.go:201`、`widget/pointer_area.go:274`。
15. `PointerArea` 的 wheel listener 可以调用 `PreventDefault`，但该默认阻止只影响收到该 typed wheel 的默认 action；普通 `PointerArea` 自身没有 default action，也没有剩余 delta 传递实现。docs event system demo 就显式调用 `ev.PreventDefault()` 作为演示。证据：`examples/docs_browser/event_system_demo.go:115`。
16. `Input` 没有注册 `pointer.Scroll` listener；它把 Gio `Editor` 事件映射到 `InputEvent`，因此不会由 FluxUI widget 层无条件截停父级页面 wheel。证据：`widget/input.go:283`。
17. `Slider` 只注册 press/release/move/drag/enter/leave/cancel，不注册 scroll；竖向父滚动不会因为 slider hover 被 FluxUI slider 代码截停。证据：`widget/slider.go:462`、`widget/slider.go:473`。
18. `DragSource` 使用 Gio `gesture.Drag` 和 transfer source filter；代码没有注册 `pointer.Scroll`，所以未拖拽时不应吞掉 wheel。拖拽期间是否由 Gio gesture/backend 抢占滚动属于后端行为，当前 FluxUI 代码没有显式 wheel 规则。证据：`widget/drag_source.go:208`、`widget/drag_source.go:351`。
19. docs browser 右侧文档正文主体是 virtualized `ListView`，不是外层 `ScrollView`；代码块和 Markdown 表格内部再使用横向 `ScrollView`，并设置 `ScrollVertical(false)`。证据：`examples/docs_browser/right_panel.go:72`、`examples/docs_browser/markdown_renderer.go:193`、`examples/docs_browser/markdown_renderer.go:260`。
20. docs demo viewport 对部分示例再包纵向 `ScrollView`，其中 `slider_basic`、`event_system_basic`、`scroll_view_basic` 等是后续手动验收嵌套 wheel 的入口。证据：`examples/docs_browser/demo_presentation.go:54`、`examples/docs_browser/demo_presentation.go:58`。

### wheel target 和父子滚动矩阵

| 场景 | wheel target | 可滚动判断 | 剩余 delta | 父级传递规则 | 结论 |
| --- | --- | --- | --- | --- | --- |
| 纵向 `ScrollView` | viewport 内的 `state.wheelTag` | Gio filter 只收 `ScrollY`；内容是否可滚动由后续 clamp 决定 | 无显式剩余值；触顶/触底被 clamp | 同轴内层命中时外层通常不同时处理；非同轴可由外层收到 | 轴向明确，边界剩余未建模 |
| 横向 `ScrollView` | viewport 内的 `state.wheelTag` | Gio filter 只收 `ScrollX` | `DeltaY` 不进入该 target；`DeltaX` 触边后无剩余 | 纵向 wheel 可落到外层纵向容器 | 已有测试覆盖 |
| `ListView` / `GridView` | Gio `layout.List` 内部 target | Gio 自己维护 Position 和边界 | FluxUI 不可见 | 由 Gio 路由决定 | 无 FluxUI `PreventDefault` gate |
| `Select` / `Menu` popup | popup 内部 `ListView` | `MaxHeight` 后形成可滚动列表 | FluxUI 不可见 | 父文档滚动是否继续依赖 Gio 路由 | 需手动验收触底传递 |
| `PointerArea` | child layout 尺寸内全轴 target | 不判断可滚动 | 不计算 | 默认 pass-through，但 listener 可 `PreventDefault` | 事件演示入口，不是滚动容器 |
| `Input` | Gio Editor 内部 target | FluxUI 不处理 wheel | 不适用 | FluxUI wrapper 不截停 | 未发现 widget 层硬截停 |
| `Slider` | track pointer target，无 scroll kinds | 不适用 | 不适用 | 父级 wheel 不应被 slider 拦截 | 未发现 widget 层硬截停 |
| `DragSource` | Gio drag source target，无 scroll filter | 不适用 | 不适用 | 未拖拽时不应拦截 wheel | 拖拽期间后端行为待验收 |
| docs 代码块/表格 | 内层横向 `ScrollView` | 只接受横向 delta | 纵向 delta 交给外层 | 外层文档列表可继续纵向滚动 | 已符合设计方向 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| 触顶/触底剩余 delta 未建模 | 高 | `ScrollView` 只把主轴 delta 加入 offset，再由布局阶段 clamp；没有记录实际消费量，也没有把未消费 delta 显式传给父滚动容器。内层同轴滚动到边界后，父页面继续滚动的行为未被现有测试证明。 | A6.3 横向/触摸板、A13 收敛修复 |
| `ListView`/`Menu`/`Select` 缺 FluxUI wheel gate | 中 | 这些组件依赖 Gio `layout.List`，无法通过 FluxUI `WheelEvent.PreventDefault` 统一控制默认滚动，也无法记录剩余 delta。 | A5.5 default action、A6.3、A11 文档 |
| PointerArea 示例会主动阻止 wheel 默认行为 | 中 | event system demo 的 PointerArea 在 wheel listener 中 `PreventDefault`，它适合作为事件演示，但若嵌在可滚动页面中，容易被误认为普通面板也应截停父滚动。 | A11 示例说明、手动验收 |
| docs browser 右侧主体是 ListView | 中 | 右侧文档内容不是 `ScrollView`，所以 ScrollView 的 `WheelEvent`、`PreventDefault` 和 `ScrollOnChange` 行为不能直接代表 docs content 主滚动。 | A6.3、A11 docs browser 验收 |
| DragSource 拖拽期间 wheel 冲突未覆盖 | 低 | FluxUI 代码未注册 scroll，但 Gio drag gesture 或平台 transfer 后端在拖拽中如何处理 wheel 没有自动测试。 | A10 DnD、手动 smoke |
| Select/Menu 触底父传递缺测试 | 低 | 已验证大量选项虚拟化，但没有自动测试 popup 列表滚到顶部/底部后父页面是否继续滚动。 | A6 后续测试补齐 |

## 验收

- 已明确 `ScrollView` wheel target：viewport 内 Gio tag，按横向/纵向设置 `ScrollRange`，并使用 `pointer.PassOp`。
- 已明确可滚动判断边界：`ScrollView` 的内容尺寸与 viewport 尺寸在 `layoutContent` 中 clamp；`ListView`、`Menu`、`Select` 的可滚动性由 Gio `layout.List` 决定。
- 已明确剩余 delta 现状：当前没有显式 leftover delta 数据结构或父级 scroll chain；触边后的主轴 delta 只会被 clamp 掉。
- 已明确父子传递规则：非同轴 wheel 可因轴向 filter 传给父级；同轴内层可滚动目标优先，父级不同时处理；内层触边后的同轴父传递未被证明。
- 已核对输入范围：`Input`、`Slider`、`DragSource` 没有 FluxUI 层 `pointer.Scroll` 注册，不属于无条件截停父页面滚动的来源。
- 已标出后续手动验收入口：docs browser 的代码块/表格横向滚动、`slider_basic`、`event_system_basic`、`select_basic`、`menu_basic`、`drag_source_basic`。

## 后续依赖

- A6.3 横向 delta、shift-wheel、touchpad 审查应继续复用本文的轴向过滤结论，并补充触摸板同时产生 `DeltaX/DeltaY` 时的策略。
- A10 DragSource/DropTarget 审查应补充拖拽中 wheel、父滚动和平台 transfer 后端的手动验收。
- A11 docs browser 示例审查应区分事件演示中的 `PreventDefault` 与普通内容滚动，不应把演示面板的截停行为推广为默认模式。
- 若后续要支持同轴 nested scroll chaining，需要在 `ScrollView` default action 中记录实际消费 delta、边界方向和剩余 delta，并补齐触顶/触底父传递测试。
