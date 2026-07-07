# A13.2 overlay 逻辑冗余审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A13.2 overlay 逻辑冗余审查。

## 事实结论

A13.2 的目标是审查 Dialog、Popup、Menu、Select、Tooltip、Toast 中 outside click、portal、animation、focus restore 的重复逻辑，并明确统一 overlay 基础设施的候选边界。当前代码已有一组分散但可复用的基础能力：`md3OverlayProgress` 统一进出场进度，`layoutMD3OverlayTransition` / `layoutMD3RevealTransition` 统一视觉过渡，`md3DismissOnOutsidePress` 统一部分 outside press 关闭判断，`event.RegisterPortal` / `event.RegisterBoundary` 提供事件路径改写，`anchoredOverlayWidget` 提供全约束锚定绘制，`modalPanelInputShield` 为 modal 面板吞掉内部输入。

### 现有公共入口

| helper / 入口 | 当前职责 | 已使用位置 | 结论 |
| --- | --- | --- | --- |
| `md3OverlayProgress` | 根据 open/visible 计算 overlay 进出场进度和是否继续渲染 | Dialog、Popup、Select popup、Menu submenu、Tooltip、Toast | 已是 overlay 动画状态公共入口，应保留并补足生命周期包装 |
| `layoutMD3OverlayTransition` | 对 child 做 opacity + Y offset 过渡 | Toast、Tooltip | 适合轻量 anchored overlay，不适合 Dialog/Popup 的 dialog-specific transition |
| `layoutMD3RevealTransition` | 对 popup/menu 做裁剪 reveal 过渡 | Select popup、Menu submenu | 适合菜单类下拉，不直接适合 modal panel |
| `md3DismissOnOutsidePress` | 用 protected rect 判定 outside press 并触发关闭 | Select、DropdownMenu | 可收敛 Select/DropdownMenu，Dialog/Popup mask click 不是同一模型 |
| `event.RegisterPortal` | 将 portal root 的逻辑父级改写为 owner path | Dialog、Popup、EventPortal | Dialog/Popup 已使用；Select/Menu/Tooltip/Toast 当前主要靠 `op.Defer` 或 anchored overlay，未统一注册 portal |
| `event.RegisterBoundary` | 在 event path 上注册 modal boundary | Dialog | Dialog 是 modal stop boundary；Popup 当前 portal-only，差异应保留或显式配置 |
| `anchoredOverlayWidget` | 用根约束把子项锚定到 N/S/Center 等位置，返回根尺寸 | Dialog、Popup、Toast | 适合全屏/根区域 overlay；不适合依赖触发器 rect 的 Select/Tooltip |
| `modalPanelInputShield` | 在面板尺寸上注册 pointer/scroll tag，避免内部点击穿透到遮罩 | Dialog、Popup | 应保留为 modal 面板基础设施，可与 modal overlay shell 组合 |

### 组件冗余矩阵

| 组件 | outside click / close | portal / mount | animation | focus restore | 冗余判断 |
| --- | --- | --- | --- | --- | --- |
| Dialog | 遮罩用 `fillWidget` + `Pressable`，仅当 `maskClosable && onOpenChange != nil` 时关闭；内部由 `modalPanelInputShield` 吞输入 | `anchoredOverlayWidget` + `RegisterPortal` + `RegisterBoundary(BoundaryStopPropagation)`；支持 global/local mount 配置但当前 Layout 路径主要按根约束使用 | 手写 open/ref 生命周期、`md3OverlayProgress("dialog-overlay")`、opened/closed 回调、`layoutMaterialDialogTransition`、section fade | 未看到打开时 focus trap、关闭时 restore、Escape 默认关闭实现 | 与 Popup 高度同构，适合抽 modal lifecycle/mask/panel shell；modal boundary 是 Dialog 专属或配置项 |
| Popup | 遮罩关闭逻辑与 Dialog 基本同构；内部同样使用 `modalPanelInputShield` | `anchoredOverlayWidget` + `RegisterPortal`，但没有 stop boundary；支持 global/local mount 配置 | 手写 open/ref 生命周期、`md3OverlayProgress("popup-overlay")`、opened/closed 回调、`layoutMaterialDialogTransition`、content fade | 未看到 focus restore / Escape 默认关闭实现 | 与 Dialog 重复最多；差异是 boundary、内容结构和尺寸策略 |
| Select | 使用 `md3DismissOnOutsidePress`，protected rect 包含 fieldRect + popupRect；选项点击会 select 并关闭 | 下拉通过 `op.Defer` + local offset 绘制，没有 `RegisterPortal` | `md3OverlayProgress("select-popup")` + `layoutMD3RevealTransition` + popup placement | field focused = opened 或 Gio focus；未看到关闭 restore 到触发器的命令 | 与 DropdownMenu outside press/placement 可收敛；与 Dialog/Popup modal mask 不应混用 |
| Menu / DropdownMenu | 普通 `Menu` 无独立 open/close；DropdownMenu 有 outside tag 和 protected rect；submenu 通过 hover/focus 激活，不是 outside click | Menu/submenu 通过 local `op.Defer` 绘制；无 portal | Menu submenu 使用 `md3OverlayProgress("menu-submenu")` + reveal transition；Menu body 本身无 overlay lifecycle | submenu 由 hover/focus 状态驱动；未看到 focus restore 规则 | Menu row surface 和 submenu reveal 可与 Select option/popup 收敛，但普通 Menu 不是 overlay owner |
| Tooltip | 依赖 child hover/focus 显示；没有 outside click | child-local `LayoutClickArea` + `op.Defer` offset 绘制；无 portal | `md3OverlayProgress("tooltip-popup")` + `layoutMD3OverlayTransition` | 只读 hover/focus，不转移 focus，也无需 restore | 适合轻量 hover overlay helper；不应接入 modal close/focus restore |
| Toast / Snackbar | duration 到期或 action 点击关闭；没有 outside click | `anchoredOverlayWidget` 锚定到顶部/中部/底部；无 portal/boundary | duration redraw + `md3OverlayProgress("toast-overlay")` + `layoutMD3OverlayTransition` | 不接收 focus restore；action 是普通按钮 | 适合 timed anchored overlay helper；不应与 menu/modal close 合并 |

### 应收敛到 helper 的部分

| 编号 | 候选 helper | 覆盖范围 | 收敛内容 | 理由 |
| --- | --- | --- | --- | --- |
| OR-01 | `modalOverlayLifecycle` | Dialog、Popup | ref drain、open state diff、`onOpen` / `onClose` / `onOpenChange`、`onOpened` / `onClosed`、quick duration、`md3OverlayProgress` | Dialog 与 Popup 生命周期代码同构，后续修 opened/closed 时容易漏一处 |
| OR-02 | `modalMaskLayer` | Dialog、Popup | mask color/alpha、全尺寸绘制、maskClosable gate、`onCancelOnly`、`onOpenChange(false)` | 两者遮罩点击关闭逻辑完全同形；集中后可统一“无 onOpenChange 不可由遮罩关闭”的受控边界 |
| OR-03 | `modalPortalPanel` | Dialog、Popup | owner path、portal scope、可选 stop boundary、`modalPanelInputShield`、center anchored overlay | portal/boundary/input shield 是 modal 核心边界，适合参数化 `boundaryMode` |
| OR-04 | `popupOutsideDismissRegion` | Select、DropdownMenu | outside tag、trigger rect、popup rect、dismiss callback、`md3DismissOnOutsidePress` | 下拉类 outside press 都是“保护触发器和弹层矩形”，不同于遮罩点击 |
| OR-05 | `anchoredTimedOverlay` | Toast、Snackbar | duration、redraw reason、progress、anchor、offset transition、action close | Toast/Snackbar 只是同一实现别名，未来若新增 timed banner 可复用 |
| OR-06 | `hoverOverlay` | Tooltip，后续可扩展 help popover | child hit area、hover/focus predicate、progress、deferred offset 绘制 | Tooltip 逻辑很薄，抽 helper 可以避免未来 hover overlay 重复写 record/defer/offset |
| OR-07 | `menuRevealOverlay` | Select popup、Menu submenu、DropdownMenu popup | placement、max height、reveal direction、deferred 绘制 | 下拉和 submenu 都使用 reveal transition，但 placement 输入不同，helper 应拆成“transition + placement adapter”两层 |

### 暂不收敛的部分

| 项 | 暂不收敛理由 | 保留边界 |
| --- | --- | --- |
| Dialog 内容结构 | Dialog 有 icon/headline/content/actions 分区、section fade、默认 confirm/cancel action | 只抽 lifecycle/mask/portal shell，不抽 Dialog body |
| Popup 用户自定义内容 | Popup 内容完全由用户定义，尺寸和 padding 更接近 generic modal panel | 可复用 modal shell，但 body 仍由 Popup 自己构造 |
| Dialog 与 Popup 的 boundary 差异 | Dialog 当前注册 stop boundary，Popup 只注册 portal；这决定 modal 事件冒泡边界 | helper 需要显式参数，不应默认把 Popup 改成 Dialog 语义 |
| Select field 视觉和 option 行为 | Select 同时承担表单 field、open state、option select+close、support/error text | 可抽 outside dismiss 和 reveal，不抽 field/option action |
| Menu submenu 激活模型 | submenu 由 hover/focus/pressed 驱动，不由 open prop 或 outside click 直接控制 | 可复用 reveal surface，不抽 open lifecycle |
| Tooltip hover 语义 | Tooltip 不应获得 focus、拦截外部点击或参与 modal boundary | 只抽 hover overlay，不接入 modal/timed helper |
| Toast duration 语义 | Toast 是时间驱动的 notification，不是用户触发器下拉或 modal | 可抽 timed overlay，不接入 outside press 或 portal boundary |
| focus restore / Escape | 当前审查未发现统一实现；直接抽 helper 会先发明行为 | 先作为缺口记录，待 A8.4/A7.3 规则落地后再实现 |

## 风险

- Dialog 与 Popup 的 lifecycle、mask、portal、transition 代码高度重复；后续修复 `onOpened/onClosed`、`maskClosable` 或 quick animation 时容易一处修复、一处遗漏。
- Dialog 注册 `RegisterBoundary(BoundaryStopPropagation)`，Popup 只注册 portal；如果未来用粗粒度 modal helper 直接合并，可能改变 Popup 冒泡到 owner 的规则。
- Select、DropdownMenu、Menu submenu、Tooltip 都使用 deferred/local overlay 绘制，但没有统一 portal 归属和 focus restore 语义；如果后续加入全局 overlay manager，需要先区分“事件路径需要改写”和“只需视觉延后绘制”。
- Select/DropdownMenu 的 outside press 依赖 protected rect 与当前帧位置；滚动、重排或 portal 化后若不复用统一 region 计算，容易重新引入误关闭或不关闭问题。
- focus restore 和 Escape 当前更多是缺口而不是重复实现；在缺少明确语义前收敛，会把未定义行为固化成公共 API。

## 验收

- 已覆盖 Dialog、Popup、Menu、Select、Tooltip、Toast/Snackbar 的 outside click、portal/mount、animation、focus restore 相关路径。
- 已确认当前公共基础设施：`md3OverlayProgress`、`layoutMD3OverlayTransition`、`layoutMD3RevealTransition`、`md3DismissOnOutsidePress`、`RegisterPortal`、`RegisterBoundary`、`anchoredOverlayWidget`、`modalPanelInputShield`。
- 已输出 OR-01 到 OR-07 的统一 overlay 基础设施候选边界。
- 已明确 modal、anchored timed、deferred menu、hover tooltip 等 overlay 类型不应被一个大而全 helper 压扁。
- 本轮只记录审查结果，不修改 widget/event/runtime 行为。

## 后续依赖

- A8.1/A8.2：实现 OR-03 前必须保持 global/local mount、portal owner、Dialog modal boundary 与 Popup portal-only 的既有差异。
- A8.3：实现 OR-02/OR-04 前应复验遮罩点击、内部点击、打开点击、Select/DropdownMenu protected rect 的关闭规则。
- A8.4/A7.3：focus restore 与 Escape 需要先补齐语义矩阵，再决定是否进入 overlay helper。
- A10.6：Overlay 控件族矩阵应作为 modal shell、timed overlay、hover overlay 收敛后的回归入口。
- A13.3：滚动、hit-test 和 deferred overlay 的命中刷新冗余审查，应复用本任务对 Select/DropdownMenu/Menu/Tooltip 的分类边界。
