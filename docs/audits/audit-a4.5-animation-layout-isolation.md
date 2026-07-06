# A4.5 动画和布局隔离审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 20:30:00 +08:00
- 负责人：Codex
- 关注：Style、Layout
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A4\\.5|动画和布局|Dialog|Popup|Toast|Snackbar|Select menu|Tabs indicator" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits`
  - `rg -n "Dialog|Popup|Toast|Snackbar|Select|Tabs|indicator|Animate|animation|md3Overlay|Overlay|Portal|Position|Transform" widget ui internal style examples/component_lab -g "*.go"`
  - `rg -n "layoutMaterialDialogTransition|dialogSectionFade|indicatorProgress|tabs-indicator-transform|layoutMD3RevealTransition|md3OverlayProgress\\(" widget/tabs_dialog_toast.go widget/selection.go widget/utils.go`
  - `rg -n "Test.*Dialog|Test.*Popup|Test.*Toast|Test.*Snackbar|Test.*Select|Test.*Tabs|layoutMD3OverlayTransition|layoutMD3RevealTransition|animatedTabIndicatorBounds" widget ui internal examples/component_lab -g "*_test.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/tabs_dialog_toast.go`
  - `widget/selection.go`
  - `widget/utils.go`
  - `widget/interactive_layout_test.go`
  - `examples/component_lab/main.go`
- 关联能力：
  - overlay 动画进度来源识别
  - Dialog/Popup 普通布局与 overlay 布局隔离
  - Toast/Snackbar 锚定 overlay 动画边界
  - Select menu reveal 动画与 trigger 尺寸隔离
  - Tabs indicator 视觉动画边界
  - 动画 redraw 与布局尺寸关系识别

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 21 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A4.5 执行前工作区干净；本任务只新增审计子文件并更新索引，不修改 runtime/widget/style 源码。 |

## 动画影响范围表

| 组件/入口 | 动画来源 | 影响范围 | 布局尺寸是否随动画变化 | 结论 |
| --- | --- | --- | --- | --- |
| Dialog | `md3OverlayProgress("dialog-overlay")` 产生进度；`layoutMaterialDialogTransition` 用 progress 计算 `offsetY`。 | 遮罩透明度、modal panel 的 Y 方向视觉位移、dialog section 透明度。 | 否。panel 仍通过 `materialPanel` 按最终宽高约束测量；transition helper 只包裹 `op.Offset`，返回 child 实际尺寸。 | 动画只影响 overlay 视觉位置和透明度，不改普通布局尺寸。 |
| Popup | 与 Dialog 共用 `md3OverlayProgress` 和 `layoutMaterialDialogTransition`，namespace 为 `dialog-overlay`/`popup` 族状态。 | 遮罩透明度、popup panel 的视觉位移、content fade。 | 否。`layoutMaterialModalPanel` 先按 width/height/fullscreen 和 viewport margin 计算约束，transition 只做 offset。 | Popup 是明确 overlay 位置动画，不参与普通文档流布局。 |
| Toast | `md3OverlayProgress("toast-overlay")`，并在 duration 未结束时请求 `animation.toast` redraw。 | Toast 容器透明度、锚点方向的视觉偏移。 | 否。`anchoredOverlayWidget` 使用当前 Max 作为 overlay 区域，child transition 返回 toast 内容测量尺寸。 | Toast 是锚定 overlay 反馈，不应改变页面普通布局。 |
| Snackbar | `Snackbar` 是 `Toast` 的别名实现，`SnackbarAction` 转发到 `ToastAction`。 | 同 Toast。 | 否。 | Snackbar 与 Toast 共享动画和布局隔离规则。 |
| Select menu | `md3OverlayProgress("select-popup")` + `layoutMD3RevealTransition`。 | 下拉菜单的 clip reveal、透明度、基于 anchor/viewport 的 popup offset。 | 对 Select 自身返回尺寸无影响。`Select.Layout` 最终返回 `toggleDims`；popup 通过 recorded op + `op.Defer` 绘制。 | menu 动画只影响 deferred popup 绘制，trigger 普通布局尺寸稳定。 |
| Tabs indicator | `md3FloatStateFor("tabs-indicator-transform")`，`animatedTabIndicatorBounds` 插值 active indicator 左右边界。 | indicator 绘制矩形的 X/Width。 | 否。`tabsRowWidget.Layout` 先测量每个 tab 并计算 row Size，再绘制 indicator；返回尺寸是 tab row 的最终尺寸。 | indicator 是纯视觉绘制动画，不改变 tabs row 布局。 |

## 关键实现规则

| 规则 | 证据 | 审查结论 |
| --- | --- | --- |
| overlay progress 只产生动画进度和 shouldRender | `md3OverlayProgress` 根据 visible target 推进 float state，并返回 `visible || progress > 0.001`。 | overlay 关闭动画期间仍可保留绘制，但 progress 本身不直接改父布局约束。 |
| overlay transition 使用最终测量尺寸 | `layoutMD3OverlayTransition`、`layoutMaterialDialogTransition` 都调用 child 并返回 child size；动画通过 opacity 或 offset 包裹绘制。 | 过渡不会把动画中的偏移当作新的布局尺寸。 |
| reveal transition clip 绘制而不缩短布局结果 | `layoutMD3RevealTransition` 先 record child 得到完整 size，再按 progress clip replay，最后返回完整 size。 | Select menu reveal 不会因为 progress 小而压缩 popup 的测量结果。 |
| Select menu 不占用 trigger 布局 | `selectWidget.Layout` popup 可见时记录 popup 绘制，并在末尾 `return toggleDims`。 | Select 下拉不会把菜单高度计入普通布局。 |
| deferred popup 在同一 frame 后绘制 | Select popup 使用 `op.Record`、offset、`op.Defer`。 | 这是明确 overlay 绘制路径，普通布局返回值仍由 trigger 决定。 |
| Tabs indicator 先布局后绘制 | `tabsRowWidget.Layout` 中先得到 `tabRects`、`indicatorRects` 和 `maxH`，随后 `drawIndicator`，最后返回 `{X: x, Y: maxH}`。 | indicator 动画无法改变 tab item 或 row 的 measured size。 |

## 事实结论

- A4.5 输入列出的 `Dialog`、`Popup`、`Toast`、`Snackbar`、`Select menu`、`Tabs indicator` 都存在动画路径，但动画主要通过 Gio ops 的 `op.Offset`、`paint.PushOpacity`、`clip.Rect` 或绘制矩形插值表达。
- Dialog 和 Popup 使用 overlay mount：遮罩和面板被放进 `Stack(mask, content)`，content 内部通过 `anchoredOverlayWidget` 以当前约束 Max 作为 overlay 区域；这属于明确 overlay 布局，不是普通组件流内布局。
- Dialog/Popup 的面板尺寸由 `materialPanel` 或 `layoutMaterialModalPanel` 按 width、height、fullscreen、viewport margin 计算；`layoutMaterialDialogTransition` 只做 Y 方向视觉偏移，返回 child 尺寸。
- Toast/Snackbar 使用 `anchoredOverlayWidget` 锚定到 top/center/bottom；进入和退出动画只改变透明度和少量 offset，消息为空或关闭动画结束时返回空尺寸。
- Select menu 先布局 trigger 得到 `toggleDims`，再按 anchor 计算 popup 位置和最大高度；popup 绘制通过 recorded op 延后，最终 `Select.Layout` 返回 trigger 尺寸。因此菜单出现、展开和关闭动画不会推开下面的普通布局。
- `layoutMD3RevealTransition` 返回完整 child size，只改变 replay 时的 clip 高度和 opacity；这避免 reveal 过程中 popup placement 因动画进度变化反复跳动。
- Tabs indicator 动画只插值 indicator 的绘制边界；tabs row 的宽高来自 tab item 测量、fullWidth/scrollable 策略和 maxH，不来自 indicator progress。
- 本次审查未发现动画进度直接写入普通布局 min/max constraints，或将视觉 transform 后的位置反向用于普通布局尺寸的路径。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| `layoutMD3RevealTransition` 会在测量阶段执行 child 绘制 op | 中 | 该 helper 用 `op.Record` 包住 child，再按 clip replay。它返回完整 size 是稳定的，但 child 如果在 Layout 中有副作用，测量和最终 replay 的边界需继续受控。当前 Select menu 的 child 是 menu panel/list，风险可接受。 |
| Dialog/Popup transition 缺少透明度统一封装 | 低 | Dialog/Popup panel 位移动画由 `layoutMaterialDialogTransition` 处理，section fade 单独处理；面板整体没有像 Toast 那样统一 opacity。该点是视觉一致性风险，不是布局隔离失败。 |
| `Popup` 与 `Dialog` namespace/状态语义需要后续 A8 深审 | 中 | 本次只确认动画与布局隔离；modal、portal、outside click、focus trap 和 owner event boundary 的完整语义属于 A8 范围。 |
| Select popup placement 依赖当前 `ctx.Position()` | 中 | 在复杂滚动或 portal 场景下，anchor position 的准确性会影响 overlay 位置；A3.2/A8 需要继续覆盖滚动容器和 overlay portal 的组合场景。 |
| Toast/Snackbar 自动化覆盖不足 | 低 | 已识别实现路径，但当前测试名搜索未发现直接针对 Toast/Snackbar 动画隔离的单测；后续若改动 toast transition，应补充 smoke。 |
| `gopls go_vulncheck ./...` 本轮超时 | 低 | vulncheck 在 120 秒超时，未得到新结果。本任务未改依赖，风险不影响 A4.5 事实结论。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出动画影响范围表 | 通过 | 已覆盖 Dialog、Popup、Toast、Snackbar、Select menu、Tabs indicator。 |
| 动画只影响视觉或明确 overlay 位置 | 通过 | offset、opacity、clip、indicator draw 均在绘制/overlay 层；普通布局返回尺寸来自 child final size 或 trigger size。 |
| 不意外改变普通布局 | 通过 | Select 返回 `toggleDims`；Tabs 返回 row measured size；Dialog/Popup/Toast 位于 overlay anchor/Stack 语义内。 |
| 标出 redraw 关系 | 通过 | Toast duration 明确请求 `animation.toast`；通用 md3 动画 state 会请求 `animation.running`，已在 A2.4 基线中记录。 |
| 标出剩余风险 | 通过 | 已记录 reveal measurement side effect、overlay 后续 A8、Toast/Snackbar 自动化覆盖不足等风险。 |
| 未修改运行时代码 | 通过 | 本任务只新增审计子文件并更新索引。 |

## 后续依赖

- A5 事件系统审查需要确认动画 offset/clip 后的 pointer target 与视觉位置是否一致，尤其是 Dialog/Popup/Select overlay。
- A8 Overlay、Portal、Modal 和 Popup 边界需要继续审查 Dialog、Popup、Menu、Select、Tooltip、Toast 的 mount、outside click、focus trap 和 owner event parent。
- A10 性能审查应继续跟踪 overlay 动画是否在 idle 后停止请求 redraw，避免动画 state 长期保持 running。
- 若后续修改 `layoutMD3RevealTransition` 或 Select popup placement，应补充针对 reveal 过程中 popup 尺寸稳定、位置不跳变的单测。
