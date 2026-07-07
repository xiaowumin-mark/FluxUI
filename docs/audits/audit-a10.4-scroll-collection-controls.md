# A10.4 滚动集合控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.4 滚动集合控件族审查。

## 事实结论

审查范围覆盖 `ScrollView`、`ListView`、`Grid`、`GridView`、`Tabs`、chips row。核心滚动实现集中在 `widget/list_grid.go`，`TabsScrollable` 在 `widget/tabs_dialog_toast.go` 中通过横向 `ScrollView` 包装 tabs row，chips 本身位于 `widget/material3_components.go`，实际 chips row 由调用方使用 `Row` 或横向 `ScrollView` 组合。

| 控件/组合 | axis | wheel | virtualization | hit refresh | Ref / callback |
| --- | --- | --- | --- | --- | --- |
| `ScrollView` | `ScrollVertical/ScrollHorizontal` 决定主轴；仅 horizontal=true 且 vertical=false 时为横向，否则默认纵向。内容主轴上限扩到 `1_000_000`，视口尺寸来自父约束。 | 注册独立 `wheelTag`，按主轴设置 Gio `pointer.Filter`：纵向只收 `ScrollY`，横向只收 `ScrollX` 且 `ScrollY=0`。`WheelEvent` 可被 `PreventDefault` 阻止默认滚动。 | 非虚拟化，整体 child 会被布局；通过 clip 裁剪视觉输出。 | 每帧重新布局 child，并用 `WithPositionOffset(-offset)` 更新子树位置；`registerWheelTarget` 用本帧 `size` 注册命中区域。滚动后下一帧 ops/filter 与 visual position 对齐。 | `ScrollRef` 命令在下一帧 drain：to start/end/offset/by；offset 被 clamp。`ScrollOnChange` 仅在 `First/Offset` 变化后触发。 |
| `ListView` | `ListAxis` 决定 Gio `layout.List.Axis`，默认纵向。 | 自身依赖 Gio `layout.List` 的滚动输入；外层若处于 `ScrollView` 命中区域，按 Gio routing 与 FluxUI wheel target 优先级分发。 | 默认 `ListVirtualized(true)`，只构建 Gio 可见窗口；关闭虚拟化时构建全部 item，超过 `nonVirtualizedLargeItemThreshold=256` 记录诊断。 | 虚拟化 item 在本帧 Gio list layout 回调中生成；`WithViewport(viewport)` 传给可见子项。列表状态基于 Gio `Position` 跨帧保留，insert/delete 的位置身份风险已有专项测试覆盖。 | 无 `ScrollRef`；`ListOnReachEnd` 基于 Gio `Position.BeforeEnd` 和 viewport fallback 判定一次性触发。 |
| `Grid` | 静态二维布局，列数固定或由 `GridMinItemWidth` 解析，内部用 `Row` + `Column`。 | 不注册 wheel；若放在 `ScrollView` 中，由父滚动容器处理。 | 无虚拟化，所有 child 都布局。 | 命中区域来自每个 child 自身和 Row/Column 布局，不注册集合级 wheel target。 | 无滚动 Ref；只是静态布局容器。 |
| `GridView` | 固定纵向列表轴，按行虚拟化；列数可由 `GridMinItemWidth` 在当前宽度下收缩。 | 依赖 Gio `layout.List` 的纵向滚动输入。 | 按 row 虚拟化，只有可见行内 cell 调用 builder；记录 `RecordVirtualizedItems`。 | 每个可见行在 list layout 回调中重新生成，row 内 cell 使用 `rowCtx.Child(i-startIdx)`。滚动后下一帧可见 cell 的 ops/filter 重新注册。 | 无 `ScrollRef`；`GridOnReachEnd` 基于行级 `Position` 触发。 |
| `Tabs` / `TabsScrollable` | 普通 tabs row 非滚动时等分或使用可用宽度；`TabsScrollable(true)` 外包横向 `ScrollView`。 | 可滚动 tabs 只接受横向 wheel delta；纵向 wheel 不会被横向 `ScrollView` 消费。 | 无虚拟化，所有 tab item 均布局。 | tab item 自身通过 `LayoutRippleArea` 注册点击命中，横向滚动时由外层 `ScrollView` 对 row 做 position offset，下一帧命中随视觉位置更新。 | `TabsRef` 下一帧 drain key 并触发 `TabsOnChange`；没有滚动 Ref，也没有自动滚到 active tab 的实现。 |
| chips row | chip 本体只注册 click/remove click，不注册 wheel；横向滚动语义来自外层 `ScrollView(ScrollHorizontal(true), ScrollVertical(false))` 或 docs/category filter 的组合。 | chip 不消费 wheel；横向 chips row 若由横向 `ScrollView` 包装，则只消费 `DeltaX`。普通 `Row` chips 不提供滚动保护，可能横向溢出。 | 无虚拟化。 | 每个 chip 的 hit area 来自 `md3ActionSurface` 的实际布局尺寸；remove 图标使用 `ctx.Scope("remove")` 独立 click target。 | 无集合 Ref；点击和 remove 由 chip callbacks 处理。 |

滚动和命中链路可以归纳为：

1. `ScrollView.Layout` 先消费 `ScrollRef` 命令，再处理 wheel event，随后按当前 offset 重新布局 child。
2. `layoutContent` 记录 viewport，把 child context 移动到 `-offset`，并对结果做 clip；wheel target 的注册区域等于本帧滚动容器 `size`。
3. `ListView/GridView` 不复用 `ScrollRef`，它们的滚动状态属于 Gio `layout.List.Position`；虚拟化 builder 只在可见窗口内运行。
4. `TabsScrollable` 没有独立滚动实现，它直接继承横向 `ScrollView` 的轴过滤、clip 和下一帧命中刷新。
5. chips row 没有专用集合组件；是否可滚动取决于调用方是否显式包一层横向 `ScrollView`。

## 风险

- `ScrollView` 是 FluxUI 自己建模 offset 的滚动容器，`ListView/GridView` 是 Gio `layout.List` 滚动容器；两类滚动状态、offset 单位和 Ref 能力不同，后续文档不能把 `ScrollRef` 承诺扩展到 `ListView/GridView`。
- 横向 `ScrollView` 只消费 `DeltaX`，没有 shift-wheel 转横向滚动策略，也没有剩余 delta 模型；这符合 A6.3 结论，但需要在 tabs/chips/code/table 手动验收中继续标明。
- `Grid` 是静态布局，不是滚动集合；大量 child 时不会虚拟化，也不会主动限制横向溢出，必须由外层 `ScrollView`、`GridView` 或父级约束承担。
- `TabsScrollable` 构建所有 tab，适合少量 tab；大量 tab 时没有虚拟化或 active tab 自动滚入视口。
- chips row 不是一等组件，调用方直接 `Row(chips...)` 时不会自动横向滚动；docs browser 的 category chips 使用横向 `ScrollView` 是正确入口，但其他示例需要逐一确认。
- `ListView/GridView` 的 hit refresh 依赖 Gio list 在下一帧重新生成可见 ops；“滚动后不移动鼠标即可点击新目标”的自动测试主要覆盖 `ScrollView`，对 `ListView/GridView` 仍应保留回归入口。

## 验收

- 已建立 `ScrollView`、`ListView`、`Grid`、`GridView`、`Tabs`、chips row 的 axis、wheel、virtualization、hit refresh、Ref/callback 矩阵。
- 已确认集合控件的滚动职责分层：`ScrollView` 负责 FluxUI wheel/default action/ScrollRef；`ListView/GridView` 负责 Gio list 虚拟化；`TabsScrollable` 继承横向 `ScrollView`；chip 本体不处理 wheel。
- 已确认横向滚动容器通过 `pointer.Filter` 限制 `ScrollY=0`，纵向 wheel 不会误触发横向 tabs/chips 滚动。
- 已确认 `ScrollView` 的 wheel target 区域来自本帧布局尺寸，child visual offset 和 hit ops 在下一帧重建，不会无条件污染整窗或兄弟控件命中。
- 已标出 `Grid` 非虚拟化、chips row 组合责任、`ListView/GridView` 无 `ScrollRef`、shift-wheel/剩余 delta 未建模等后续风险。

## 后续依赖

- A6.2 / A6.3 / A6.4：嵌套滚动、横向 delta 策略和滚动后命中刷新仍是滚动集合控件的核心回归依据。
- A11.3：docs browser 左右栏、代码块、表格、搜索 chips 需要继续验证父子滚动互不污染。
- A12.3：组件回归矩阵应覆盖 `ScrollView` 横纵轴、`ListView/GridView` 虚拟化、`TabsScrollable`、chips row 横向滚动。
- 后续修复若为 `ListView/GridView` 增加 `ScrollRef` 或统一 offset API，必须同时更新 A6.1、A9.1 和本文件中的职责边界。
