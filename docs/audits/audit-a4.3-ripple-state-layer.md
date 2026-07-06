# A4.3 ripple 和 state layer 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 19:20:05 +08:00
- 负责人：Codex
- 关注：Style、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A4\\.3|ripple|state layer|StateLayer|stateLayer|Ripple" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "ripple|Ripple|state layer|StateLayer|stateLayer|DrawState|resolveDecorationState|DrawRipple" -S .`
  - `rg -n "LayoutRippleArea|LayoutRippleOverlayArea|DrawStateLayerCircle|StateLayer\\(|materialAnimatedStateLayerOpacity|materialStateLayerOpacity|layoutStateLayerTouchTarget|layoutRippleTouchTarget|button-state-layer|touch-state-layer|md3-state-layer" widget\\button.go widget\\tabs_dialog_toast.go widget\\selection.go widget\\checkbox.go widget\\switch.go widget\\material3_components.go widget\\navigation.go widget\\media_card.go widget\\slider.go`
  - `go test ./internal ./style ./widget ./ui`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/ripple.go`
  - `internal/render.go`
  - `style/material3.go`
  - `style/interaction.go`
  - `widget/utils.go`
  - `widget/button.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
  - `widget/selection.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/navigation.go`
  - `widget/media_card.go`
  - `widget/slider.go`
  - `internal/ripple_test.go`
  - `widget/utils_test.go`
  - `widget/interactive_layout_test.go`
  - `widget/material3_defaults_test.go`
- 关联能力：
  - ripple primitive 绘制边界
  - state layer 混色和圆形绘制边界
  - ripple/state layer 绘制顺序
  - hover/pressed 触发条件
  - disabled 下交互反馈禁用
  - touch target 与 layout size 区分

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 19 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A4.3 执行前工作区干净；本任务只新增 `audit-a4.3-ripple-state-layer.md` 并更新索引，不修改 runtime/widget/style 源码。 |

## 绘制入口表

| 入口 | 所属包 | 调用场景 | 绘制内容 | 是否参与 layout size | 是否扩大 hit area |
| --- | --- | --- | --- | --- | --- |
| `style.StateLayer(container, onColor, opacity)` | `style` | Button、Tabs、Select row、Navigation 等表面背景混色 | 返回混合后的 `color.NRGBA`，由组件作为背景色继续绘制。 | 否。只计算颜色。 | 否。只影响颜色。 |
| `Context.DrawRipple(clickable, size, spec)` | `internal` | Button 内部背景阶段、`LayoutRippleArea` / `LayoutRippleOverlayArea` 内部 | 在 `size` 的 rounded rect clip 内遍历 clickable press history，绘制扩散圆。 | 否。只记录 draw ops。 | 否。使用传入 size 裁剪。 |
| `Context.DrawStateLayerCircle(clickable, center, diameterDp, spec, stateOpacity)` | `internal` | Checkbox、Radio、Switch、Slider handle 等圆形控制面 | 在指定圆形 clip 内绘制 state layer，并可叠加以 center 为原点的 ripple。 | 否。只记录 draw ops。 | 否。绘制圆与事件目标分离。 |
| `Context.LayoutRippleArea(clickable, spec, child)` | `internal` | Tabs、Select trigger/item、Navigation、Dropdown trigger 等 | Gio clickable 注册包住 child；ripple 绘制在 child 之前。 | 返回 clickable/child 的 measured size。 | 否。hit area 等于该 helper 布局的 child size。 |
| `Context.LayoutRippleOverlayArea(clickable, spec, child)` | `internal` | `md3ActionSurface`、Card 等 opaque surface | Gio clickable 注册包住 child；ripple 绘制在 child 之后。 | 返回 clickable/child 的 measured size。 | 否。hit area 等于该 helper 布局的 child size。 |
| `layoutRippleTouchTarget` | `widget` | 需要最小 touch target 但 child 可小于 48dp 的组件 | 先测量 child，再用 offset 布置较大的 ripple/click target，最后返回 child size。 | 返回 child size。 | 是显式 touch target 设计，不是 ripple 自行扩大。 |
| `layoutStateLayerTouchTarget` | `widget` | Checkbox、Radio、Switch 等 | 先测量 child，再布置较大的 click target，最后绘制圆形 state layer/ripple。 | 返回 child size。 | 是显式 Material touch target 设计，不是 state layer 自行扩大。 |

## 绘制顺序

| 组件/路径 | 顺序 | 结论 |
| --- | --- | --- |
| `Button` / `internal.LayoutButton` | widget 先把 hover/pressed state layer 混入 `background`；internal 背景阶段填充 rounded rect；同一背景阶段在 rounded clip 内 `DrawRipple`；随后绘制 child；最后绘制 border 和 focus indicator。 | state layer 属于背景色，ripple 在背景之上、内容之下，focus 在最上层。 |
| `md3ActionSurface` | 先按 hover/pressed 把 state layer 混入 `bg`；通过 `LayoutRippleOverlayArea` 记录 surface child；停止宏后先绘制 shadow，再 replay surface；ripple overlay 在 surface 内容之后；focus indicator 最后绘制。 | 适合 Card/IconButton/FAB 等 opaque surface；ripple 不影响 surface 测量。 |
| `Tabs item` | 先计算 `stateOpacity` 并生成 `tabBg`；`layoutMaterialTabContent` 绘制 tab 内容背景；`LayoutRippleArea` 注册 item 命中并绘制 ripple；focus indicator 最后绘制。 | 每个 tab 独立 item hit，不注册整条 tabs row。 |
| `Select row` / `Dropdown trigger` | item/trigger 背景先混入 state layer；`LayoutRippleArea` 负责 bounded ripple 和点击区域；focus indicator 单独叠加。 | 行项目 ripple 有界到 item content，不改变 popup 或 trigger 尺寸。 |
| `Checkbox` / `RadioGroup` / `Switch` | child 本体先绘制；`layoutStateLayerTouchTarget` 追加 click target，再绘制圆形 state layer 和以圆心为原点的 ripple。 | 圆形 state layer 与 control center 对齐；layout size 仍取 child。 |
| `Slider` | 轨道/handle 计算固定 state layer 尺寸；hover/pressed/dragged 决定 opacity；`DrawStateLayerCircle(nil, ...)` 只绘制圆形 state layer，不接入 ripple history。 | slider state layer 是 handle 视觉反馈，不注册额外 clickable。 |

## 触发条件

| 反馈 | 触发来源 | disabled 行为 | redraw 行为 | 备注 |
| --- | --- | --- | --- | --- |
| hover state layer | `event.Clickable.Hovered()` 或 `Snapshot().Hovered` | `materialAnimatedStateLayerOpacity(..., disabled=true)` 返回 0；LowCPU 策略下 hover 返回 0。 | 通过 `md3AnimateStateLayerOpacity` 创建/推进 motion state，并在动画运行时请求 `animation.running`。 | Balanced 策略下 hover 直接 snap，不创建 motion state。 |
| pressed state layer | `event.Clickable.Pressed()` 或 `Snapshot().Pressed` | disabled 跳过。 | pressed 进入/退出动画通过 `md3AnimateStateLayerOpacity` 推进。 | pressed 在 LowCPU 下仍保留动画状态。 |
| bounded ripple | Gio `widget.Clickable.History()` press history | disabled 路径不调用 `LayoutRippleArea` 或不调用 `DrawRipple`。 | `calculateRippleFrame` 未结束时 `drawRipplePress` 请求 `animation.ripple`。 | 持续按住且扩散完成后停止 invalidation；release/cancel 进入 fade。 |
| circular state layer ripple | `DrawStateLayerCircle` 接收 clickable 时读取 history | disabled 路径不调用 helper 或 clickable 为 nil。 | 与 bounded ripple 相同，由 `drawRipplePress` 控制。 | Slider 传入 nil，只绘制 state layer，不绘制 ripple。 |
| focus indicator | Gio focus 或组件 focus 状态 | disabled 下 `md3FocusProgress` target 为 0。 | 单独 focus animation，不属于 A4.3 的 ripple/state layer。 | 作为绘制层级参照记录。 |

## 事实结论

- `style.StateLayer` 只是颜色混合函数，不读事件、不注册输入、不写 layout；state layer 的事实来源仍是 A4.2 记录的 hover/pressed/focused/disabled/selected/error 状态。
- `internal.DrawRipple` 和 `internal.DrawStateLayerCircle` 都是 draw primitive：它们依据传入尺寸或圆心/直径绘制，并用 rounded rect 或 ellipse clip 限制可见区域，不计算组件尺寸。
- `LayoutRippleArea` 与 `LayoutRippleOverlayArea` 都通过 Gio clickable 注册输入，并返回 child/clickable 的 measured size。差异只在绘制顺序：前者 ripple 在 child 之前，后者 ripple 在 child 之后。
- Button 的 state layer 是在 widget 层把 `background` 与 `foreground` 混色后传给 `internal.LayoutButton`；internal 再在按钮 rounded rect 背景阶段绘制 ripple，最后绘制 border/focus。
- `md3ActionSurface` 统一了 IconButton、FAB、Card 等 Material 表面的 state layer/ripple：state layer 先混入 background，ripple overlay 在 surface 内容后绘制，shadow 在 replay surface 前绘制。
- 选择控件使用 `layoutStateLayerTouchTarget` 保持 child layout size，同时显式布置 Material 最小 touch target。这个 helper 可让 hit target 大于视觉 child，但扩大来源是 touch-target 语义，不是 ripple 或 state layer 绘制本身。
- 既有测试覆盖 ripple frame active/start/held/release/cancel/invalid bounds，以及 state layer idle 不建 state、退出后释放、Balanced/LowCPU 策略；组件级绘制顺序主要依赖源码审阅和 material defaults 字符串测试。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| `layoutRippleTouchTarget` / `layoutStateLayerTouchTarget` 返回 child size 但注册更大 touch target | 中 | 这是 Material 最小触控目标设计，可能让事件命中区大于视觉区域。后续验收不能误判为 ripple 扩大 hit area，但也需要文档明确哪些组件采用该语义。 |
| ripple/state layer 绘制顺序缺少直接视觉单元测试 | 中 | 当前 primitive 生命周期有测试，组件层级多依赖源码和 visual regression。若后续调整 `LayoutRippleArea` 与 `LayoutRippleOverlayArea` 的调用位置，可能改变 opaque surface 上 ripple 可见性。 |
| `internal` 依赖 `style` 的 ripple timing | 中 | `internal/ripple.go` 直接使用 `style.InteractionRippleExpand/Fade`，延续 A1.1 的反向依赖风险。A4.3 未修复，只记录边界。 |
| `md3ActionSurface` 与 `Button` 是两套表面路径 | 低 | 二者绘制顺序相近但实现不同；后续新增组件若选错路径，可能造成 ripple 在内容下/上不符合预期。 |
| 圆形 state layer 与 touch target 中心需要调用方传对 | 低 | `layoutStateLayerTouchTarget` 依赖 `layerCenter` 回调；调用方若传错中心，layout/hit 不变但视觉反馈会偏移。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出 ripple/state layer 绘制点 | 通过 | 已列出 `style.StateLayer`、`DrawRipple`、`DrawStateLayerCircle`、`LayoutRippleArea`、`LayoutRippleOverlayArea`、touch target helpers。 |
| 输出绘制顺序 | 通过 | 已按 Button、md3ActionSurface、Tabs、Select、选择控件、Slider 记录顺序。 |
| 输出触发条件 | 通过 | 已区分 hover state layer、pressed state layer、bounded ripple、circular ripple、focus indicator。 |
| ripple 不改变 layout | 通过 | `DrawRipple`/`DrawStateLayerCircle` 只绘制；`LayoutRippleArea`/`LayoutRippleOverlayArea` 返回 child/clickable measured size。 |
| ripple 不自行扩大 hit area | 通过 | ripple primitive 不注册输入；命中区只来自 Gio clickable 布局或明确的 Material touch-target helper。 |
| disabled 不触发 ripple/state layer | 通过 | 组件 disabled 路径跳过 clickable/ripple，`materialAnimatedStateLayerOpacity(..., disabled=true)` 返回 0。 |
| 未修改运行时代码 | 通过 | 本任务只新增审计子文件并更新索引。 |

## 后续依赖

- A4.4 cursor 策略审查应沿用 A4.3 结论：cursor 命中来源应与 clickable/touch target 一致，不应由 ripple/state layer 绘制区域反推。
- A5 事件系统审查需要继续确认 Material touch target 与 pointer target registry 的关系，特别是视觉 child 小于 48dp 时的命中边界。
- A7 Widget 矩阵审查应逐个组件确认是否采用 `LayoutRippleArea`、`LayoutRippleOverlayArea`、`layoutStateLayerTouchTarget`，并补充组件级视觉/命中测试。
- 建议补充 focused state layer 策略测试：当前 state layer helper 主要消费 hover/pressed，focus 主要由 focus indicator 表达；若后续引入 focus state layer，需要统一权威来源和绘制顺序。
