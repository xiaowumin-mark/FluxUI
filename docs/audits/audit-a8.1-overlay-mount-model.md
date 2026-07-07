# A8.1 Overlay 挂载模型审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。
- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Runtime、Layout、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A8\.1|Overlay|Dialog|Popup|Menu|Select|Tooltip|Toast|Snackbar|z-order|mount|portal" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type .*Dialog|func .*Dialog|type .*Popup|func .*Popup|type .*Menu|func .*Menu|type .*Select|func .*Select|type .*Tooltip|func .*Tooltip|type .*Toast|func .*Toast|Overlay|Portal|globalOverlay|z|Stack|op\.Offset|clip\.Rect|event\.Op|pointer\.Filter" widget ui internal layout examples -g "*.go"`
  - `go test ./event ./internal ./widget ./examples/component_lab ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/events.go`
  - `internal/context.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
  - `widget/selection.go`
  - `widget/utils.go`
  - `ui/extended_types.go`
- 关联能力：
  - Dialog / Popup modal overlay
  - Menu / DropdownMenu / Select popup
  - Tooltip local overlay
  - Toast / Snackbar anchored overlay
  - event portal / boundary
  - outside press dismissal
  - z-order by Stack and op.Defer

## 事实结论

1. A8.1 的目标是审查 Dialog、Popup、Menu、Select dropdown、Tooltip、Toast 的挂载模型，输出 global/local mount、z-order、layout parent、logical owner 表；验收要求每种 overlay 的布局边界和事件边界明确。证据：`docs/project-audit-task-breakdown.md:380`、`docs/project-audit-task-breakdown.md:384`、`docs/project-audit-task-breakdown.md:385`。
2. runtime 支持事件 portal：`RegisterEventPortal` 会把 overlay/portal root 的逻辑事件父级改写为 owner；event path 在遇到 `EventBoundaryStop` 时会截断，在遇到 redirect boundary 时会改写父级。证据：`internal/events.go:364`、`internal/events.go:367`、`internal/events.go:1075`、`internal/events.go:1083`。
3. Dialog 默认 `DialogMountGlobal`，配置项 `DialogMountMode` / `DialogGlobalOverlay` 存在；注释要求放在根 `Stack` 顶层才能获得真正全局 modal 行为。当前 `Layout` 未读取 `d.config.mount` 做分支，因此 global/local 语义主要由调用方放置位置和父约束决定。证据：`widget/tabs_dialog_toast.go:725`、`widget/tabs_dialog_toast.go:781`、`widget/tabs_dialog_toast.go:838`、`widget/tabs_dialog_toast.go:967`。
4. Dialog 打开时用 `md3OverlayProgress` 维护显隐动画；遮罩和内容通过 `Stack(mask, content)` 绘制，content 后绘制，因而位于 mask 上方。遮罩是 `fillWidget`，尺寸取当前约束 `Max`，可选点击关闭。证据：`widget/tabs_dialog_toast.go:1009`、`widget/tabs_dialog_toast.go:1043`、`widget/tabs_dialog_toast.go:1111`、`widget/tabs_dialog_toast.go:1599`。
5. Dialog content 使用 `anchoredOverlayWidget(..., gioLayout.Center)`，该 helper 把子树约束设为当前 `Constraints.Max` 的 exact size 并返回整个 overlay 区域尺寸。Dialog 内部 `dialog-portal` 注册到 owner，并额外注册 `BoundaryStopPropagation`，所以面板内部事件不会继续冒泡到 owner 外侧。证据：`widget/tabs_dialog_toast.go:1100`、`widget/tabs_dialog_toast.go:1102`、`widget/tabs_dialog_toast.go:1103`、`widget/tabs_dialog_toast.go:1572`。
6. Popup 与 Dialog 共用 global/local 配置和中心锚定 overlay，遮罩与内容同样通过 `Stack(mask, content)` 分层；Popup panel 注册 `popup-portal` 到 owner，但没有像 Dialog 一样注册 stop boundary。证据：`widget/tabs_dialog_toast.go:1716`、`widget/tabs_dialog_toast.go:1786`、`widget/tabs_dialog_toast.go:1869`、`widget/tabs_dialog_toast.go:1973`、`widget/tabs_dialog_toast.go:1975`、`widget/tabs_dialog_toast.go:1982`。
7. `modalPanelInputShield` 会在面板实际尺寸区域注册 pointer tag 并消费 press/release/move/drag/enter/leave/cancel/scroll，从 Gio 命中层防止面板内部事件穿透到遮罩。证据：`widget/tabs_dialog_toast.go:1640`、`widget/tabs_dialog_toast.go:1651`、`widget/tabs_dialog_toast.go:1659`。
8. Toast 是非 modal overlay，`Snackbar` 是 `Toast` 的别名包装。Toast 通过 `anchoredOverlayWidget` 按 `ToastTop/Center/Bottom` 锚定，返回当前约束区域尺寸；它没有 portal、遮罩或 outside click，只有可选 action 点击和 duration/onClose。证据：`widget/tabs_dialog_toast.go:1371`、`widget/tabs_dialog_toast.go:1430`、`widget/tabs_dialog_toast.go:1528`、`widget/tabs_dialog_toast.go:1552`。
9. `Menu` 本身不是弹出层容器，只是菜单 panel。它按普通布局返回自身尺寸，子菜单在菜单行内部用 `op.Offset` 和 `op.Defer` 延后绘制到行右侧。证据：`widget/material3_components.go:470`、`widget/material3_components.go:542`、`widget/material3_components.go:619`、`widget/material3_components.go:646`、`widget/material3_components.go:651`。
10. `DropdownMenu` 是 trigger + menu 的局部弹出组合。trigger 正常参与布局并决定返回尺寸；popup panel 先 record 测量，再根据 trigger 尺寸和 viewport 计算 placement，最后通过 `op.Offset` 加 `op.Defer` 绘制在 trigger 附近。证据：`widget/material3_components.go:799`、`widget/material3_components.go:899`、`widget/material3_components.go:903`、`widget/material3_components.go:912`、`widget/material3_components.go:928`、`widget/material3_components.go:935`。
11. `DropdownMenu` 的外部点击关闭由 `md3DismissOnOutsidePress` 负责。它注册一个覆盖 viewport 的 pass-through pointer tag，保护 trigger rect 和 popup rect；点击保护区外时派发 `pointerdown`，若未被 preventDefault 才调用关闭。证据：`widget/material3_components.go:919`、`widget/material3_components.go:924`、`widget/utils.go:50`、`widget/utils.go:71`、`widget/utils.go:81`。
12. Select dropdown 不是复用 `DropdownMenu`，而是在 `selectWidget.Layout` 内部实现相同形态：field 正常布局并返回 field 尺寸，dropdown panel 通过 record/measure、placement、`op.Offset`、`op.Defer` 绘制，outside press 保护 field rect 和 popup rect。证据：`widget/selection.go:485`、`widget/selection.go:527`、`widget/selection.go:708`、`widget/selection.go:721`、`widget/selection.go:732`、`widget/selection.go:745`。
13. Tooltip 是 child 的局部 hover/focus overlay。它用 `LayoutClickArea` 注册 child 命中区域，tooltip body 通过 `op.Offset` 和 `op.Defer` 绘制在 child 下方；返回尺寸仍是 child 尺寸，没有 portal、遮罩或 outside click。证据：`widget/material3_components.go:1748`、`widget/material3_components.go:1754`、`widget/material3_components.go:1759`、`widget/material3_components.go:1795`、`widget/material3_components.go:1802`。
14. 当前没有集中 overlay manager 或全局 z-index registry。z-order 主要来自调用方 Widget 顺序、`Stack` 子项顺序、以及 `op.Defer` 的执行顺序；多个 overlay 同时出现时，谁后布局/后 defer 通常谁在上层。证据：`widget/tabs_dialog_toast.go:1111`、`widget/tabs_dialog_toast.go:1982`、`widget/material3_components.go:935`、`widget/selection.go:745`。

### Overlay 挂载模型表

| 组件 | 挂载类型 | layout parent / 尺寸边界 | z-order 来源 | logical owner / 事件边界 | 外部点击 |
| --- | --- | --- | --- | --- | --- |
| `Dialog` | 声明为 global/local；当前实现按所在父约束 | `anchoredOverlayWidget` 覆盖当前约束区域，建议根 `Stack` 顶层 | `Stack(mask, content)`，content 在 mask 上 | `dialog-portal -> owner`，并 stop boundary | mask 可点击关闭，panel input shield 防穿透 |
| `Popup` | 声明为 global/local；当前实现按所在父约束 | 同 Dialog，中心锚定当前约束区域 | `Stack(mask, content)` | `popup-portal -> owner`，无 stop boundary | mask 可点击关闭，panel input shield 防穿透 |
| `Menu` | 普通 panel；子菜单局部 overlay | Menu 自身正常布局；submenu 挂在 row 旁 | row 内 `op.Defer` | 无 portal；沿布局 path | 无统一 outside click，仅子菜单 hover/focus |
| `DropdownMenu` | trigger 本地 anchored popup | trigger 返回尺寸；popup 受父约束和 placement 限制 | `op.Defer` after trigger | 无 portal；事件 path 仍是局部子树 | viewport pass-through tag，保护 trigger/popup rect |
| `Select dropdown` | field 本地 anchored popup | field 返回尺寸；popup 受父约束和 placement 限制 | `op.Defer` after field | 无 portal；option 是 select 子树 | viewport pass-through tag，保护 field/popup rect |
| `Tooltip` | child 本地 hover/focus popup | child 返回尺寸；tooltip 不扩展布局 | `op.Defer` after child | 无 portal；tooltip body 无独立外部边界 | 无 outside click |
| `Toast` / `Snackbar` | 非 modal anchored overlay | 覆盖当前约束区域，按 top/center/bottom 锚定 | 调用方顺序 + anchored overlay | 无 portal；action 是 toast 子树 | duration/action/onClose，无 outside click |

## 风险

1. `DialogMount` / `PopupMount` API 暴露 global/local 语义，但当前 Layout 未使用 `mount` 分支改变约束来源；如果调用方没有把 global overlay 放在根 `Stack` 顶层，遮罩只覆盖父布局分配区域。
2. Popup 注册 portal 但没有 stop boundary；后续 A8.2 需要确认这是预期的“能冒泡到 owner”行为，还是 modal popup 应该截断的缺口。
3. Menu、DropdownMenu、Select、Tooltip 都依赖 `op.Defer` 和局部 layout 顺序，没有集中 z-index；多个局部 popup 交叠时缺少可配置层级。
4. DropdownMenu 和 Select 的 outside click 判断依赖 `ctx.Position()` 与 protected rectangles；在复杂 transform、scroll 或 portal 场景下可能出现视觉位置和命中保护区不一致。
5. Toast/Snackbar 没有集中队列或 overlay manager，示例中通过调用方 `Stack` 和 key/serial 管理；多个 Toast 同时存在时层级和生命周期由调用方决定。

## 验收

- 已输出 Dialog、Popup、Menu、Select dropdown、Tooltip、Toast/Snackbar 的 global/local mount、z-order、layout parent、logical owner 对照表。
- 已明确 Dialog 与 Popup 的布局边界来自当前父约束，真正全局效果依赖根 `Stack` 顶层放置。
- 已明确 Dialog 使用 portal + stop boundary，Popup 使用 portal 但不 stop，局部 popup/tooltip/toast 不注册 portal。
- 已明确 DropdownMenu 和 Select 的 outside press 规则，以及 Tooltip/Toast 不具备 outside click 关闭机制。
- 已标出当前没有集中 overlay manager / z-index registry，后续任务不能假设已有统一层级调度。

## 后续依赖

- A8.2 portal 和 modal 边界审查应复用本文件对 Dialog stop boundary 与 Popup portal-only 的差异结论。
- A8.3 outside click 审查应重点覆盖 `md3DismissOnOutsidePress` 的 protected rect、pass-through pointer tag、scroll/transform 场景。
- A8.4 focus trap 审查应确认 Dialog/Popup/Select/Menu 是否需要统一焦点范围与 Escape 行为。
- A10.6 Overlay 控件族审查可基于本矩阵继续梳理 outside click、portal、animation、focus restore 的重复逻辑。
