# A10.1 基础交互控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Widget、Event、Style
- 范围：Button、Pressable、ClickArea、IconButton、FAB
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph explore "A10.1 Button Pressable ClickArea IconButton FAB OnClick OnHover ripple disabled focus"`
  - `codegraph explore "widget/material3_components.go IconButton iconButtonWidget FloatingActionButton FAB Layout OnClick OnHover disabled ripple focus"`
  - `codegraph explore "widget/button.go buttonWidget Layout state layer ripple pressed hover focus disabled DispatchClickEvent"`
  - `codegraph explore "event Dispatcher DispatchClickEvent DispatchHoverEvent RegisterFocusTarget UseClickable LayoutButton hover pressed ripple"`
  - `codegraph explore "md3ActionSurface md3ActionSurfaceSpec layoutStateLayerTouchTarget LayoutRippleArea IconButton FloatingActionButton"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/material3_components.go`
  - `widget/utils.go`
  - `widget/event_defaults.go`
  - `event/dispatcher.go`
  - `event/pointer.go`
  - `event/keyboard.go`
  - `internal/render.go`
  - `internal/ripple.go`

## 事实结论

1. A10.1 的目标是审查基础交互控件族的 click、hover、pressed、focus、ripple、disabled 矩阵，并把旧 `OnClick`/`OnHover` 与新事件边界写清楚。该范围对应路线图中的基础交互控件族：Button、Pressable、ClickArea、IconButton、FAB。证据：`docs/project-audit-roadmap.md:306`、`docs/project-audit-task-breakdown.md:436`、`docs/project-audit-task-breakdown.md:440`、`docs/project-audit-task-breakdown.md:441`、`docs/project-audit-task-breakdown.md:442`。
2. `Button` 系列通过 `buttonConfig.dispatcher` 保存旧 `OnClick`/`OnHover`，布局时使用 `event.UseClickable` 获取 Gio clickable 状态。非 disabled/loading 时注册 runtime focus target，并在键盘默认激活、Ref 命令、Gio clicked event 三个入口调用 `Dispatcher.DispatchClickEvent`。证据：`widget/button.go:27`、`widget/button.go:88`、`widget/button.go:95`、`widget/button.go:169`、`widget/button.go:172`、`widget/button.go:174`、`widget/button.go:180`、`widget/button.go:187`。
3. `Button` 的新旧 click 边界是“先新事件，后旧回调”：`DispatchClickEvent` 构造/补齐 cancelable `click` 事件，调用 `DispatchPointerEvent` 后，只有未被取消时才执行旧 `Click` handler。因此 `PreventDefault` 可以阻止旧 `OnClick`。证据：`event/dispatcher.go:16`、`event/dispatcher.go:18`、`event/dispatcher.go:22`、`event/dispatcher.go:27`、`event/dispatcher.go:28`、`event/dispatcher.go:41`、`event/dispatcher.go:42`、`event/dispatcher.go:44`。
4. `Button` 的旧 `OnHover` 不经过 FluxUI typed hover event。它来自 `clickable.Snapshot(ctx, true)` 的 `Hovered` 值，再由 `HoverChangedWithSnapshot` 做变化过滤，变化时直接调用 `dispatcher.DispatchHover`。证据：`widget/button.go:194`、`widget/button.go:195`、`event/pointer.go:130`、`event/pointer.go:152`、`event/pointer.go:164`、`event/dispatcher.go:47`、`event/dispatcher.go:48`。
5. `Button` 的 pressed/focus 视觉来源来自 `event.Clickable` snapshot。hover/pressed 会通过 `style.StateLayer` 叠加背景；focus 通过 `md3FocusProgress` 写入 `ButtonSpec.FocusOpacity`，最后由 `internal.Context.LayoutButton` 绘制 focus indicator。证据：`widget/button.go:194`、`widget/button.go:199`、`widget/button.go:209`、`widget/button.go:212`、`widget/button.go:238`、`widget/button.go:257`、`internal/render.go:815`、`internal/render.go:818`。
6. `Button` 的 ripple 由 `internal.Context.LayoutButton` 在 clickable 非 nil 且未 disabled 时绘制。ripple 使用 `DrawRipple`，绘制在按钮背景阶段，不改变 layout size；disabled 或 loading 时 Button 不消费 click，但当前传给 `LayoutButton` 的 `Disabled` 只包含 `b.config.disabled`，loading 依赖上层跳过事件分发和替换内容，不等同于 render disabled。证据：`widget/button.go:169`、`widget/button.go:170`、`widget/button.go:187`、`widget/button.go:263`、`internal/render.go:747`、`internal/render.go:757`、`internal/render.go:758`、`internal/render.go:783`。
7. `Pressable` 是无固定视觉样式的通用点击区域，`ClickArea` 是 deprecated alias。二者通过 `event.Dispatcher{Click: onClick}` 桥接新 `click` 事件和旧 callback，使用 `LayoutClickArea` 注册命中区域，不绘制 hover/pressed/focus/ripple。证据：`widget/click_area.go:20`、`widget/click_area.go:22`、`widget/click_area.go:23`、`widget/click_area.go:27`、`widget/click_area.go:61`、`widget/click_area.go:62`、`widget/click_area.go:77`、`widget/click_area.go:80`。
8. `Pressable`/`ClickArea` 当前没有 disabled option，也没有旧 `OnHover`。它们始终注册 focus target，键盘默认激活和 Ref 命令都会派发 `click`，并遵守 `PreventDefault` 对旧 `onClick` 的取消规则。证据：`widget/click_area.go:11`、`widget/click_area.go:16`、`widget/click_area.go:40`、`widget/click_area.go:44`、`widget/click_area.go:63`、`widget/click_area.go:68`、`widget/click_area.go:69`。
9. `IconButton` 系列配置独立于 `ButtonOption`，只有 `IconButtonOnClick`、disabled、selected、loading、size、颜色和 decoration 等选项，没有 `OnHover`，也没有复用 `event.Dispatcher`。布局时直接循环 `clickable.Clicked(ctx)` 并调用旧 `onClick`。证据：`widget/material3_components.go:1084`、`widget/material3_components.go:1095`、`widget/material3_components.go:1139`、`widget/material3_components.go:1143`、`widget/material3_components.go:1176`、`widget/material3_components.go:1178`、`widget/material3_components.go:1179`。
10. `IconButton` 的 hover/pressed/focus/ripple 视觉由 `md3ActionSurface` 消费 `event.Clickable` 实现：`Hovered`、`Pressed`、`Focused(ctx)` 决定 state layer、颜色动画、ripple overlay 和 focus indicator。disabled/loading 会阻止 `onClick`；但 IconButton 没有注册 runtime focus target，也没有通过 FluxUI `click` 分发，因此键盘默认激活、`PreventDefault`、capture/bubble listener 与旧 `onClick` 没有统一桥接。证据：`widget/material3_components.go:105`、`widget/material3_components.go:116`、`widget/material3_components.go:117`、`widget/material3_components.go:118`、`widget/material3_components.go:127`、`widget/material3_components.go:201`、`widget/material3_components.go:215`、`widget/material3_components.go:221`。
11. `FAB`/`SmallFloatingActionButton`/`LargeFloatingActionButton`/`ExtendedFloatingActionButton` 同样使用独立 `FloatingActionButtonOnClick`，布局时直接消费 `clickable.Clicked(ctx)`，再交给 `md3ActionSurface` 绘制 surface、state layer、ripple、focus indicator。FAB 有 disabled，但没有 loading、旧 `OnHover`、新 `click` 分发和 runtime focus target 注册。证据：`widget/material3_components.go:1284`、`widget/material3_components.go:1296`、`widget/material3_components.go:1305`、`widget/material3_components.go:1317`、`widget/material3_components.go:1337`、`widget/material3_components.go:1339`、`widget/material3_components.go:1388`。
12. `md3ActionSurface` 的 disabled 只影响视觉和 ripple/点击区域布局路径：disabled 时设置 disabled foreground/container，不调用 `LayoutRippleOverlayArea`。它可以绘制 focus indicator，但 focus 能否由键盘进入取决于调用方是否注册 focus target；IconButton/FAB 当前缺少这一步。证据：`widget/material3_components.go:121`、`widget/material3_components.go:122`、`widget/material3_components.go:123`、`widget/material3_components.go:199`、`widget/material3_components.go:201`、`widget/material3_components.go:215`、`event/keyboard.go:86`、`event/keyboard.go:87`。

### 交互矩阵

| 控件 | click 来源 | 旧 OnClick | 新 click 事件 | OnHover | pressed/hover 视觉 | focus | ripple | disabled/loading |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Button 系列 | Gio clicked、keyboard focus activate、ButtonRef | 新事件未取消后触发 | 有，cancelable/bubbles | 有，仅旧回调，变化过滤 | `Clickable.Snapshot` + state layer | 注册 runtime focus target，绘制 focus indicator | `LayoutButton` bounded ripple | disabled/loading 阻止事件；disabled 视觉明确，loading 替换内容但不等于 render disabled |
| Pressable | Gio clicked、keyboard focus activate、PressableRef/ClickAreaRef | 新事件未取消后触发 | 有，cancelable/bubbles | 无 | 无固定视觉 | 始终注册 runtime focus target | 无 | 无 disabled option |
| ClickArea | 同 Pressable | 同 Pressable | 同 Pressable | 无 | 无固定视觉 | 同 Pressable | 无 | deprecated alias，无 disabled option |
| IconButton 系列 | Gio clicked | 直接触发 | 无统一分发 | 无 | `md3ActionSurface` state layer | 只读取 Gio clickable focus；未注册 runtime focus target | `LayoutRippleOverlayArea` | disabled/loading 阻止旧 onClick；disabled 视觉由 `md3ActionSurface` |
| FAB 系列 | Gio clicked | 直接触发 | 无统一分发 | 无 | `md3ActionSurface` state layer | 只读取 Gio clickable focus；未注册 runtime focus target | `LayoutRippleOverlayArea` | disabled 阻止旧 onClick；无 loading 选项 |

## 风险

1. IconButton/FAB 与 Button/Pressable 的事件边界不一致：它们不经过 `DispatchPointerEvent`，所以 capture/target/bubble listener、`PreventDefault`、事件审计链路对旧 `onClick` 不生效。
2. IconButton/FAB 没有 runtime focus target 注册。`md3ActionSurface` 能读取 Gio clickable focus 并绘制 focus indicator，但缺少 FluxUI focus path、keyboard default activation 和 tab order 语义。
3. IconButton/FAB 没有旧 `OnHover`，也没有新 hover typed event；调用方无法像 Button 那样收到 hover change callback。
4. Button loading 状态阻止事件和 Ref 命令，但 `ButtonSpec.Disabled` 只反映 disabled，不反映 loading；loading 下仍按非 disabled surface/ripple 参数进入 render 语义，后续若调整 ripple/命中需要避免把 loading 误判为完整 disabled。
5. Pressable/ClickArea 没有 disabled option，作为低层 escape hatch 可接受，但如果业务把它当按钮使用，会缺少 disabled、hover、pressed、ripple 和视觉反馈的一致语义。
6. `ClickArea` 标注 deprecated 但仍是公开兼容入口；后续修复不得删除签名或改变其作为 Pressable alias 的行为。

## 验收

- 已输出 Button、Pressable、ClickArea、IconButton、FAB 的 click、hover、pressed、focus、ripple、disabled/loading 矩阵。
- 已明确旧 `OnClick`/`OnHover` 与新事件边界：Button/Pressable/ClickArea 走新 cancelable click 后旧回调；IconButton/FAB 当前只直接调用旧 onClick。
- 已标出 Button 的 `PreventDefault` 可取消旧 `OnClick`，以及 `OnHover` 仍是旧回调而非 typed hover event。
- 已标出 IconButton/FAB 的主要缺口：无新 click 分发、无 runtime focus target、无旧 OnHover。
- 已确认本任务只做审查记录，没有修改运行时代码或组件行为。

## 后续依赖

- A10 后续修复若要统一基础交互族，应优先抽出共享 action-control 事件桥接，至少让 IconButton/FAB 支持 cancelable click dispatch 和 runtime focus target。
- 建议补充回归测试：Button `PreventDefault` 阻止旧 `OnClick`、Pressable/ClickArea keyboard activation、IconButton/FAB 是否参与 runtime focus path、IconButton/FAB click listener 是否可取消。
- API 文档应明确：Button 的 `OnHover` 是旧 callback；Pressable/ClickArea 是无样式低层点击区域；IconButton/FAB 当前不是 ButtonOption 家族，事件语义与 Button 不完全一致。
