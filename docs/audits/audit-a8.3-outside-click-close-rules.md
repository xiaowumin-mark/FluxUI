# A8.3 outside click 关闭规则审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `codegraph explore "A8.3 outside click close rules DialogMaskClosable PopupMaskClosable Select Menu md3DismissOnOutsidePress maskClosable protected rect internal click scroll close popup"`
  - `rg -n "A8\.3|outside click|DialogMaskClosable|PopupMaskClosable|Select|Menu|外部点击|遮罩点击" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "md3DismissOnOutsidePress|dismiss|outside|popupRect|triggerRect|fieldRect|protected|activeSubmenu|onOpenChange\(.*false|open = false|state\.open" widget/material3_components.go widget/selection.go`
  - `rg -n "DismissOnOutsidePress|outside|MaskClosable|maskClosable|DialogMask|PopupMask|Dropdown|Select|Menu.*outside|protected|stayOpenOnOutsideClick" -g "*test.go" .`
  - `go test ./event ./internal ./widget ./examples/component_lab ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
  - `widget/selection.go`
  - `widget/utils.go`
  - `widget/interactive_layout_test.go`

## 事实结论

1. A8.3 的目标是审查 `DialogMaskClosable`、`PopupMaskClosable`、Select、Menu 的 outside click 关闭规则，输出外部点击、内部点击、打开点击、遮罩点击规则表；验收要求内部按钮、空白、可滚动组件不会误关闭弹窗。证据：`docs/project-audit-task-breakdown.md:396`、`docs/project-audit-task-breakdown.md:400`、`docs/project-audit-task-breakdown.md:401`、`docs/project-audit-task-breakdown.md:402`。
2. Dialog 默认 `maskClosable: true`，`DialogMaskClosable(false)` 只修改配置。真正关闭入口在遮罩 `fillWidget`：只有 `open && maskClosable && onOpenChange != nil` 时遮罩可点击，点击后先调用 `onCancelOnly`，再调用 `onOpenChange(maskCtx, false)`。证据：`widget/tabs_dialog_toast.go:776`、`widget/tabs_dialog_toast.go:778`、`widget/tabs_dialog_toast.go:818`、`widget/tabs_dialog_toast.go:1038`、`widget/tabs_dialog_toast.go:1043`。
3. Popup 与 Dialog 同构：默认 `maskClosable: true`，`PopupMaskClosable(false)` 修改配置；遮罩 `fillWidget` 的关闭条件同样是 `open && maskClosable && onOpenChange != nil`，点击后调用 `onCancelOnly` 和 `onOpenChange(false)`。证据：`widget/tabs_dialog_toast.go:1713`、`widget/tabs_dialog_toast.go:1715`、`widget/tabs_dialog_toast.go:1749`、`widget/tabs_dialog_toast.go:1940`、`widget/tabs_dialog_toast.go:1945`。
4. Dialog/Popup 的内部区域不会依赖 protected rect，而是依赖 z-order 和 `modalPanelInputShield`：遮罩先于 content 绘制，panel 位于遮罩上方；panel 内部由 `modalPanelInputShield` 注册 Gio pointer tag 消费 press/release/move/drag/enter/leave/cancel/scroll，防止内部点击和内部滚动落到遮罩。证据：`widget/tabs_dialog_toast.go:1072`、`widget/tabs_dialog_toast.go:1088`、`widget/tabs_dialog_toast.go:1970`、`widget/tabs_dialog_toast.go:1971`、`widget/tabs_dialog_toast.go:1977`。
5. 现有回归测试覆盖 Dialog/Popup：在打开且 maskClosable 的 overlay 中，panel 中心 press/release 不增加 closeCalls；遮罩区域 press/release 会触发关闭。证据：`widget/interactive_layout_test.go:335`、`widget/interactive_layout_test.go:349`、`widget/interactive_layout_test.go:363`、`widget/interactive_layout_test.go:403`、`widget/interactive_layout_test.go:417`、`widget/interactive_layout_test.go:421`、`widget/interactive_layout_test.go:434`。
6. `md3DismissOnOutsidePress` 是 DropdownMenu/Select 复用的外部点击 helper：它在 viewport 上注册 pass-through pointer tag，只消费 `pointer.Press`，把事件位置转换成点，若点不在 protected rect 列表内，则派发一个 cancelable FluxUI `pointerdown`；只有该 pointerdown 未被 `PreventDefault` 时才调用 `onDismiss`。证据：`widget/utils.go:50`、`widget/utils.go:61`、`widget/utils.go:62`、`widget/utils.go:70`、`widget/utils.go:80`、`widget/utils.go:87`、`widget/utils.go:90`。
7. DropdownMenu 的打开点击入口是 trigger 自身的 clickable，点击后本地 `open = !open` 并调用 `onOpenChange(open)`。Popup 可见时，它构造 `triggerRect` 和 `popupRect` 作为 protected 区；当 `open && onOpenChange != nil && !stayOpenOnOutsideClick` 时才启用 outside dismiss。证据：`widget/material3_components.go:799`、`widget/material3_components.go:815`、`widget/material3_components.go:821`、`widget/material3_components.go:919`、`widget/material3_components.go:921`、`widget/material3_components.go:924`。
8. DropdownMenu 的内部选择默认关闭弹窗：`menuCfg.onSelect` 先调用原始 onSelect，若 `MenuItem.KeepOpen` 为 true 则保持打开，否则调用 `onOpenChange(false)`。因此内部 menu item click 是“选择后关闭”，不是 outside click 误关闭。证据：`widget/material3_components.go:885`、`widget/material3_components.go:890`、`widget/material3_components.go:893`、`widget/material3_components.go:924`。
9. 普通 `Menu` 只是一个菜单 panel，不持有 open 状态，也不注册 outside click；`MenuStayOpenOnOutsideClick` 只是写入 `menuConfig`，该配置在 `DropdownMenu` 的 outside dismiss 判断中生效。证据：`widget/material3_components.go:375`、`widget/material3_components.go:446`、`widget/material3_components.go:470`、`widget/material3_components.go:515`、`widget/material3_components.go:919`。
10. Select 的打开点击入口是 field 自身：`layoutSelectField` 中 clickable 点击会切换 `state.opened` 并调用 `onOpen(state.opened)`。Popup 可见时，Select 构造 `fieldRect` 和 `popupRect` 作为 protected 区；outside press 关闭时先检查 `state.opened`，再设置 false 并调用 `onOpen(false)`。证据：`widget/selection.go:750`、`widget/selection.go:752`、`widget/selection.go:765`、`widget/selection.go:771`、`widget/selection.go:727`、`widget/selection.go:729`、`widget/selection.go:732`。
11. Select 内部 option click 是“选择后关闭”：option 激活后调用 `onChange`，若仍打开则设置 `state.opened = false` 并调用 `onOpen(false)`。disabled Select 会在 layout 开始时强制关闭。证据：`widget/selection.go:488`、`widget/selection.go:549`、`widget/selection.go:551`、`widget/selection.go:556`、`widget/selection.go:570`、`widget/selection.go:576`。
12. Select/DropdownMenu 的可滚动菜单区域被包含在 `popupRect` 中，outside helper 只监听 press，不监听 wheel/scroll；因此在 popup 矩形内的滚动组件不会因 outside press helper 触发关闭。证据：`widget/material3_components.go:899`、`widget/material3_components.go:923`、`widget/selection.go:708`、`widget/selection.go:731`、`widget/utils.go:71`。

### outside click 规则表

| 组件 | 打开点击 | 内部点击 | 内部空白/滚动 | 外部点击 | 遮罩点击 |
| --- | --- | --- | --- | --- | --- |
| `Dialog` | 外部由调用方控制 `open`；Ref/按钮可打开 | panel 在 content 层，内部 press 被 input shield 保护，不触发 mask close | panel 尺寸内的 pointer/scroll 被 shield 消费 | 无独立 outside helper | `maskClosable && onOpenChange != nil` 时关闭 |
| `Popup` | 外部由调用方控制 `open`；Ref/按钮可打开 | 同 Dialog，panel 内点击不触发 mask close | 同 Dialog | 无独立 outside helper | `maskClosable && onOpenChange != nil` 时关闭 |
| `DropdownMenu` | trigger click 切换 open 并回调 `onOpenChange` | item click 默认选择后关闭；`KeepOpen` 可保持 | popupRect 保护整块 menu，包括滚动内容 | protected rect 外 press 且未被 `PreventDefault` 时关闭 | 无遮罩 |
| `Menu` | 自身无 open 状态 | item click 由 `onSelect` 处理 | 普通 panel 内部布局 | 自身不注册 outside click；被 DropdownMenu 包装时由 DropdownMenu 管 | 无遮罩 |
| `Select` | field click 切换 `state.opened` 并回调 `onOpen` | option click 选择后关闭；disabled option 不激活 | popupRect 保护整块 option list，包括虚拟化/滚动内容 | fieldRect/popupRect 外 press 且未被 `PreventDefault` 时关闭 | 无遮罩 |

## 风险

1. Dialog/Popup 的遮罩关闭依赖 Gio z-order 和 `modalPanelInputShield`；如果后续调整 `Stack(mask, content)` 顺序或移除 shield，内部按钮/空白/滚动可能重新穿透到遮罩。
2. Select/DropdownMenu 的 protected rect 基于 `ctx.Position()`、trigger/field 尺寸和 popup offset 计算；复杂 transform、嵌套 scroll 或非标准 viewport 场景仍可能出现视觉位置与保护矩形不一致。
3. `md3DismissOnOutsidePress` 只监听 `pointer.Press`，不会处理 Escape、focusout 或窗口失焦；这些关闭策略应在 A8.4/A7 相关任务里单独建模。
4. 普通 `Menu` 暴露 `MenuStayOpenOnOutsideClick`，但自身没有 outside click 注册；该选项主要在 `DropdownMenu` 包装场景生效，单独使用 `Menu` 时不能期待自动关闭。
5. 现有测试直接覆盖 Dialog/Popup 内部 press 与遮罩 press；Select/DropdownMenu 的 outside press protected rect 缺少同等级交互回归测试，目前主要依赖源码审查和布局测试。

## 验收

- 已输出 Dialog、Popup、DropdownMenu、Menu、Select 的外部点击、内部点击、打开点击、遮罩点击规则表。
- 已确认 Dialog/Popup 内部 panel press 不会误触发 mask close，遮罩 press 可按 `maskClosable` 规则关闭。
- 已确认 DropdownMenu/Select 的打开触发区和 popup 区都进入 protected rect，内部点击/滚动不会被 outside helper 当作外部点击。
- 已区分“内部选择后按设计关闭”和“outside click 误关闭”：DropdownMenu item、Select option 的关闭是选择默认行为，不是 protected rect 失效。

## 后续依赖

- A8.4 focus trap 审查需要补充 Escape/focusout 与 outside click 的关系，尤其是 `MenuStayOpenOnFocusout` 和 Select 的 focus 行为。
- A10.6 Overlay 控件族审查应把 Dialog/Popup mask close、DropdownMenu/Select protected rect、普通 Menu 无 outside helper 的差异纳入矩阵。
- 后续测试补强可增加 Select/DropdownMenu 的 press 坐标回归：trigger/field、popup 内部、popup 外部、scroll 区域四类坐标。
