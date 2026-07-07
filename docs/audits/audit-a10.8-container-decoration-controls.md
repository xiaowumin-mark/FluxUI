# A10.8 容器和装饰控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.8 容器和装饰控件族审查。

## 事实结论

审查范围覆盖 `Container`、`ContainerDecoration`、`Card`、`Stack`、`Row`、`Column`、`Padding`。布局实现主要分布在 `widget/container.go`、`widget/media_card.go`、`widget/flex.go`、`widget/stack_center.go`，底层约束和绘制入口在 `internal/render.go` 与 `internal/ripple.go`。

| 组件 | constraints | decoration | event area | state layer / ripple |
| --- | --- | --- | --- | --- |
| `Row` | `widget.Row` 转为 `LayoutFlex(Horizontal)`；普通子项是 rigid，`Flexed/Expanded` 子项按剩余主轴空间分配。 | 无内置 decoration。 | 不注册 pointer/key target。 | 无。 |
| `Column` | 与 Row 相同，轴为 `Vertical`。在纵向滚动宽松上限中可能测出很高内容。 | 无内置 decoration。 | 不注册 pointer/key target。 | 无。 |
| `Stack` | 当前 `widget.Stack` 把所有非 nil 子项都包装为 `layout.Stacked`，输出为 Gio Stack 尺寸。 | 无内置 decoration；常用于叠放背景、遮罩、overlay。 | Stack 自身不注册事件；事件区域来自各子项自身。 | 无。 |
| `Padding` | 直接调用 `LayoutInset`，子项收到扣除 inset 后的约束，输出为子项尺寸加 inset。 | 只改变布局 inset，不绘制背景。 | 不注册事件；若 child 注册事件，事件区由 child 决定。 | 无。 |
| legacy `Container` | 先 `LayoutInset(Margin)`，再 `LayoutSurface(Padding)`。输出包含 margin、padding 和 child。 | 只支持旧 `style.Style` 的 background、padding、margin、radius。 | 不注册事件。 | 无。 |
| `ContainerDecoration` passive | `LayoutInset(Margin)` + `LayoutSurface(Padding)`；transform 只影响绘制，不改变 layout size。 | 支持 bg、gradient、padding、margin、radius、corner、border、opacity、circle、shadow、image、transform。 | 无交互配置且无状态 decoration 时不注册事件。 | 无自动 state layer。 |
| `ContainerDecoration` interactive | 与 passive 使用同一 visual layout；先录制视觉 macro，再用同一 `dims` 注册点击区，最后回放视觉。 | disabled/pressed/hover 按 `Decoration.Merge` 合并，绘制前 `stripStateDecoration`。 | `layoutTransformedClickArea` 先扣除 margin 得到 inner size；padding 和背景 surface 属于命中区，margin 不属于命中区。 | 只使用用户显式提供的 Hover/Pressed/Disabled decoration；不自动添加 Material state layer。 |
| `Card` / `FilledCard` | 默认是 `ContainerDecoration(deco, child)`；padding 默认 12dp，radius 默认 theme medium。 | Filled 默认 `SurfaceContainerHighest`；Elevated 加 tonal elevation shadow；Outlined 加 1px outline variant；`CardDecoration` 可覆盖。 | 非点击 Card 不注册事件。可点击 Card 不包 `Button`，而是在 card surface 外层直接注册 card-bounded clickable。 | 可点击 Card 使用 `LayoutRippleOverlayArea`，ripple 绘制在 surface 上方，尺寸等于 card surface，不扩大 layout 或 hit area。 |

### event area 对照表

| 场景 | 布局区域 | 视觉区域 | 事件命中区域 | 结论 |
| --- | --- | --- | --- | --- |
| `ContainerDecoration` 无交互 | 包含 margin + surface + padding + child。 | surface 不含外 margin；shadow 可能画到 surface 外侧。 | 无。 | 纯装饰不会变成 pointer target。 |
| `ContainerDecoration` 有 `OnDecoClick` / hover / pressed | 布局仍包含 margin。 | surface + padding + child；transform 后视觉跟随矩阵。 | 扣除 margin 后的 inner size；transform 时 click area 使用同一矩阵。 | 容器装饰不会把外 margin 或整行错误注册成命中区域。 |
| `ContainerDecoration` padding | padding 扩大 surface 尺寸。 | padding 是背景表面的一部分。 | padding 属于可点击区域。 | 这是“容器表面可点”的语义；若只想文本可点，应把事件放到 child。 |
| `ContainerDecoration` shadow | shadow 不进入 constraints；由 surface spec 绘制。 | 可画到 surface 外围。 | 不扩大 click area。 | 阴影不会意外扩大交互区域。 |
| `ContainerDecoration` transform | transform 不改变返回尺寸。 | surface 按 laid-out size 做 affine transform。 | interactive 路径用同一 transform 注册 click area。 | 当前设计是视觉和命中同步变换，但复杂旋转仍建议坐标测试。 |
| 可点击 `Card` | card surface 的布局尺寸。 | card surface + ripple overlay。 | `LayoutRippleOverlayArea` 的 clickable size 等于 child 回调返回的 card size。 | ripple/state layer 不扩大交互区域。 |
| `Row` / `Column` / `Stack` / `Padding` | 各自布局结果。 | 无内置视觉，除非 child 自己绘制。 | 无自身事件注册。 | 容器布局 primitive 不会主动创建整行命中区。 |

### state layer 和 decoration 来源

`ContainerDecoration` 是底层装饰容器，不自动套 Material 3 state layer。它只在存在 `Hover`、`Pressed`、`Disabled` decoration 或交互 callback 时进入 interactive path；视觉状态来自用户提供的 state decoration。

`Card` 是较高层组件。非点击 Card 只有 surface decoration；可点击 Card 使用 `event.UseClickable` 和 `LayoutRippleOverlayArea` 形成按压反馈。该反馈是 bounded ripple overlay，不改变 `ContainerDecoration` 的 measured size，也不把命中区扩展到 card 外。

## 风险

- `widget.Stack` 注释仍写“第一个子项为 Expanded，其余为 Stacked”，但实现把所有子项都作为 `Stacked`；A3.1 已记录该注释/实现差异，A10.8 继续沿用事实。
- `ContainerDecoration` 的 Hover/Pressed/Disabled decoration 可以改变 padding/margin；这会导致交互态布局跳动。当前实现允许这种能力，但设计上应建议状态 decoration 避免改变尺寸字段。
- `ContainerDecoration` 没有默认 state layer；如果调用方绑定 `OnDecoClick` 但不提供 Hover/Pressed decoration，交互可以工作但没有额外视觉反馈。
- padding 被视为容器 surface 的一部分，因此 interactive `ContainerDecoration` 的 padding 区域可点击；这不是扩大错误，但文档应提醒“容器可点”和“文本可点”的语义差异。
- shadow 不扩大 hit area，视觉阴影区域点击不会命中容器；这符合边界规则，但对大阴影卡片可能需要设计层说明。
- `ContainerDecoration` transform 路径已有 layout-size 测试，但缺少旋转/缩放后的坐标级 hit rect 自动测试。
- `CardDecoration` 可覆盖 padding/radius/border/shadow；如果状态或自定义 decoration 改变 surface 尺寸，可点击 Card 的 ripple/hit 仍跟随最终 surface size，但可能出现布局跳动。

## 验收

- 已建立 `Container`、`Card`、`Stack`、`Row`、`Column`、`Padding` 的 constraints、decoration、event area、state layer 矩阵。
- 已确认 `Row`、`Column`、`Stack`、`Padding` 作为布局 primitive 不注册自身事件区域。
- 已确认 legacy `Container` 和 passive `ContainerDecoration` 只做 layout/draw，不注册 pointer target。
- 已确认 interactive `ContainerDecoration` 的命中区扣除 margin，包含 padding/background surface，不会把外 margin、整行或整窗注册成命中区。
- 已确认 `Card` 非点击时不注册事件；可点击时使用 card-bounded `LayoutRippleOverlayArea`，不会通过 Button wrapper 或额外 touch target 扩大交互区域。
- 已确认 ripple/state layer 绘制不改变 layout size；现有测试覆盖 Card 不使用 Button wrapper、ContainerDecoration hover change-only、transform 不改变 layout size。

## 后续依赖

- A3.1：继续沿用 Row/Column/Stack/Padding/Container 的约束矩阵，尤其是 Stack 注释/实现差异。
- A3.4：继续沿用 ContainerDecoration margin 不命中、padding 命中、ripple 不扩区的结论。
- A4.1 / A4.2 / A4.3：若调整 decoration state 合并、state layer 或 ripple 绘制顺序，需要同步更新本文件的 state layer 边界。
- A10.1：Card 的 clickable/ripple 行为应与基础交互控件族区分；Card 是容器表面点击，不应退回 Button wrapper。
- A12.3：建议补坐标级测试：ContainerDecoration margin 外点击不触发、shadow 区域不触发、transform 后视觉/命中一致、Card ripple/hit 不超出 surface。
