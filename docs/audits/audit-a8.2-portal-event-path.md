# A8.2 Portal event path 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Runtime
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `codegraph explore "RegisterPortal RegisterBoundary DispatchEvent Dialog Popup modal boundary owner target"`
  - `rg -n "RegisterPortal|RegisterBoundary|StopPropagation|portal|Portal|boundary|Boundary|dispatch|Dispatch|bubble|capture|owner" event internal widget ui -g "*.go"`
  - `go test ./event ./internal ./widget ./examples/component_lab ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/events.go`
  - `event/event.go`
  - `event/event_test.go`
  - `widget/event_boundary.go`
  - `widget/tabs_dialog_toast.go`
  - `ui/extended_types.go`

## 事实结论

1. A8.2 的目标是审查 portal registration、owner target、modal boundary，输出 overlay/portal/dialog/modal 的 event path 规则；验收要求 Popup 能按规则冒泡到 owner，Dialog modal 能按规则截断。证据：`docs/project-audit-task-breakdown.md:388`、`docs/project-audit-task-breakdown.md:392`、`docs/project-audit-task-breakdown.md:393`、`docs/project-audit-task-breakdown.md:394`。
2. 公开 API 层的 `event.RegisterPortal(ctx, owner)` 只做空值保护并委托给 runtime；注释明确它用于 layout parent 与 component owner 不一致的 overlay/portal root。`event.RegisterBoundary(ctx, opts...)` 默认 `BoundaryStop`，可通过 `BoundaryRedirectTo(target)` 改成 redirect。证据：`event/event.go:84`、`event/event.go:92`、`event/event.go:114`、`event/event.go:128`。
3. runtime portal 注册规则是：取当前 `ctx.pathID` 作为 portal target，把 `owner` 规范化；若 owner 为 0，则回退到 root；随后用 `RegisterEventTargetOptions(target, owner, EventTargetOptions{})` 改写该 target 的逻辑父级。证据：`internal/events.go:367`、`internal/events.go:371`、`internal/events.go:372`、`internal/events.go:376`。
4. runtime boundary 注册规则是：以当前 context path 作为 target，保留原本 `eventParentFor(target)` 作为父级，同时在 `EventTargetOptions` 中记录 boundary policy；`RegisterEventTargetOptions` 会更新已存在 target 的 Parent，并且只在 policy 非空时覆盖 Boundary。证据：`internal/events.go:320`、`internal/events.go:337`、`internal/events.go:340`、`internal/events.go:353`。
5. event path 构造从 dispatch target 开始向上走 `eventParentFor(current)`；遇到 `EventBoundaryStop` 直接返回当前 path，遇到 `EventBoundaryRedirect` 则把下一跳 parent 改为 redirect target，redirect 为空时回到 root。path 有 256 个节点上限，并用 `pathIndex` 防循环。证据：`internal/events.go:1074`、`internal/events.go:1082`、`internal/events.go:1083`、`internal/events.go:1085`、`internal/events.go:1087`、`internal/events.go:1100`。
6. `DispatchEvent` 在派发前会确保 target 已注册，并把 `event.path` 固定为 `r.eventPath(target)`；capture 从 root 方向倒序执行到 target 的父级，target 阶段先 capture listener 后 bubble listener，bubble 阶段按 `path[1:]` 从 target 父级向 root 执行。证据：`internal/events.go:637`、`internal/events.go:645`、`internal/events.go:661`、`internal/events.go:668`、`internal/events.go:676`、`internal/events.go:685`。
7. 普通 portal 的路径规则已经有单元测试覆盖：`RegisterPortal(portal, owner)` 后，portal 内部 target 的 composed path 是 `insidePortal -> portal -> owner -> root`，owner 和 root 的 bubble listener 都会被调用。证据：`event/event_test.go:411`、`event/event_test.go:415`、`event/event_test.go:428`、`event/event_test.go:431`。
8. stop boundary 的路径规则也有单元测试覆盖：`RegisterBoundary(boundary)` 后，内部事件 composed path 是 `insideBoundary -> boundary`，root listener 不会被调用；redirect boundary 覆盖 `insideRedirect -> redirectBoundary -> redirectTarget -> root`。证据：`event/event_test.go:435`、`event/event_test.go:448`、`event/event_test.go:451`、`event/event_test.go:457`、`event/event_test.go:472`。
9. `widget.EventPortal(child, owner)` 是组件级 portal wrapper：它创建 `ctx.Scope("event-portal")`，注册 portal 到 owner，然后把 child 布局在 `portalCtx.Child(0)`。`widget.EventBoundary` 默认 stop，也支持 redirect。证据：`widget/event_boundary.go:67`、`widget/event_boundary.go:73`、`widget/event_boundary.go:80`、`widget/event_boundary.go:90`、`widget/event_boundary.go:94`。
10. Dialog modal 的事件路径是 portal + stop boundary：Dialog 在 layout 时先保存 owner 为 `ctx.PathID()`，在 centered overlay 内创建 `dialog-portal` scope，先 `RegisterPortal(portalCtx, owner)`，再 `RegisterBoundary(portalCtx, BoundaryStopPropagation())`。因此面板内部事件路径会包含 dialog portal target，但不会继续冒泡到 owner/root。证据：`widget/tabs_dialog_toast.go:1072`、`widget/tabs_dialog_toast.go:1073`、`widget/tabs_dialog_toast.go:1074`、`widget/tabs_dialog_toast.go:1075`、`widget/tabs_dialog_toast.go:1076`。
11. Popup 的事件路径是 portal-only：Popup 同样保存 owner，创建 `popup-portal` scope 并注册 portal 到 owner，但没有注册 stop boundary。因此 popup panel 内部 FluxUI 事件按 portal 规则继续冒泡到 owner，再继续沿 owner 的路径冒泡。证据：`widget/tabs_dialog_toast.go:1921`、`widget/tabs_dialog_toast.go:1922`、`widget/tabs_dialog_toast.go:1923`、`widget/tabs_dialog_toast.go:1924`、`widget/tabs_dialog_toast.go:1925`。
12. Dialog 和 Popup 的 Gio 层 pointer 穿透由 `modalPanelInputShield` 处理，它在面板实际尺寸区域注册 Gio pointer tag 并消费 press/release/move/drag/enter/leave/cancel/scroll；这不等同于 FluxUI event path 截断，真正的 FluxUI modal 截断只来自 `RegisterBoundary(...StopPropagation())`。证据：`widget/tabs_dialog_toast.go:1617`、`widget/tabs_dialog_toast.go:1624`、`widget/tabs_dialog_toast.go:1625`、`widget/tabs_dialog_toast.go:1639`、`widget/tabs_dialog_toast.go:1641`。

### Event path 规则表

| 场景 | 注册点 | composed path 规则 | owner 可观察 | root 可观察 | modal 截断 |
| --- | --- | --- | --- | --- | --- |
| 普通子树 | 无 portal/boundary | target -> layout parent ... -> root | 若 owner 是祖先则可观察 | 可观察 | 否 |
| `EventPortal` / Popup | portal root 注册到 owner | target -> portal root -> owner -> owner ancestors -> root | 可观察 | 可观察 | 否 |
| Dialog modal | portal root 注册到 owner，同时同一 portal root stop boundary | target -> dialog portal root 截止 | 不可观察 panel 内部事件 | 不可观察 panel 内部事件 | 是 |
| `EventBoundary` stop | boundary target 注册 stop | target -> boundary 截止 | 若在 boundary 外则不可观察 | 不可观察 | 是 |
| `EventBoundary` redirect | boundary target 注册 redirect | target -> boundary -> redirect target -> redirect ancestors | redirect target 可观察 | 可观察 | 否，改写父级 |

## 风险

1. Dialog 在同一个 `portalCtx` 上先注册 portal 再注册 stop boundary；当前 `RegisterEventTargetOptions` 保留 Parent 并追加 Boundary，因此行为正确。但这是隐含依赖，后续改动注册表更新逻辑时必须保留“portal parent + boundary policy 可共存”的语义。
2. Popup 是 portal-only，符合本次验收“能冒泡到 owner”；如果后续把 Popup 当作严格 modal 使用，需要显式增加 boundary 或新增配置，不能假设 `modalPanelInputShield` 会截断 FluxUI 事件冒泡。
3. `RegisterEventPortal` 的 owner 为 0 时会回退 root；调用方漏传 owner 不会失败，而是变成 root-owned portal，可能让父组件无法观察本应归属自己的 popup 事件。
4. stop boundary 截断的是 FluxUI composed path，不是 Gio 原始 input routing；面板/遮罩的原始 pointer 层仍依赖 Gio tag、clip 和布局顺序，后续 A8.3 outside click 审查需要继续验证两层语义是否一致。
5. 当前测试覆盖了 runtime portal/boundary 基本路径，但没有直接覆盖 Dialog 与 Popup 组件布局后的事件路径差异；组件级回归主要依赖源码审查和较宽的 widget/example 测试。

## 验收

- 已明确 portal registration 会把 portal root 的逻辑父级改写到 owner，普通 portal / Popup 的路径会继续冒泡到 owner 和 root。
- 已明确 Dialog modal 在 portal root 上额外注册 stop boundary，因此 panel 内部事件路径在 dialog portal root 截断，不继续冒泡到 owner。
- 已输出 overlay/portal/dialog/modal 的 event path 规则表，区分普通子树、portal-only、modal stop boundary、redirect boundary。
- 已标出 Popup portal-only 与 Dialog portal + stop boundary 的语义差异，后续修复不得把两者混同。

## 后续依赖

- A8.3 outside click 审查应复用本文结论，把 FluxUI event path 与 Gio pointer pass-through / protected rect 分开验证。
- A8.4 focus trap 审查需要确认 Dialog 的 modal event boundary 是否与 focus scope/focus restore 同步，Popup 是否仍允许 owner 观察键盘事件。
- A10.6 Overlay 控件族审查应把 `RegisterPortal`、`RegisterBoundary`、`modalPanelInputShield` 的重复使用点纳入统一矩阵。
