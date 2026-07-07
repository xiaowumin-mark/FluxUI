# A10.6 Overlay 控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A10.6 Overlay 控件族审查。

## 事实结论

审查范围覆盖 `Dialog`、`Popup`、`Menu`、`DropdownMenu`、`Tooltip`、`Toast`、`Snackbar`。主实现分布在 `widget/tabs_dialog_toast.go` 和 `widget/material3_components.go`；`Snackbar` 是 `Toast` 的别名包装，`Menu` 本体是普通菜单面板，真正负责打开/关闭的弹出组合是 `DropdownMenu`，Select 下拉边界已在 A10.2/A8 系列中记录。

| 组件 | portal | modal | outside click / mask | animation | focus / Escape |
| --- | --- | --- | --- | --- | --- |
| `Dialog` | 在 centered overlay 内创建 `dialog-portal`，注册到 owner，并注册 `BoundaryStopPropagation`。 | 是事件层 modal：panel 内 FluxUI event path 在 dialog portal 截断；Gio pointer/scroll 由 `modalPanelInputShield` 防穿透。 | 默认 `DialogMaskClosable(true)`；只有 `open && maskClosable && onOpenChange != nil` 时遮罩可点击，触发 `onCancelOnly` 和 `onOpenChange(false)`。 | `md3OverlayProgress("dialog-overlay")`，默认 500ms enter / 150ms exit；`DialogQuick(true)` 设为 0。遮罩透明度和 panel transition 都跟随 progress。 | 无默认 auto-focus、focus trap、restore focus 或 Escape close；modal 差异只体现在 event boundary 和 pointer shield。 |
| `Popup` | 在 centered overlay 内创建 `popup-portal` 并注册到 owner。 | 非严格 modal：有遮罩和 panel shield，但没有 Dialog 的 stop boundary，panel 内 FluxUI event 可按 portal 规则冒泡到 owner。 | 与 Dialog 同构，默认 `PopupMaskClosable(true)`；遮罩关闭依赖 `onOpenChange`。 | `md3OverlayProgress("popup-overlay")`，默认 500ms enter / 150ms exit；`PopupQuick(true)` 跳过动画。 | 无默认 auto-focus、focus trap、restore focus 或 Escape close。 |
| `Menu` | 无 portal；作为普通 panel 按当前布局 path 渲染。子菜单用 row 内 `op.Defer` 延后绘制。 | 否。 | 本体无 open 状态，也不注册 outside click；`MenuStayOpenOnOutsideClick` 等配置主要供 `DropdownMenu` 包装场景使用。 | row/selection 用状态动画；子菜单用 `md3OverlayProgress("menu-submenu")` + reveal transition。 | menu row 注册 focus target 和 activation；`MenuDefaultFocus`、focusout、restore 配置入口存在，但未发现默认 auto-focus/restore/Escape 实现。 |
| `DropdownMenu` | 无 portal；popup 通过 trigger 本地 `op.Defer` 绘制，逻辑仍在当前子树。 | 否。 | trigger click / focus activation 切换 open；open 时 `md3DismissOnOutsidePress` 注册 viewport pass-through press target，保护 trigger rect 和 popup rect，保护区外且未 `PreventDefault` 才关闭。item select 默认关闭，`KeepOpen` 可保持。 | `md3OverlayProgress("dropdown-menu-popup")` + `layoutMD3RevealTransition`；placement 会按 viewport 上/下翻转并限制高度。 | trigger 和 row 是 focus target；无默认打开后 focus first item、focusout close、restore focus 或 Escape close。 |
| `Tooltip` | 无 portal；局部 `op.Defer` 绘制在 child 附近。 | 否。 | 无 outside click；显示条件为未 disabled、label 非空、child hover 或 focused。 | `md3OverlayProgress("tooltip-popup")` + `layoutMD3OverlayTransition`，只影响视觉透明度/偏移。 | child 区域使用 `LayoutClickArea` 获取 hover/focus snapshot；无独立 focus 管理或 Escape。 |
| `Toast` | 无 portal；`anchoredOverlayWidget` 按当前约束区域 top/center/bottom 锚定。 | 否。 | 无 outside click 或遮罩；按 duration 自动关闭，action click 会执行 `onAction` 并关闭。 | `md3OverlayProgress("toast-overlay")` + overlay transition；duration 未到期时请求 `animation.toast` redraw。 | 无 focus trap/restore/Escape；action 是普通按钮行为。 |
| `Snackbar` | 同 `Toast`。 | 否。 | `Snackbar(message, opts...)` 直接返回 `Toast(message, opts...)`；`SnackbarAction` 是 `ToastAction`。 | 同 `Toast`。 | 同 `Toast`。 |

统一边界可以归纳为：

1. 只有 `Dialog` 和 `Popup` 使用 `RegisterPortal`；只有 `Dialog` 额外注册 stop boundary。
2. 只有 `Dialog` / `Popup` 有遮罩；遮罩关闭都要求 `maskClosable` 和 `onOpenChange` 同时存在。
3. `DropdownMenu` 的外部点击不是遮罩，而是 viewport 级 pass-through press target + protected rect；普通 `Menu` 和 `Tooltip` 不具备 outside click 关闭。
4. 所有 overlay 动画都复用 `md3OverlayProgress`；Dialog/Popup/Toast 用 overlay offset/opacity，Menu/DropdownMenu 用 reveal，Tooltip 用局部 overlay transition。
5. 当前 overlay 族没有统一 overlay manager、z-index registry、focus trap、restore focus 栈或 Escape 默认关闭。z-order 依赖 `Stack` 顺序、调用方放置顺序和 `op.Defer` 顺序。

## 风险

- `DialogMount` / `PopupMount` 暴露 global/local API，但当前实现仍按当前父约束绘制；真正全局遮罩依赖调用方把 overlay 放在根 `Stack` 顶层。
- `Dialog` 与 `Popup` 都有遮罩和 panel shield，但只有 Dialog 截断 FluxUI event path；后续不要把 Popup 误写成严格 modal。
- `DialogOnCancelOnly` / `PopupOnCancelOnly` 注释包含 Escape，但当前代码只在遮罩点击路径调用；没有默认 Escape 关闭。
- `MenuDefaultFocus`、`MenuStayOpenOnFocusout`、`MenuSkipRestoreFocus` 等配置入口当前未形成完整 focus 管理行为，文档需避免承诺 Material/Web 级默认 focus restore。
- `DropdownMenu` outside click 依赖 `ctx.Position()`、trigger/popup rect 和 viewport；嵌套 scroll、transform、defer 叠层场景仍应保留回归入口。
- `Tooltip` 和 `Snackbar` 缺少直接测试覆盖；Toast 有基础布局覆盖，但没有 action/duration/focus 行为矩阵测试。
- 没有全局 overlay 栈时，多个 overlay 同时打开的关闭优先级、Escape 仲裁和 z-order 只能由调用方组合顺序间接决定。

## 验收

- 已建立 `Dialog`、`Popup`、`Menu`、`DropdownMenu`、`Tooltip`、`Toast`、`Snackbar` 的 portal、modal、outside click、animation、focus 矩阵。
- 已确认统一边界策略：portal 只属于 Dialog/Popup；modal stop boundary 只属于 Dialog；遮罩关闭只属于 Dialog/Popup；protected-rect outside press 属于 DropdownMenu/Select 类局部 popup；Tooltip/Toast/Snackbar 不做 outside click。
- 已确认 overlay 动画来源统一为 `md3OverlayProgress`，但 transition 形态按组件分为 centered modal、local reveal、tooltip offset 和 anchored toast。
- 已确认当前 overlay 族无默认 focus trap、restore focus 和 Escape close，相关行为需要 KeyboardScope 或后续组件级实现补齐。
- 已标出现有测试覆盖：Dialog/Popup 内外点击、popup placement、DropdownMenu/Select 约束布局；Tooltip/Snackbar 和 overlay focus/Escape 仍是后续测试缺口。

## 后续依赖

- A8.1 / A8.2 / A8.3 / A8.4：本文件复用 overlay 挂载、portal path、outside click、focus/Escape 的专项结论；后续修复需同步更新这些文件。
- A7.2 / A7.3：若新增 Escape close、Arrow navigation、focus trap，需要定义局部 KeyboardScope、可取消默认行为和 topmost overlay 仲裁。
- A9.1 / A9.2：DialogRef/PopupRef 打开关闭命令若引入 focus restore，需要记录命令生命周期与受控 open 的优先级。
- A12.3：回归矩阵应补 Tooltip hover/focus 显隐、Snackbar action/duration、DropdownMenu outside protected rect、Dialog modal Tab 越界、Escape 关闭顺序。
