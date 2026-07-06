# A3.1 基础布局约束矩阵

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 19:25:00 +08:00
- 负责人：Codex
- 关注：Layout
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "func (Row|Column|Stack|Center|Container|Padding|Fixed|Fill)|FixedWidth|FixedHeight|Fill|Sizing|Constraints|layout\\.|Container\\(" layout ui widget internal docs\\audits\\project-audit-baseline.md`
  - `rg -n "func \\(c \\*Context\\) Layout(Flex|Stack|Inset|Surface)|type (FlexChild|StackChild)|func layoutDecorationShell|func \\(e \\*expandWidthWidget\\)|func \\(e \\*expandHeightWidget\\)|func \\(f \\*fixedSizeWidget\\) Layout|func \\(c \\*containerWidget\\) Layout|func \\(p \\*paddingWidget\\) Layout|func clampPointToConstraints" internal\\render.go widget\\utils.go widget\\sizing.go widget\\container.go widget\\media_card.go`
  - `rg -n "FixedWidth|FillWidth|FillHeight|Padding\\(|Center\\(|Stack\\(|Row\\(|Column\\(|constraints|Constraints|Exact|Min|Max" layout widget ui -g "*_test.go"`
  - `go test ./layout ./internal ./ui ./widget`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `layout/flex.go`
  - `layout/stack.go`
  - `internal/render.go`
  - `widget/flex.go`
  - `widget/stack_center.go`
  - `widget/container.go`
  - `widget/sizing.go`
  - `widget/media_card.go`
  - `widget/utils.go`
  - `widget/list_grid.go`
  - `ui/extended_types.go`
- 关联能力：
  - Gio constraints 传递
  - Row/Column Flex 约束
  - Stack/Center 约束
  - Container/Padding inset 约束
  - Fixed/Fill sizing 约束
  - 有限约束和宽松测量边界

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 13 |
| `git status --short` | 无脏文件输出 |
| 判断 | A3.1 执行前工作区干净；本任务只新增 `audit-a3.1-basic-layout-constraints.md` 并更新索引，不修改 runtime/widget/layout 代码。 |

## 约束链路总览

```text
widget.Row / widget.Column
  -> internal.Context.LayoutFlex
  -> gioui.org/layout.Flex
  -> Rigid/Flexed 子项按 Gio Flex 规则分配主轴空间

widget.Stack
  -> layout.Stack
  -> internal.Context.LayoutStack
  -> gioui.org/layout.Stack

widget.Center
  -> gioui.org/layout.Center.Layout

widget.Container / widget.ContainerDecoration / widget.Padding
  -> internal.Context.LayoutInset
  -> gioui.org/layout.Inset
  -> internal.Context.LayoutSurface
  -> gioui.org/layout.Background

widget.FixedWidth / FixedHeight / FixedSize
  -> fixedSizeWidget.Layout
  -> 子项 exact 约束后再 clamp

widget.FillWidth / FillHeight / Fill
  -> expandWidthWidget / expandHeightWidget
  -> 把对应轴 Min/Max 改成父级 Max
```

本项目没有单独的“无限约束”类型。基础布局面对宽松测量时主要有两种表现：父级 `Max` 为 `0` 或无有效可用空间；滚动容器在主轴上把 `Max` 临时设置为 `1_000_000` 作为内容测量上限。A3.1 矩阵按“有限 Max”和“无有效 Max / 宽松测量上限”记录行为，后者会在 A3.2 继续细化。

## Constraints 输入/输出矩阵

| 组件 | 子项输入约束 | 输出尺寸规则 | 有限 Max 行为 | 无有效 Max / 宽松测量行为 | 证据 |
| --- | --- | --- | --- | --- | --- |
| `Row` | `widget.Row` 把子项转成 `internal.FlexChild`，由 `LayoutFlex(Horizontal)` 传给 Gio Flex；普通子项为 `Rigid`，`Flexed/Expanded` 子项为 `Flexed`。 | 输出为 Gio Flex `Dimensions.Size`。 | 主轴先布局 rigid，再把剩余主轴空间分给 flexed；交叉轴按子项最大尺寸和父约束决定。 | 若父级没有有效主轴空间，rigid/flexed 均受 Gio 当前 Max 限制；在滚动横轴的 `1_000_000` 上限下，Row 可测出很宽内容。 | `widget/flex.go:22-85`、`internal/render.go:1464-1489` |
| `Column` | 与 `Row` 相同，但轴为 `Vertical`。 | 输出为 Gio Flex `Dimensions.Size`。 | 主轴为高度；`Expanded`/`Flexed` 消耗剩余高度。 | 在纵向滚动容器中可能看到 `Max.Y=1_000_000`，因此 `FillHeight` 或 flexed 子项可能产生巨大测量尺寸。 | `widget/flex.go:27-85`、`widget/list_grid.go:192-201` |
| `Stack` | `widget.Stack` 当前把所有非 nil 子项都包装为 `layout.Stacked`；`layout.Stack` 再转为 Gio `Stacked`。 | 输出为 Gio Stack `Dimensions.Size`，通常为 stacked 子项最大尺寸再受父约束约束。 | 子项以父 Max 测量，但 Stacked 子项不会被强制填满父级。 | 无有效 Max 时无法自动扩展；宽松测量上限下会取最大子项尺寸。 | `widget/stack_center.go:36-59`、`layout/stack.go:20-37`、`internal/render.go:1492-1509` |
| `Center` | 使用 `gioLayout.Center.Layout`；Gio Direction 会清空子项最小约束，保留最大约束。 | 输出为 Gio Direction `Dimensions.Size`，子项绘制居中。 | 子项可按内容缩小，父级输出仍受原父约束约束。 | 无有效 Max 时只能在当前可用区域内居中；不会主动制造空间。 | `widget/stack_center.go:14-30` |
| `Container` legacy | 先对 margin 调用 `LayoutInset`，再对 surface padding 调用 `LayoutSurface`。 | 输出包含 margin、padding 和 child 尺寸；背景使用内容加 padding 后的尺寸。 | margin/padding 会缩减传给子项的 Max；返回尺寸再加回 inset。 | 若父级 Max 过小，Gio Inset 会把子项 Max 压到非负范围，容器不会突破父约束。 | `widget/container.go:79-127`、`internal/render.go:256-307` |
| `ContainerDecoration` | passive/interactive 都走 margin inset + surface；interactive 额外注册 click area，但尺寸使用同一 layout 结果。 | 输出为 visual layout 的 `dims`；transform 只影响绘制/命中区，不改变 layout size。 | padding/margin/背景尺寸与 legacy Container 一致；交互态 decoration 可改变 padding/margin 时会改变尺寸。 | 宽松测量下仍由 child + padding/margin 决定；transform 不参与尺寸扩张。 | `widget/container.go:129-250` |
| `Padding` | 只调用 `LayoutInset`，子项收到减去 inset 后的约束。 | 输出为子项尺寸加 inset 后的 Gio Inset 尺寸。 | 在有限 Max 下不会让子项突破父约束。 | 无有效 Max 时子项空间可能被压到 0；宽松测量上限下会按内容加 padding。 | `widget/container.go:105-111`、`widget/container.go:309-317` |
| `FixedWidth` | 把目标 dp 宽度 clamp 到父约束区间，然后设置 `Min.X=Max.X=width`。 | 输出宽度强制为 clamp 后 width，高度来自子项并受当前约束 clamp。 | 目标宽度大于父 Max 时降到父 Max，小于父 Min 时升到父 Min。 | 若父 `Max.X=0`，固定宽会被 clamp 到 0；它不把 0 Max 当作无限空间。 | `widget/sizing.go:8-31`、`widget/media_card.go:742-800` |
| `FixedHeight` | 与 `FixedWidth` 对称，设置 `Min.Y=Max.Y=height`。 | 输出高度强制为 clamp 后 height，宽度来自子项并受当前约束 clamp。 | 目标高度大于父 Max 时降到父 Max，小于父 Min 时升到父 Min。 | 若父 `Max.Y=0`，固定高会被 clamp 到 0。 | `widget/sizing.go:16-31`、`widget/media_card.go:742-800` |
| `FixedSize` | 同时固定宽高。 | 输出宽高均为 clamp 后尺寸。 | 不能突破父 Min/Max。 | 父 Max 为 0 的轴会塌缩到 0。 | `widget/sizing.go:24-31`、`widget/media_card.go:742-800` |
| `FillWidth` | `Max.X > 0` 时设置 `Min.X=Max.X=Max.X`；`Max.X <= 0` 时直接布局子项。 | 有效 Max 下输出宽度为父 Max；无有效 Max 下保持子项自然宽。 | 可预测地填满父可用宽度。 | `Max.X <= 0` 时是 no-op，不会塌缩。 | `widget/sizing.go:33-45`、`widget/utils.go:1415-1439` |
| `FillHeight` | 设置 `height=Max.Y`，小于 0 时归零，然后设置 `Min.Y=Max.Y=height`。 | 输出高度至少为 height，并 clamp 到 exact height。 | 可预测地填满父可用高度。 | `Max.Y=0` 时会把子项 exact 到 0；与 `FillWidth` 的 no-op 语义不对称。 | `widget/sizing.go:38-79` |
| `Fill` | `FillHeight(FillWidth(child))`。 | 宽高分别按 FillWidth/FillHeight 规则。 | 有效 Max 下填满父级宽高。 | 宽轴 `Max<=0` 时可能保持自然宽，高轴 `Max=0` 时塌缩到 0；在滚动主轴大上限下可能填到 `1_000_000`。 | `widget/sizing.go:43-45` |

## 现有测试覆盖

| 场景 | 测试 | 覆盖结论 |
| --- | --- | --- |
| 基础交互控件尺寸不额外膨胀 | `widget.TestDefaultInteractiveWidgetsDoNotAddDecorationPadding` | 间接覆盖部分容器/交互装饰不会默认增加尺寸。 |
| overlay/popup 受约束布局 | `widget.TestDropdownAndSelectPopupLayoutWithConstrainedSpace`、`TestDropdownAndSelectPopupLayoutInsideScrollView` | 间接覆盖 Column + ScrollView + popup 组合在有限视口下不会返回空尺寸。 |
| list/grid 横纵布局 | `widget/list_grid_test.go` 中多处 Row/Column/List/Grid 组合 | 覆盖滚动集合里的基础 Row/Column 使用，但不是 Row/Column 自身的约束矩阵测试。 |
| container transform | `widget/container_transform_test.go` | 覆盖 transform 不改变基础 layout 尺寸的组合场景。 |

当前缺少针对 `Row`、`Column`、`Stack`、`Center`、`Fixed*`、`Fill*` 在有限/零 Max/滚动大上限三类约束下的表驱动单元测试。A3.1 先记录事实矩阵，后续修复应补测试再改行为。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| `widget.Stack` 注释与实现不一致 | 中 | 注释称“第一个子项为 Expanded，其余为 Stacked”，但实现把所有子项都作为 `layout.Stacked`；`layout.StackChild.expanded` 也没有公开构造入口。后续依赖 Expanded 语义的 overlay/background 可能误判。 |
| `FillWidth` 与 `FillHeight` 在 `Max<=0` 时不对称 | 中 | `FillWidth` 在 `Max.X<=0` 时直接返回子项自然尺寸，`FillHeight` 在 `Max.Y=0` 时会 exact 到 0；组合 `Fill` 在零高度约束下可能塌缩。 |
| `Fill*` 在滚动主轴大上限下可能产生巨大内容尺寸 | 中 | `ScrollView` 主轴使用 `1_000_000` 测量上限；如果内容在主轴使用 `FillHeight`/`FillWidth`，可能填满测量上限而不是视口。A3.2 需要继续审查。 |
| `Center` 不更新 Flux `Context.Position` | 中 | `Center` 依赖 Gio Direction 的绘制变换；传给子项的 Flux context 没有显式 `WithPositionOffset`。若子树用 `ctx.Position()` 做 overlay/viewport 定位，可能拿到未居中前的位置。 |
| `ContainerDecoration` 交互态 decoration 可能改变布局尺寸 | 低 | hover/pressed/disabled decoration 会参与 visual layout；如果交互态改变 padding/margin，hover/press 可能触发布局跳动。 |
| 基础布局缺少直接约束矩阵测试 | 低 | 现有测试多为组合场景；后续修改基础布局时容易只靠视觉 smoke 发现回归。 |

## 事实结论

- `Row` 和 `Column` 是 Gio Flex 的 Flux 封装；普通子项为 rigid，`Flexed/Expanded` 子项按权重占用剩余主轴空间。
- `Stack` 的公开 widget 目前只使用 `Stacked` 子项，没有实际暴露 Expanded stack 子项语义；这与注释不一致。
- `Center` 直接复用 Gio `layout.Center`，子项最小约束会被清空，输出由 Gio Direction 约束回父级范围。
- `Container`、`ContainerDecoration` 和 `Padding` 的约束核心是 `LayoutInset`：先缩减子项 Max，再把 inset 加回输出尺寸；surface 背景尺寸来自内容加 padding 后的尺寸。
- `FixedWidth`、`FixedHeight`、`FixedSize` 会把目标 dp 尺寸 clamp 到父级 Min/Max 后传给子项，不能突破父约束。
- `FillWidth` 只在 `Max.X>0` 时强制填满；`FillHeight` 在 `Max.Y=0` 时会强制 0 高，二者在无有效 Max 下语义不完全一致。
- 滚动容器用 `1_000_000` 作为主轴内容测量上限，这不是基础布局本身的无限约束，但会放大 Fill/Flex 行为，需由 A3.2 继续审查。

## 验收

- 已列出 `Row`、`Column`、`Stack`、`Center`、`Container`、`Padding`、`FixedWidth/Height/Size`、`FillWidth/Height/Fill` 的 constraints 输入/输出矩阵。
- 已区分有限父级 Max、无有效 Max、滚动宽松测量上限下的尺寸行为。
- 已标出 Stack 注释/实现不一致、FillWidth/FillHeight 零 Max 不对称、Fill 与滚动大上限组合风险。
- 已记录现有测试覆盖和缺口，避免后续把既有语义变化误判为审查引入。
- `go test ./layout ./internal ./ui ./widget` 通过。
- `git diff --check` 通过。

## 后续依赖

- A3.2 滚动容器布局审查需要重点验证 `ScrollView` 的 `1_000_000` 主轴测量上限与 `Fill*`、`Flexed` 的组合行为。
- A3.3 横向内容溢出审查需要沿用本矩阵判断 Row、FixedWidth、FillWidth 在横向宽松约束下是否产生不可控宽度。
- A3.4 overlay/dialog/popup 布局审查需要复核 `Center` 未更新 Flux `Position` 对定位和事件命中区域的影响。
- A10.8 容器和装饰控件族审查需要继续验证交互态 decoration 是否允许改变 padding/margin 并造成 layout jitter。
- 若后续修改基础布局语义，应先补充表驱动约束测试覆盖有限 Max、零 Max、滚动大上限三类输入。
