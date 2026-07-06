# A3.4 hit area 和 layout area 对齐审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 18:57:14 +08:00
- 负责人：Codex
- 关注：Layout、Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A3\\.4|hit area|layout area|PointerArea|Pressable|ClickArea|Container decoration|Tabs item" docs\\project-audit-roadmap.md docs\\project-audit-task-breakdown.md docs\\audits\\project-audit-baseline.md`
  - `rg -n "type .*PointerArea|PointerArea|ClickArea|Pressable|Container|Tabs|TabItem|Hit|Area|pointer.InputOp|event.Op" widget internal ui layout -g "*.go"`
  - `rg -n "LayoutClickArea|LayoutRippleArea|DrawRipple|DrawStateLayerCircle" internal\\render.go internal\\ripple.go`
  - `rg -n "Tabs\\(|Pressable\\(|ClickArea\\(|ContainerDecoration.*size|LayoutRippleArea|LayoutClickArea|hover target|blank area|hit" widget -g "*_test.go"`
  - `rg -n "func \\(w \\*pointerAreaWidget\\) Layout|func registerPointerArea|clip.Rect|func \\(c \\*clickAreaWidget\\) Layout|LayoutClickArea|func \\(c \\*decorationContainerWidget\\) layoutInteractive|layoutTransformedClickArea|func \\(t \\*tabsWidget\\) Layout|LayoutRippleArea|func \\(r \\*tabsRowWidget\\) Layout|tabsTabSurfaceWidget|TestBlankAreaMovesDoNotTriggerHoverTarget|TestContainerDecorationOnHoverIsChangeOnly|TestPointerAreaDispatchesPointerWheelAndSyntheticEvents" widget\\pointer_area.go widget\\click_area.go widget\\container.go widget\\utils.go widget\\tabs_dialog_toast.go internal\\render.go internal\\ripple.go widget\\interactive_layout_test.go widget\\pointer_area_test.go`
  - `go test ./internal ./ui ./widget`
  - `git diff --check`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/pointer_area.go`
  - `widget/click_area.go`
  - `widget/container.go`
  - `widget/utils.go`
  - `widget/tabs_dialog_toast.go`
  - `internal/render.go`
  - `internal/ripple.go`
  - `widget/interactive_layout_test.go`
  - `widget/pointer_area_test.go`
- 关联能力：
  - PointerArea 子尺寸命中注册
  - Pressable/ClickArea 无样式点击区域
  - ContainerDecoration 视觉区域和 margin 命中边界
  - Tabs item 等分/滚动命中边界
  - Ripple 不改变 layout/hit size
  - 空白区域不触发 hover target

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 16 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A3.4 执行前工作区干净；本任务只新增 `audit-a3.4-hit-layout-area.md` 并更新索引，不修改 runtime/widget/layout 源码。 |

## 命中区域链路

```text
PointerArea(child)
  -> child.Layout(ctx.Child(0)) 得到 childDims.Size
  -> clip.Rect(Max: childDims.Size) + event.Op(tag)
  -> Flux pointer event target 使用当前 ctx.PathID()

Pressable/ClickArea(child)
  -> event.UseClickable(ctx)
  -> ctx.LayoutClickArea(clickable.Handle(), child)
  -> Gio Clickable.Layout 的尺寸等于 child 返回尺寸

ContainerDecoration(decoration, child, interaction opts)
  -> LayoutInset(margin) 包住视觉布局
  -> 视觉尺寸 dims 包含 margin
  -> layoutTransformedClickArea 用 outerSize - margin 得到 innerSize
  -> 只在 margin 内侧注册 LayoutClickArea；padding/背景表面属于命中区

Tabs item
  -> 每个 tab item 独立 event.UseClickable(tabCtx)
  -> LayoutRippleArea(clickable.Handle(), tabContent)
  -> item hit size 等于 tabsTabSurfaceWidget 的布局尺寸
  -> 非 scrollable/fullWidth 等分父宽；scrollable 按内容最小宽
```

## 视觉区域、布局区域、事件命中区域对照表

| 输入范围 | 视觉区域 | 布局区域 | 事件命中区域 | 证据 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `PointerArea` | 自身不绘制视觉，只包裹 child。 | 返回 `childDims.Size`，空 child 或 disabled 时不额外扩张。 | `registerPointerArea` 使用 `clip.Rect(image.Rectangle{Max: size})` 注册 `event.Op`，`size` 来自 child layout；`size.X <= 0 || size.Y <= 0` 时不注册。 | `widget/pointer_area.go:156`、`widget/pointer_area.go:201`、`widget/pointer_area.go:210` | 合格。默认不会把父行、窗口或未测量区域注册成命中区。 |
| `PointerArea` pass-through | 无视觉变化。 | 与 child 相同。 | 默认 `passThrough=true`，通过 Gio `pointer.PassOp` 允许事件继续传给 sibling；这影响阻挡语义，不扩大命中矩形。 | `widget/pointer_area.go:55`、`widget/pointer_area.go:201` | 合格。pass-through 不是命中范围扩张。 |
| `Pressable` | 无固定视觉样式，视觉完全来自 child。 | `LayoutClickArea` 返回 Gio `Clickable.Layout` 的 `dims.Size`；child 在回调内决定尺寸。 | Gio clickable 包住 child 回调返回的尺寸；Focus target 注册在同一 ctx path，用于键盘 activation，不产生整窗 pointer hit。 | `widget/click_area.go:56`、`widget/click_area.go:80`、`internal/render.go:875` | 合格。Pressable hit area 等于 child layout area。 |
| `ClickArea` | 旧 API 别名，无视觉。 | 直接调用 `Pressable(child, onClick, opts...)`。 | 与 Pressable 完全一致。 | `widget/click_area.go:21`、`widget/click_area.go:23` | 合格。兼容 API 未引入新的命中边界。 |
| `ContainerDecoration` passive | 背景、边框、阴影、图片、圆角、transform 由 surface 绘制。 | `LayoutInset(margin)` 后返回外层尺寸，因此 layout area 包含 margin。 | passive 模式没有交互回调，不注册点击/hover 命中。 | `widget/container.go:141` | 合格。纯视觉容器不会成为 pointer target。 |
| `ContainerDecoration` interactive | 视觉布局先录制为 macro，随后回放；背景和 padding 属于视觉 surface。 | `dims` 是包含 margin 的外层尺寸。 | `layoutTransformedClickArea` 先扣除 margin 得到 innerSize，再在 margin 内侧注册 `LayoutClickArea`；padding 和背景 surface 是命中区，margin 不是。 | `widget/container.go:156`、`widget/container.go:231`、`widget/utils.go:1342` | 合格。不会把外 margin 或整行注册为命中区。 |
| `ContainerDecoration` transform | 视觉通过 `layoutSurfaceWithTransform` 应用 transform。 | transform 不改变返回布局尺寸。 | `layoutTransformedClickArea` 在 margin 内侧对 clickable 应用同一 transform，再注册 click area；事件区跟随 transform 后的视觉区域。 | `widget/container.go:220`、`widget/utils.go:1356` | 合格但需继续回归。当前实现意图是视觉和 hit 同步变换，建议后续用旋转/缩放用例补专门测试。 |
| `Tabs item` | item surface 由 `tabsTabSurfaceWidget` 绘制背景、边框、padding 和 label/icon。 | `tabsTabSurfaceWidget` 以 min width/min height 起步，并 clamp 到当前 constraints；row 中每个 item 有独立 x offset。 | 每个 item 内部独立 `UseClickable`，通过 `LayoutRippleArea` 注册，hit size 等于 item surface size；不是整条 tabs row。 | `widget/tabs_dialog_toast.go:189`、`widget/tabs_dialog_toast.go:263`、`widget/tabs_dialog_toast.go:312`、`widget/tabs_dialog_toast.go:570` | 合格。单个 tab 命中区与单项布局格对齐。 |
| Tabs non-scrollable/fullWidth | 选中 indicator 横跨或居中于 item，底部分隔线覆盖整 row。 | `equalWidth := fullWidth || !scrollable`，每个 child 的 Min/Max.X 被设置为父宽等分值。 | 每个 tab 的 hit area 等于等分后的 item 宽度；视觉底线是整 row 绘制但不是单个 item hit。 | `widget/tabs_dialog_toast.go:295`、`widget/tabs_dialog_toast.go:312` | 合格。等分模式中“整格可点”是布局语义，不是错误整行命中。 |
| Tabs scrollable | row 可宽于 viewport，由外层 horizontal ScrollView 裁剪/滚动。 | item 宽度按内容最小宽；row 总宽为各 item 累加。 | hit area 随 item 在 scroll content 内移动；viewport 外由 ScrollView 裁剪，后续与 A3.2 的滚动裁剪语义关联。 | `widget/tabs_dialog_toast.go:301`、`widget/tabs_dialog_toast.go:520` | 合格但依赖滚动容器裁剪。后续组合验收需覆盖 scrollable Tabs 的 viewport 外点击。 |
| Ripple state layer | ripple 绘制在 click area 后或前，半径按 press/history 计算。 | `LayoutRippleArea` 明确返回 clickable layout 的 `dims.Size`，绘制不会改变 child measured size。 | click target 仍来自 `Clickable.Layout` 的 child size；ripple 绘制不扩大 hit area。 | `internal/ripple.go:108`、`internal/ripple.go:110` | 合格。ripple 不改变 layout/hit 尺寸，后续 A4.3 可继续审查视觉边界。 |

## 现有测试覆盖

| 能力 | 测试 | 覆盖结论 |
| --- | --- | --- |
| 空白区域不误命中 hover target | `TestBlankAreaMovesDoNotTriggerHoverTarget` | 已覆盖 Button 代表性 clickable 在窗口空白区域移动不会触发 hover target。 |
| ContainerDecoration hover 变化语义 | `TestContainerDecorationOnHoverIsChangeOnly` | 已覆盖同一 container 内移动只触发一次 hover change；间接证明 container 注册的是自身交互目标。 |
| PointerArea 原始 pointer/wheel/synthetic 事件 | `TestPointerAreaDispatchesPointerWheelAndSyntheticEvents` | 已覆盖 PointerArea 事件分发、坐标、capture、click/dblclick/aux/contextmenu/wheel。 |
| Pressable ref 兼容 | `refs_test.go` 中 pressable case | 覆盖 ref command 可调度，但未断言 hit size。 |
| Tabs ref 兼容 | `refs_test.go` 中 tabs case | 覆盖 ref command 可切换，但未断言单 tab hit rect。 |
| Tabs 基础布局 smoke | `interactive_layout_test.go` 中 tabs case | 覆盖 tabs 可在交互组件矩阵中布局，不直接验证命中矩形。 |

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| 缺少直接 hit rect 断言 | 中 | 现有测试更多覆盖事件结果和 hover target，不直接读取 Gio router 的 ActionAt/命中矩形。Pressable、ContainerDecoration margin、Tabs item 建议补坐标级回归测试。 |
| ContainerDecoration transform 命中需专门回归 | 中 | 实现中视觉 transform 和 click transform 使用同一矩阵，但复杂旋转/缩放下是否完全符合用户直觉需要截图或坐标测试。 |
| Tabs scrollable 依赖 ScrollView 裁剪 | 中 | A3.2 已记录 ScrollView viewport/content 边界；本任务仅从 Tabs 侧确认 item hit 不等于整 row。viewport 外点击是否完全不可达，应在组合场景验收中继续覆盖。 |
| Focus target 与 pointer hit 不是同一概念 | 低 | Pressable 会注册 focus target 用于键盘 activation。审查结论只针对 pointer hit area，后续 A5 事件顺序审查需要继续区分 keyboard focus scope。 |
| ContainerDecoration padding 属于命中区 | 低 | padding/背景 surface 被视为容器视觉区域的一部分，因此可点；如果业务只希望文本可点，应使用子组件级 Pressable，而不是容器 onClick。 |

## 事实结论

- `PointerArea` 的命中注册来自 child layout 后的 `childDims.Size`，并通过 `clip.Rect(Max: size)` 约束 Gio `event.Op`；没有证据显示它把父行或窗口注册为命中区。
- `Pressable` 与旧 `ClickArea` 都通过 `LayoutClickArea` 包住 child，命中尺寸等于 child 回调返回尺寸；旧 API 只是别名兼容，不改变边界。
- `ContainerDecoration` passive 模式不注册 pointer target；interactive 模式中 layout area 包含 margin，但 hit area 明确扣除 margin，padding/背景表面属于视觉和命中区域。
- `ContainerDecoration` transform 路径对视觉和 click area 使用同一 transform 矩阵，当前设计目标是变换后的视觉区域与命中区域对齐。
- `Tabs item` 每个 item 独立使用 clickable 和 ripple；非滚动模式下 item 命中区等于等分布局格，滚动模式下等于 content item surface，不是整条 tabs row 或窗口。
- `LayoutRippleArea` 只绘制 bounded ripple 并返回 child/clickable 的 measured size；ripple 不扩大 layout area，也不扩大 hit area。
- 当前未发现 A3.4 指定组件把整行或整窗错误注册成命中区域。

## 验收

- 已覆盖 `PointerArea`、`Pressable`、`ClickArea`、`ContainerDecoration`、`Tabs item` 的视觉区域、布局区域、事件命中区域对照。
- 已明确 `ContainerDecoration` 的 margin 不参与 hit，padding/背景 surface 参与 hit。
- 已明确 Tabs 等分模式中整格可点属于 item layout 语义，不是整行误注册。
- 已记录 PointerArea、Pressable/ClickArea、ContainerDecoration、Tabs 的关键实现证据。
- 已标出现有测试覆盖和缺口，后续修复不会把既有命中行为误判为审查引入。
- `go test ./internal ./ui ./widget` 通过。
- `git diff --check` 通过；命令提示 `docs/audits/project-audit-baseline.md` 下次 Git touch 会 LF -> CRLF，这是行尾属性提示，不是 diff check 错误。

## 后续依赖

- A4.3 ripple 和 state layer 审查需要复核 ripple/state layer 不改变 layout 和 hit area 的结论。
- A5.4 PointerArea 影响范围审查应补充坐标级测试：child 外、同 row 空白、窗口空白、pass-through sibling 命中。
- A10.4 滚动集合控件族审查需要验证 scrollable Tabs 在 viewport 外的 item 不响应点击。
- A11.1 docs browser 组合场景审查需要把 tabs、chips、代码块和表格的 scroll viewport 命中边界纳入手动验收。
- 后续若调整 ContainerDecoration transform 或 margin 逻辑，必须补 Pressable/ContainerDecoration/Tabs 的 hit area 回归测试，避免视觉区域、布局区域、事件区域再次偏移。
