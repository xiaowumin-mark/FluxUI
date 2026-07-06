# A3.2 滚动容器布局审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 18:39:11 +08:00
- 负责人：Codex
- 关注：Layout、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `gopls go_search ScrollView`
  - `gopls go_search ListView`
  - `gopls go_search Grid`
  - `gopls go_search ScrollRef`
  - `rg -n "type .*Scroll|ScrollView|ListView|Grid|ScrollRef|ScrollOnChange|Axis|offset|Offset|Position|Viewport|Content" widget layout internal ui event -g "*.go"`
  - `rg -n "func \\(s \\*scrollWidget\\) Layout|func \\(s \\*scrollWidget\\) layoutContent|scrollContentLimit|func \\(s \\*scrollWidget\\) processWheelEvents|func \\(s \\*scrollWidget\\) applyWheelDefault|func scrollView(SetPosition|ClampOffset|MaxOffset|ScrollByFraction)|func \\(l \\*listViewWidget\\) Layout|func \\(l \\*listViewWidget\\) layoutNonVirtualized|func \\(l \\*listViewWidget\\) wrapListBody|func \\(g \\*gridWidget\\) Layout|func \\(g \\*gridViewWidget\\) Layout|func resolveGridColumns|func buildGrid|func viewportForContext|func ScrollAttachRef|type ScrollRef|func \\(r \\*ScrollRef\\) Scroll" widget/list_grid.go widget/scroll_ref.go`
  - `rg -n "TestScrollViewWheelScrollsContent|TestHorizontalScrollViewRequiresHorizontalWheelDelta|TestHorizontalScrollViewScrollsOverClickableChild|TestVerticalWheelOverHorizontalScrollViewScrollsOuter|TestHorizontalScrollViewClickAfterScrollUsesUpdatedVisualPosition|TestNestedScrollPrefersInnerListViewUnderPointer|TestNestedScrollPrefersInnerScrollViewUnderPointer|TestListViewVirtualizationBuildsVisibleItemsOnly|TestGridViewVirtualizationBuildsVisibleCellsOnly|TestScrollRefQueuesCommands|TestListViewStateFollowsPositionOnInsertAndDelete|TestGridViewStateFollowsPositionOnInsertAndDelete|TestNonVirtualizedLargeListReportsWarning" widget/list_grid_test.go widget/list_grid_state_mismatch_test.go widget/refs_test.go`
  - `go test ./layout ./internal ./ui ./widget`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `docs/audits/audit-a3.1-basic-layout-constraints.md`
  - `widget/list_grid.go`
  - `widget/scroll_ref.go`
  - `widget/list_grid_test.go`
  - `widget/list_grid_state_mismatch_test.go`
  - `widget/refs_test.go`
  - `ui/extended_types.go`
- 关联能力：
  - ScrollView 主轴测量和裁剪
  - ListView/GridView Gio List 虚拟化
  - Grid 静态布局和响应式列数
  - ScrollRef 主轴命令
  - wheel 事件轴向过滤
  - content size、viewport size、offset clamp

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 14 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A3.2 执行前工作区干净；本任务只新增 `audit-a3.2-scroll-container-layout.md` 并更新索引，不修改 runtime/widget/layout 源码。 |

## 滚动容器规则总览

```text
ScrollView
  -> 单 child 自定义测量
  -> 主轴 Max 临时放大到 1_000_000
  -> contentMajor = childDims.Size 在主轴上的尺寸
  -> viewportMajor = clamp 后输出尺寸在主轴上的尺寸
  -> offset clamp 到 [0, contentMajor - viewportMajor]
  -> clip 到 viewport size 后按 -offset 重放 child ops

ListView
  -> virtualized=true: Gio layout.List
  -> virtualized=false: Row/Column 全量布局
  -> vertical 默认 expandWidth
  -> Horizontal 通过 ListAxis(Horizontal) 切换 Gio List Axis

Grid
  -> Grid: 静态 Row/Column 网格，不滚动、不虚拟化
  -> GridView: 按行虚拟化的垂直 Gio List
  -> GridMinItemWidth: 根据父 Max.X 和 padding/colGap 下调列数

ScrollRef
  -> 命令入队并触发 invalidator
  -> 下一帧 ScrollView drainCommands
  -> 命令全部解释为当前滚动主轴上的 offset
```

## 主轴、交叉轴、content、viewport、offset 矩阵

| 组件 | 主轴规则 | 交叉轴规则 | content size | viewport size | offset 规则 | 证据 |
| --- | --- | --- | --- | --- | --- | --- |
| `ScrollView` 默认纵向 | `resolveAxis(vertical=true, horizontal=false)` 得到 Gio `Vertical`；测量 child 时把 `Max.Y` 改为 `1_000_000`，`Min.Y=0`。 | 交叉轴 X 沿用父约束，不主动放大；child 宽度仍受父 `Max.X` 限制。 | `contentMajor = childDims.Size.Y`。content 宽度来自 child 测量结果。 | 输出 size 先把主轴裁到父 `Max.Y`，再整体 clamp 到父 Min/Max；viewport 矩形使用父 `Constraints.Max` 与 `ctx.Position()`。 | `state.list.Position.Offset` 作为像素 offset，负值归零，最终 clamp 到 `[0, contentMajor - viewportMajor]`。绘制时 clip 到输出 size，并以 `Y=-offset` 重放 child ops。 | `widget/list_grid.go:187-293`、`widget/list_grid.go:350-383` |
| `ScrollView` 横向 | 只有 `ScrollHorizontal(true)` 且 `ScrollVertical(false)` 时得到 Gio `Horizontal`；测量 child 时把 `Max.X` 改为 `1_000_000`，`Min.X=0`。 | 交叉轴 Y 沿用父约束，不主动放大；child 高度仍受父 `Max.Y` 限制。 | `contentMajor = childDims.Size.X`。 | 输出宽度裁到父 `Max.X` 并 clamp；高度来自 child 再受父 Min/Max 约束。 | offset 与纵向同一套函数，只是绘制偏移改为 `X=-offset`；`ScrollOnChange` 上报 x，y 为 0。 | `widget/list_grid.go:154-160`、`widget/list_grid.go:194-200`、`widget/list_grid.go:238-290` |
| `ScrollView` 双轴配置 | `resolveAxis` 不支持真正双轴滚动；当 horizontal 和 vertical 同时为 true 时仍返回 Vertical。 | 另一轴只是普通交叉轴约束。 | 只计算一个主轴的 contentMajor。 | 只形成一个主轴 offset 和一个 viewportMajor。 | 只能滚一个轴；不存在同时维护 x/y offset 的状态。 | `widget/list_grid.go:1155-1163` |
| `ListView` 虚拟化 | 默认 Vertical，可用 `ListAxis(Horizontal)` 切换；直接设置 `state.list.Axis` 后调用 Gio `layout.List.Layout`。 | 交叉轴由每个 item 的 Gio 布局结果和父约束决定；纵向列表外层会 `expandWidth`，横向列表不会默认 expandHeight。 | Gio List 的 `Position.Length` 是估算总长度；每个可见 item 由 builder 懒构建。 | `state.viewportMaj` 记录 `dims.Size` 的主轴尺寸；viewport 传给子项用于 overlay/诊断。 | offset 由 Gio List 内部维护，`OnReachEnd` 用 `BeforeEnd`、`First+Count` 和 viewport 比例兜底判断。 | `widget/list_grid.go:717-760`、`widget/list_grid.go:801-835` |
| `ListView` 非虚拟化 | 不使用 Gio List；纵向构造 `Column`，横向构造 `Row`。 | 与 Row/Column 规则一致；纵向仍经 `wrapListBody` expandWidth。 | content size 是全部 item 实际布局合成结果。 | 无独立滚动 viewport，只有普通父约束输出。 | 无内部 offset；不触发 `ListOnReachEnd` 的 Gio List 状态链路。 | `widget/list_grid.go:763-799` |
| `Grid` 静态网格 | 无滚动主轴；按列切行，行内 `Row`，行间 `Column`。 | 交叉轴就是普通 Row/Column 结果。 | content size 是所有子项全量布局结果。 | 无独立 viewport；只受父约束。 | 无 offset；大量子项只记录非虚拟化警告。 | `widget/list_grid.go:962-967`、`widget/list_grid.go:1111-1137` |
| `GridView` | 固定使用垂直 Gio List，按 rowCount 虚拟化；不支持横向滚动网格。 | 每行是 `Row`，列数由 base columns 或 `GridMinItemWidth` 解析；最后一行补 empty cell 保持列结构。 | content 主轴是 rowCount 行的总高度估算；单元格只构建可见行范围内的 index。 | `state.viewportMaj = dims.Size.Y`；viewport 传给行内 cell。 | offset 由 Gio List 维护，`GridOnReachEnd` 按行数判断尾部。 | `widget/list_grid.go:970-1042`、`widget/list_grid.go:1045-1065` |
| `ScrollRef` | `ScrollToStart/End/Offset/By` 均解释为当前 ScrollView 的主轴命令；Top/Bottom 只是垂直语义别名。 | 不包含交叉轴命令。 | 不直接读取 content size。 | 不直接读取 viewport size。 | `ScrollToOffset` 负值归零；命令下一帧消费，`ScrollBy` 按 content 长度比例换算像素后 clamp。 | `widget/scroll_ref.go:22-102`、`widget/list_grid.go:122-151`、`widget/list_grid.go:386-397` |

## 事件和命中边界

| 场景 | 现有规则 | 验证覆盖 |
| --- | --- | --- |
| ScrollView wheel 默认行为 | ScrollView 注册 pass-through wheel tag；纵向只接收 Y 范围，横向只接收 X 范围；wheel 事件先进入 Flux event dispatch，未取消时才执行默认滚动。 | `TestScrollViewWheelScrollsContent`、`TestHorizontalScrollViewRequiresHorizontalWheelDelta` |
| 嵌套滚动 | 轴向 filter 让横向内层不吃纵向 wheel；内层滚动区域在 pointer 下优先接收匹配轴事件。 | `TestNestedScrollPrefersInnerListViewUnderPointer`、`TestNestedScrollPrefersInnerScrollViewUnderPointer`、`TestVerticalWheelOverHorizontalScrollViewScrollsOuter` |
| clickable child | ScrollView wheel tag 采用 pass-through，不阻断 child click；横向滚动后命中使用更新后的视觉变换。 | `TestScrollViewWheelTargetDoesNotBlockChildClick`、`TestHorizontalScrollViewScrollsOverClickableChild`、`TestHorizontalScrollViewClickAfterScrollUsesUpdatedVisualPosition` |
| Scrollbar | scrollbar 使用 viewport 比例驱动 offset fraction；拖拽或轨道点击会请求 redraw。 | 主要由 `drawScrollBar`/`scrollViewScrollByFraction` 代码路径覆盖，缺少专门 scrollbar 拖拽单测。 |

## 现有测试覆盖

| 能力 | 测试 | 结论 |
| --- | --- | --- |
| ScrollView 纵向滚动和 onChange | `widget.TestScrollViewWheelScrollsContent` | 已覆盖纵向 wheel 触发 offset 和 onChange。 |
| ScrollView 横向轴过滤 | `widget.TestHorizontalScrollViewRequiresHorizontalWheelDelta` | 已覆盖横向容器忽略 Y wheel、响应 X wheel。 |
| ScrollView 子项命中 | `widget.TestScrollViewWheelTargetDoesNotBlockChildClick`、`widget.TestHorizontalScrollViewClickAfterScrollUsesUpdatedVisualPosition` | 已覆盖 pass-through 和滚动后的 click 位置。 |
| 嵌套滚动 | `widget.TestNestedScrollPrefersInnerListViewUnderPointer`、`widget.TestNestedScrollPrefersInnerScrollViewUnderPointer`、`widget.TestVerticalWheelOverHorizontalScrollViewScrollsOuter` | 已覆盖内外层按 pointer 和 axis 分配 wheel。 |
| ListView/GridView 虚拟化 | `widget.TestListViewVirtualizationBuildsVisibleItemsOnly`、`widget.TestGridViewVirtualizationBuildsVisibleCellsOnly` | 已覆盖只构建可见子集和 perf stats。 |
| 大型非虚拟化警告 | `widget.TestNonVirtualizedLargeListReportsWarning` | 已覆盖非虚拟化大型列表告警。 |
| ScrollRef 入队 | `widget.TestScrollRefQueuesCommands` | 已覆盖命令顺序，但未覆盖命令应用后的布局 offset。 |
| 列表/网格状态位置绑定 | `widget.TestListViewStateFollowsPositionOnInsertAndDelete`、`widget.TestGridViewStateFollowsPositionOnInsertAndDelete` | 明确当前状态按位置继承，这是既有语义和风险，不是 A3.2 引入。 |

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| ScrollView 主轴 `1_000_000` 测量上限放大 Fill/Flex 行为 | 中 | A3.1 已记录 `FillHeight`/`FillWidth` 在滚动主轴大 Max 下可能填满测量上限。A3.2 确认 ScrollView 会把单 child 主轴 Max 改为 `1_000_000`，因此 child 内部使用 Fill/Flex 时可能得到超大 content size。 |
| ScrollView 不是双轴滚动容器 | 中 | API 同时提供 `ScrollVertical` 和 `ScrollHorizontal`，但 axis 解析只有一个主轴；horizontal 和 vertical 同时为 true 时落回 Vertical。调用方若期待双轴滚动会误判。 |
| `ScrollOnChange` offset 不是像素 | 中 | ScrollView 内部 offset 是像素，但 onChange 上报 `first + off/1024`。因为 ScrollView 始终 `First=0`，外部收到的是缩放后的近似值，不是原始像素。 |
| ScrollRef 缺少横纵语义区分 | 中 | `ScrollToTop/Bottom` 是 `Start/End` 别名，绑定横向 ScrollView 时同样生效为左右滚动；这是正式 API 兼容行为，但文档和后续测试需明确。 |
| Grid 静态布局不滚动 | 低 | A3.2 输入包含 Grid，但当前 `Grid` 只是静态 Row/Column 网格；真正滚动和虚拟化在 `GridView`。后续审查不要把 `Grid` 当成滚动容器。 |
| GridView 仅垂直虚拟化 | 中 | GridView 没有横向轴选项，横向溢出取决于 cell/Row 宽度和父约束，需在 A3.3 继续审查。 |
| Scrollbar 拖拽缺少专门测试 | 低 | 代码有 scrollbar fraction 到 offset 的路径，但现有测试主要覆盖 wheel 和 click 命中，未直接覆盖 scrollbar drag/track。 |

## 事实结论

- `ScrollView` 是单 child 滚动容器，自行测量、裁剪并按 offset 重放 ops；纵向默认，横向必须同时设置 `ScrollVertical(false)` 与 `ScrollHorizontal(true)`。
- `ScrollView` 的滚动范围来自 child 在主轴上的实际布局尺寸；viewport 来自输出 size，offset 会被 clamp 到 `[0, contentMajor - viewportMajor]`。
- `ScrollView` 的交叉轴不会被放大到 `1_000_000`，因此横向滚动时高度仍受父约束，纵向滚动时宽度仍受父约束。
- `ListView` 默认虚拟化并委托 Gio `layout.List` 管理 offset、Length、First、Count；非虚拟化模式退化为 Row/Column 全量布局。
- `Grid` 是静态网格，不是滚动容器；`GridView` 是按行虚拟化的垂直滚动网格。
- `ScrollRef` 命令在下一帧由 `ScrollView` 消费，所有命令都只作用于当前主轴。
- 现有测试已覆盖 wheel 轴向、嵌套滚动、子项命中、虚拟化构建和位置状态风险；缺少 scrollbar 拖拽、ScrollRef 应用后 offset、ScrollView + Fill/Flex 大上限的直接矩阵测试。

## 验收

- 已列出 `ScrollView`、`ListView`、`Grid`、`GridView`、`ScrollRef` 的主轴和交叉轴规则。
- 已明确 content size、viewport size、offset clamp 的来源。
- 已明确横向滚动必须通过 `ScrollVertical(false)` + `ScrollHorizontal(true)` 形成单一横向主轴。
- 已明确纵向滚动默认只处理 Y wheel，横向滚动只处理 X wheel，双轴滚动当前不成立。
- 已标注 Grid 静态布局与 GridView 滚动虚拟化的边界，避免后续把两者混为同一个滚动模型。
- 已记录既有测试覆盖和缺口，避免把既有行为误判为本轮审查引入问题。

## 后续依赖

- A3.3 横向内容溢出审查需要继续验证横向 ScrollView、GridView 行内 Row、GridMinItemWidth 与 Fixed/Fill 宽度组合。
- A3.4 hit area 和 layout area 对齐审查需要复核 ScrollView 裁剪、`WithPositionOffset(-offset)` 和 pointer pass-through 对命中区域的影响。
- A6.1 ScrollView 和 ListView offset 审查需要进一步拆解 `ScrollOnChange`、ScrollRef 命令应用和 Gio List Position 的语义差异。
- A10.4 滚动集合控件族审查需要扩展到 Tabs、chips row，并补齐 scrollbar drag/track 的交互测试。
- 若后续修复 ScrollView 的大上限或双轴语义，应先补充 ScrollView + Fill/Flex、ScrollRef 应用后 offset、scrollbar drag 的表驱动测试。
