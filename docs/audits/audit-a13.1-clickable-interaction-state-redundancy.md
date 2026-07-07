# A13.1 clickable 和交互态冗余审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A13.1 clickable 和交互态冗余审查。

## 事实结论

A13.1 的目标是审查 Button、Pressable、ClickArea、Checkbox、Switch、Tabs、Select option 中 hover、pressed、focused、ripple/state layer 的重复实现，并标出可收敛到 helper 的部分与暂不收敛的理由。当前代码已经存在三层公共能力：`event.UseClickable` / `Clickable.Snapshot` 统一读取 hover/pressed/focused，`resolveDecorationState` / `withDefaultStates` 统一状态装饰合并，`layoutStateLayerTouchTarget` 统一部分小型选择控件的 touch target、state layer 和 hit area。但这些 helper 还没有覆盖所有 clickable 族控件。

### 现有公共入口

| helper / 入口 | 当前职责 | 已使用位置 | 结论 |
| --- | --- | --- | --- |
| `event.UseClickable` | 为当前 path 绑定稳定 `ClickableState`，并绑定 runtime/path | Button、Pressable、Checkbox、Switch、RadioGroup item、Tabs item、Select field、Select option | 已是 clickable 状态所有权入口，应保留 |
| `Clickable.Snapshot(ctx, includeFocus)` | 一次性读取 hovered/pressed/focused，并上报 interaction diagnostics | Button、Checkbox、Switch、Tabs item | 应优先使用，避免多次 `Hovered()` / `Pressed()` / `Focused()` 分散调用 |
| `ClickedEvent` + `dispatchClickDefault` | 把 Gio click 转 FluxUI cancelable click，再执行默认行为或旧回调 | Button、Checkbox、Switch、RadioGroup item、Select field/option | 可抽成“drain click default action” helper |
| `RegisterFocusTarget` + `FocusActivate` | 注册键盘 activation 默认行为 | Pressable、Checkbox、Switch、RadioGroup item、Select field/option | 可抽成 clickable focus helper，但 Tabs 当前未接入 |
| `resolveDecorationState` | disabled > pressed > hovered 的 decoration 合并 | Button、Checkbox、Switch、RadioGroup、Input 等 | 已可复用；暂不处理 focused 状态是现有语义边界 |
| `withDefaultStates` | 为选择类控件补默认 hover/pressed/disabled decoration | Checkbox、Switch、RadioGroup | 可继续收敛到 selection-control helper |
| `layoutStateLayerTouchTarget` | 在不改变 measured size 的前提下注册扩展 touch target、画圆形 state layer/ripple | Checkbox、Switch、RadioGroup | 适合 Checkbox/Switch/Radio 族；不直接适配 Tabs/Select option 的整行矩形背景 |
| `materialStateLayerOpacity` / `materialAnimatedStateLayerOpacity` | 根据 hover/pressed/disabled 计算 state layer opacity | Checkbox、Switch、RadioGroup、Tabs、Select option、Slider | 可统一调用口径，但不同组件动画命名空间仍需保留 |
| `md3InteractionTiming` | 根据 hover/pressed/focused/disabled 计算动画时长和 easing | Button、Checkbox、Switch、RadioGroup、Tabs、Select | 已是交互动画时序的公共入口 |

### 组件冗余矩阵

| 组件 | 当前状态读取 | 当前 click / focus | 当前视觉反馈 | 冗余判断 |
| --- | --- | --- | --- | --- |
| Button | `clickable.Snapshot(ctx, true)`，再 `HoverChangedWithSnapshot` | ref 命令和 click event 直接走 dispatcher；focus 注册在更早逻辑中 | 自行计算 state layer、颜色动画、focus opacity，再交给 `LayoutButton` | click draining、disabled/loading gate、hover callback 可收敛；视觉部分因 Button variant/shadow/border/loading 复杂，暂不整体收敛 |
| Pressable / ClickArea | `event.UseClickable`，不读取 hover/pressed/focused | 手写 `RegisterFocusTarget`、ref drain、`ClickedEvent` loop | `LayoutClickArea`，无视觉反馈 | 低层行为很薄，适合复用 click/focus/ref helper；无需接入 state layer |
| Checkbox | `clickable.Snapshot(ctx, true)` | 手写 focus activation、ref drain、click default toggle | `withDefaultStates` + `resolveDecorationState` + `layoutStateLayerTouchTarget` | 与 Switch/RadioGroup 高度重复，应收敛 selection boolean 控件 helper |
| Switch | `clickable.Snapshot(ctx, true)` | 手写 focus activation、ref drain、click default toggle | 与 Checkbox 同样的 state decoration + touch target，但有 thumb/track/pressed progress | 可共享交互 shell；thumb/track 视觉状态暂不收敛 |
| RadioGroup item | 多次调用 `Hovered()` / `Pressed()`，未统一成局部 snapshot | 手写 focus activation、click default select | 与 Checkbox 类似的 default states + `layoutStateLayerTouchTarget` | 应改为 snapshot 并复用 selection item helper；单选 no-op 语义保留在调用方 |
| Tabs item | `clickable.Snapshot(tabCtx, !disabled)` | 直接 `clickable.Clicked`，未注册 FluxUI focus target / default action | 手写 `materialAnimatedStateLayerOpacity`、`LayoutRippleArea`、focus indicator、indicator animation | state layer/focus/ripple 可部分收敛；indicator、scrollable/full-width 布局暂不收敛 |
| Select field | 多次调用 `Hovered()` / `Pressed()` / `Focused()` | 手写 focus activation、click default toggle open | 通过 `md3ActionSurface` 处理 surface feedback，另有 outline/label/arrow 动画 | 可改为 snapshot + click/focus helper；field 语义含 open/error/label，不宜并入 selection option helper |
| Select option | 多次调用 `Hovered()` / `Pressed()` / `Focused()` | 手写 focus activation、click default select+close | 手写 row bg state layer、`LayoutRippleArea`、focus indicator | 与 Tabs item 同类“整行 selectable action”，适合抽 rectangular option action helper |

### 应收敛到 helper 的部分

| 编号 | 候选 helper | 覆盖范围 | 收敛内容 | 理由 |
| --- | --- | --- | --- | --- |
| CR-01 | `drainClickableDefaultAction(ctx, clickable, disabled, action)` | Button、Pressable、Checkbox、Switch、RadioGroup item、Select field/option | `for ClickedEvent` loop、`dispatchClickDefault`、disabled gate | 代码形态重复，且关系到 default action 可取消性，集中后更不容易漏 gate |
| CR-02 | `registerClickableFocusAction(ctx, disabled, action)` | Pressable、Checkbox、Switch、RadioGroup item、Select field/option，后续 Tabs | `RegisterFocusTarget` + `FocusDisabled` / `FocusActivate` | 键盘 activation 逻辑重复，集中后可统一 Enter/Space 默认行为边界 |
| CR-03 | `drainBoolRefCommands(ctx, ref, disabled, current, onChange)` | Checkbox、Switch | bool set/toggle、disabled 时丢弃、真实变化才回调 | Checkbox 与 Switch 代码基本同构 |
| CR-04 | `selectionInteractionShell` | Checkbox、Switch、RadioGroup item | snapshot、default states、state layer touch target、label hover color 基础口径 | 三者都使用圆形 control-centered state layer，命中区和视觉区规则一致 |
| CR-05 | `selectableRowSurface` | Tabs item、Select option，后续 Menu item | row/card 矩形背景、hover/pressed state layer、ripple、focus indicator | Tabs item 与 Select option 都是整行/整块 action surface，当前手写逻辑重复 |
| CR-06 | `clickableInteractionSnapshot(ctx, clickable, disabled, includeFocus)` | RadioGroup、Select field、Select option | 避免同帧多次调用 `Hovered()` / `Pressed()` / `Focused()` | 减少重复 input section 计数和状态读取，诊断更稳定 |
| CR-07 | `hoverCallbackFromSnapshot` | Button，后续可用于 Pressable 扩展 | `HoverChangedWithSnapshot` + 旧 `OnHover` dispatch | Button 当前已经局部正确，抽 helper 可避免其他控件复制错误版本 |

### 暂不收敛的部分

| 项 | 暂不收敛理由 | 保留边界 |
| --- | --- | --- |
| Button 的 `LayoutButton` surface | Button variant 包含 filled/tonal/outlined/text/elevated、shadow、border、loading icon、focus opacity，和选择控件的圆形 state layer 不同 | 可抽 click/focus/drain，视觉 surface 暂保留 Button 专属 |
| Pressable / ClickArea 的无视觉语义 | 它们是低层 escape hatch，承诺“不附带视觉反馈” | 只收敛行为 helper，不添加 ripple/state layer |
| Switch thumb/track 动画 | Switch 有 checkedProgress、pressedProgress、thumb icon、track/thumb 双色动画 | 只共享外层交互 shell，内部 switch visual 保留 |
| Tabs indicator | indicator 依赖 active/previous index、scrollable/fullWidth 和行内测量 rect | 不并入 selectable row helper，只复用 row state layer/focus/ripple |
| Select field open/error/label/arrow 动画 | Select field 同时承担表单 field、popup trigger、error/supporting text 和 arrow progress | 只抽 snapshot/click/focus，不抽 field visual |
| Select option select+close 默认行为 | option action 既更新 value 又关闭 popup，并触发 `onOpen(false)` | helper 只承载 row surface；action 本身由 Select 传入 |
| `resolveDecorationState` 的 focused 合并 | 当前 helper 只处理 disabled/pressed/hovered，Input 等组件手写 focused/error 优先级 | focused 与 error/selected/open 交织较多，需 A4/A10 后统一设计 |

## 风险

- RadioGroup item、Select field、Select option 当前多次调用 `Hovered()` / `Pressed()` / `Focused()`，每次会触发底层 input section 统计和 snapshot 观察，可能让 perf diagnostics 的 `input-ops/frame` 偏高或状态观察顺序不够直观。
- Checkbox、Switch、RadioGroup item 的 focus activation、click default action、disabled gate 和 state layer touch target 高度重复，后续修复 default action 或命中区域时容易出现一处修、一处漏。
- Tabs item 当前用 `clickable.Clicked` 直接切换 active，没有像其他控件一样通过 `ClickedEvent` + `dispatchClickDefault` 走 FluxUI cancelable click/default action 链路；这既是冗余，也是行为边界差异。
- Select option 与 Tabs item 都手写 `LayoutRippleArea`、state layer、focus indicator，但 Select option 还要关闭 popup；若未来修 focus ring 或 ripple 顺序，需要跨文件同步。
- 直接抽大而全 helper 会压扁 Button、Select field、Tabs indicator、Switch thumb/track 的真实差异，反而增加参数复杂度；收敛应从 click/focus/drain/snapshot 这类窄 helper 开始。

## 验收

- 已覆盖 Button、Pressable、ClickArea、Checkbox、Switch、Tabs item、Select field 和 Select option 的 clickable / hover / pressed / focused / ripple 相关路径。
- 已确认当前已有公共入口：`event.UseClickable`、`Clickable.Snapshot`、`resolveDecorationState`、`withDefaultStates`、`layoutStateLayerTouchTarget`、`md3InteractionTiming`。
- 已输出重复逻辑清单，并标出 CR-01 到 CR-07 的 helper 收敛候选。
- 已明确 Button surface、Pressable 无视觉语义、Switch thumb/track、Tabs indicator、Select field、Select option action、focused decoration 优先级暂不强行收敛的理由。
- 本轮只记录审查结果，不修改 widget/event/runtime 行为。

## 后续依赖

- A5.3/A5.5：若实现 CR-01/CR-02，应先对齐旧事件桥接和 default action 可取消性，避免 helper 改变 `PreventDefault` 生效边界。
- A7.3：Tabs item 如接入 focus target/default action，需要同步键盘默认行为矩阵，明确 Enter/Space 是否可取消。
- A10.1/A10.2/A10.4：基础交互、选择控件、滚动集合控件族的矩阵应作为 helper 收敛后的回归入口。
- A12.1/A12.2：将多次 `Hovered()` / `Pressed()` / `Focused()` 收敛为 snapshot 后，应回归 `input-ops/frame` 和高频事件 benchmark。
- A13.2/A13.3：overlay、滚动和 hit-test 的后续冗余审查应沿用“窄 helper 优先、复杂视觉暂不压扁”的收敛口径。
