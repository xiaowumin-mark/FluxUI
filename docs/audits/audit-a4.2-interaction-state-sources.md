# A4.2 交互态视觉来源审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 19:32:00 +08:00
- 负责人：Codex
- 关注：Style、Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A4\\.2|交互态视觉来源|Batch|Style" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `gopls go_search hover|pressed|focused|selected|error|disabled`
  - `rg -n "func \\(c \\*Clickable\\) (Hovered|Pressed|Focused|Snapshot|HoverChanged)|ObserveInteractionSnapshot|type Clickable struct|type InteractionSnapshot" event/pointer.go internal/interaction.go internal/clickable.go`
  - `rg -n "RegisterFocusTarget|Focused\\(|RequestFocus|focusTarget|FocusDisabled|FocusedTarget" event/keyboard.go internal/events.go`
  - `rg -n "resolveDecorationState|withDefaultStates|materialAnimatedStateLayerOpacity|materialStateLayerOpacity|md3DrawFocusIndicator|md3SelectionProgress|md3InteractionTiming|stripStateDecoration" widget/utils.go`
  - `rg -n "hovered|pressed|focused|disabled|selected|error|Focus|RegisterFocusTarget|resolveDecorationState|withDefaultStates|md3ActionSurface|InputError|SelectError|ChipSelected|IconButtonSelected|ListItemSelected|MenuSelectedKey" widget/button.go widget/input.go widget/selection.go widget/checkbox.go widget/switch.go widget/material3_components.go widget/tabs_dialog_toast.go`
  - `rg -n "Hover|Pressed|Focus|Disabled|Selected|Error|state layer|Decoration|Selection" -g "*_test.go" widget internal event style`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/pointer.go`
  - `event/keyboard.go`
  - `internal/clickable.go`
  - `internal/events.go`
  - `internal/interaction.go`
  - `widget/utils.go`
  - `widget/button.go`
  - `widget/input.go`
  - `widget/selection.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
- 关联能力：
  - hover/pressed/focused 输入状态来源
  - disabled/selected/error 组件语义状态来源
  - Decoration 状态解析边界
  - Material state layer 状态消费
  - focus indicator 状态消费
  - 状态来源重复计算风险识别

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 18 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A4.2 执行前工作区干净；本任务只新增 `audit-a4.2-interaction-state-sources.md` 并更新索引，不修改 runtime/event/widget/style 源码。 |

## 状态来源表

| 状态 | 权威来源 | 读取/设置入口 | 视觉消费入口 | 是否存在重复计算 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `hovered` | Gio `widget.Clickable.Hovered()` 经 `internal.ClickableState` 暴露，再由 `event.Clickable` 统一读取。 | `event.UseClickable` 绑定当前 `PathID`；`Clickable.Hovered()`、`Snapshot()` 读取。 | `resolveDecorationState`、`materialAnimatedStateLayerOpacity`、`md3ActionSurface`、Button/Tabs/Select item 等组件。 | 低。widget 不应自行判定 hover；多数通过 `event.Clickable`。 | 权威来源在 event/internal clickable，widget 只消费快照。 |
| `pressed` | Gio `widget.Clickable.Pressed()` 经 `internal.ClickableState` 暴露，再由 `event.Clickable` 统一读取。 | `Clickable.Pressed()`、`Snapshot()`、`ClickedEvent()`。 | state layer pressed opacity、ripple、Button 背景、选择控件 pressed progress。 | 低。Switch 有 `md3SwitchPressedProgress`，但它消费 pressed 值，不产生 pressed。 | 权威来源在 event/internal clickable，动画层只消费。 |
| `focused` | 两套焦点视角：普通 clickable focus 来自 Gio `ctx.Gtx.Focused(&clickable.button)`；组件树 focus target 来自 `Runtime.focusTarget`。 | clickable 用 `Clickable.Focused(ctx)`；键盘/默认动作用 `event.RegisterFocusTarget`、`RequestFocus`、`FocusedTarget`。 | `md3DrawFocusIndicator`、`md3FocusProgress`、Input/Select focus border、Tabs focus indicator。 | 中。Gio focus 与 Flux component-tree focus 并存，语义不同。 | clickable 视觉 focus 以 Gio focus 为准；键盘导航和默认动作以 runtime focus registry 为准，不能混写为同一个状态。 |
| `disabled` | 组件 API 配置或 item 配置，例如 `Disabled`、`InputDisabled`、`SelectDisabled`、`MenuItem.Disabled`、`TabItem.Disabled`。 | 组件配置结构的 `disabled` 字段；需要焦点时同步传给 `event.FocusDisabled(true)`。 | disabled content/container、跳过 click dispatch、focus target disabled、Decoration disabled state。 | 低到中。每个组件本地持有 disabled；event 只接收 focus disabled 投影。 | 权威来源是 widget 配置；event 不重新计算 disabled，只保存焦点可达性投影。 |
| `selected` | 组件 API 或 item 数据，例如 `ChipSelected`、`IconButtonSelected`、`ListItemSelected`、`MenuSelectedKey` / `MenuItem.Selected`。 | 组件配置字段 `selected`、菜单 `selectedKey` 与 `MenuItem.Selected`。 | `md3SelectionProgress`、selected container/foreground、check mark、indicator。 | 低。event 不参与 selected。 | 权威来源是 widget/app 语义状态；style 只提供动画时长和颜色工具。 |
| `error` | 表单组件 API，例如 `InputError`、`InputErrorText`、`SelectError`、`SelectErrorText`。 | `inputConfig.error`、`selectConfig.error` 与 error text。 | error border、label/supporting text、indicator color。 | 低。event 不参与 error。 | 权威来源是 widget/app 校验语义；不是输入事件状态。 |

## 视觉消费链路

```text
Gio pointer/focus state
  -> internal.ClickableState
  -> event.Clickable Snapshot/Hovered/Pressed/Focused
  -> widget resolves Decoration/state layer/ripple/focus indicator
  -> style supplies color/opacity/duration constants

Widget semantic state
  -> disabled / selected / error config fields
  -> widget blocks events or maps to event.FocusDisabled where needed
  -> widget picks Material colors/Decoration branch/selection progress
  -> style supplies disabled colors, state-layer colors and animation timing
```

## 事实结论

- `hovered` 与 `pressed` 的事实来源是 `internal.ClickableState` 对 Gio `widget.Clickable` 的封装；`event.Clickable` 负责把它们变成稳定的 `InteractionSnapshot`，并上报 runtime interaction diagnostics。
- `focused` 需要区分视觉焦点和组件树焦点：普通 clickable 控件的视觉 focus 来自 Gio focus，键盘导航、默认 activation 和 focus event 来自 runtime `focusTarget` registry。两者在当前实现中并非同一份状态。
- `resolveDecorationState` 只统一消费 `disabled`、`pressed`、`hovered`，优先级为 disabled > pressed > hovered；focused 的通用 Decoration 分支没有统一入口，Button/Tabs 等主要通过 focus indicator 表达，Input 对 `Decoration.Focused` 做了局部手动合并。
- `disabled` 是组件配置状态，不由 event 推导。组件会用它跳过 click dispatch、设置 `ReadOnly` 或向 `RegisterFocusTarget` 投影 `FocusDisabled(true)`，event registry 只保存焦点可达性投影。
- `selected` 是组件/app 语义状态，不属于输入事件。Menu、ListItem、IconButton、Chip 等组件以配置或 item 数据作为权威来源，再用 `md3SelectionProgress` 和 Material 颜色生成视觉。
- `error` 是表单校验语义状态，不属于输入事件。Input/Select 以 `InputError`、`SelectError` 配置作为权威来源，视觉层只消费为 error border、label/supporting text 颜色。
- `style` 包只提供颜色、opacity、duration、Decoration 结构和状态层工具；它不拥有 hover/pressed/focused/disabled/selected/error 的事实来源。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| Gio focus 与 runtime focus 名称相同但语义不同 | 中 | clickable 视觉 focus 用 `ctx.Gtx.Focused(&button)`，键盘事件和默认动作使用 `Runtime.focusTarget`。后续若把两者强行合并，可能破坏 keyboard focus event 或视觉 focus indicator。 |
| focused 未纳入通用 Decoration 状态解析 | 中 | 新组件如果只调用 `resolveDecorationState`，会自然覆盖 hover/pressed/disabled，但不会自动消费 `Decoration.Focused`。需要显式绘制 focus indicator 或手动合并 focused decoration。 |
| disabled 既在 widget 配置又在 focus registry 有投影 | 中 | `FocusDisabled(true)` 只是焦点可达性投影，不是 disabled 的权威来源。后续修复不能从 event focus target 反推组件 disabled。 |
| selected/error 与 event 状态混淆风险 | 低 | selected/error 当前完全由 widget/app 配置决定。若未来为它们新增事件，需要保持事件只是通知或默认动作，不应替代组件语义状态。 |
| 通用 `md3ActionSurface` 内部再次读取 clickable 状态 | 低 | 调用方有时先读取 `Snapshot`，`md3ActionSurface` 内部又调用 `Hovered/Pressed/Focused`；目前来源一致，但后续若扩展 side effect 应避免读两份不同状态。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出 hover、pressed、focused、disabled、selected、error 状态来源表 | 通过 | 已按输入状态和语义状态区分权威来源、入口和视觉消费。 |
| 每个状态只有一个权威来源 | 通过 | hover/pressed 归 event/internal clickable；focused 拆分为 Gio 视觉 focus 与 runtime component focus；disabled/selected/error 归 widget/app 配置。 |
| 避免 widget/event 各算一份 | 通过 | widget 对 hover/pressed/focused 主要消费 `event.Clickable`；event 对 disabled 只接收 `FocusDisabled` 投影，不反向维护组件 disabled。 |
| 风险边界明确 | 通过 | 已记录 focused 双语义、Decoration focused 缺省、disabled 投影、selected/error 非 event 状态。 |
| 未修改运行时代码 | 通过 | 本任务只新增审计子文件并更新索引。 |

## 后续依赖

- A4.3 ripple 和 state layer 审查应继续确认 `md3ActionSurface`、Button、Tabs、Select item、Chip remove button 的 state layer/ripple 绘制顺序是否只消费 event 状态，不扩大 hit area。
- A4.4 cursor 策略审查应沿用本结论：cursor hover 判断应以同一套 pointer/clickable target 为准，避免组件私自用窗口级状态覆盖整窗 cursor。
- A5 事件系统审查需要单独确认 Gio focus、runtime focus target、keyboard focus event 的映射关系，避免后续把视觉焦点和组件树焦点误合并。
- 建议补一组聚焦测试：focused decoration 消费边界、`FocusDisabled` 不等于组件 disabled 权威来源、selected/error 不触发 event interaction diagnostics。
