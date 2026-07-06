# A7.1 Focus target 注册审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：键盘、焦点和可访问性审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Runtime
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A7\.1|Focus target|FocusManager|KeyboardScope|focus target" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type FocusManager|FocusManager|KeyboardScope|FocusTarget|focusTarget|RequestFocus|TabIndex|Disabled|Hidden|Focusable|Focus" internal event ui widget layout style theme system examples -g "*.go"`
  - `rg -n "RegisterFocusTarget|FocusDisabled|FocusHidden|FocusTabIndex|KeyboardScope|RequestFocus|Focused\(" widget -g "*.go"`
  - `rg -n "RegisterFocusTarget|FocusDisabled|FocusActivate|key\.InputOp|key\.Focus|editor" widget/input.go internal -g "*.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/keyboard.go`
  - `event/keyboard_test.go`
  - `internal/events.go`
  - `widget/keyboard_scope.go`
  - `widget/button.go`
  - `widget/input.go`
  - `widget/selection.go`
  - `widget/tabs_dialog_toast.go`
- 关联能力：
  - runtime focus target 每帧注册和清理
  - `FocusManager`、`RegisterFocusTarget`、`RequestFocus` API 边界
  - `KeyboardScope` focusable、disabled、autoFocus、tabIndex 规则
  - Button/Select/Dialog/Input 的 focus target 归属
  - FluxUI focus path 与 event target path 对应关系

## 事实结论

1. A7.1 的目标是审查 `FocusManager`、`KeyboardScope`、`Input`、`Button`、`Select`、`Dialog` 的 focus target 注册、disabled、hidden、tab index 规则，并确认 focus path 和 target path 能对应。证据：`docs/project-audit-task-breakdown.md:336`、`docs/project-audit-task-breakdown.md:340`、`docs/project-audit-task-breakdown.md:341`。
2. FluxUI 的 focus target 权威注册入口是 `event.RegisterFocusTarget`，它把 `FocusTargetOptions` 收集后委托给 `Runtime.RegisterFocusTarget`；公开选项包括 `FocusTabIndex`、`FocusDisabled`、`FocusHidden`、`FocusActivate`。证据：`event/keyboard.go:57`、`event/keyboard.go:65`、`event/keyboard.go:72`、`event/keyboard.go:79`、`event/keyboard.go:86`。
3. `FocusManagerFor(ctx)` 只是绑定当前 context 的 runtime facade；`Target`、`Request`、`Blur`、`Move` 都直接调用 runtime 的 `FocusedTarget`、`RequestFocus`、`BlurFocus`、`MoveFocus`，自身不保存独立 focus 状态。证据：`event/keyboard.go:100`、`event/keyboard.go:111`、`event/keyboard.go:119`、`event/keyboard.go:127`、`event/keyboard.go:135`。
4. runtime 在每帧 `beginEventFrame` 中清空 `focusTargets`、`shortcuts` 和 `focusOrder`，并重置 `nextFocusOrder`；因此 focus target、tab order 和 shortcut listener 都是本帧布局树重建结果。证据：`internal/events.go:282`、`internal/events.go:293`、`internal/events.go:295`、`internal/events.go:296`。
5. runtime 在 `endEventFrame` 中检查当前 focused target 是否仍在本帧 `focusTargets` 中且 `focusable()`；如果目标缺失、disabled 或 hidden，则调用 `changeFocus(..., 0)` 清焦点。证据：`internal/events.go:302`、`internal/events.go:306`、`internal/events.go:307`、`internal/events.go:309`。
6. `focusTarget.focusable()` 的规则是未 disabled 且未 hidden；`tabStop()` 在 focusable 基础上要求 `TabIndex >= 0`。因此负 tab index 目标可被程序 `RequestFocus`，但不会进入 tab 导航。证据：`internal/events.go:252`、`internal/events.go:256`、`event/keyboard.go:57`。
7. `Runtime.RegisterFocusTarget` 会同时调用 `RegisterEventTarget(target, eventParentFor(target))`，所以 focus target 必然也是当前帧 event target；focus path 使用同一套 `eventPath`/parent 链。证据：`internal/events.go:402`、`internal/events.go:407`、`internal/events.go:423`、`internal/events.go:867`。
8. `RequestFocus` 只接受本帧已注册且 `focusable()` 的目标；未注册、disabled、hidden 目标都会返回 false。证据：`internal/events.go:447`、`internal/events.go:455`、`internal/events.go:456`。
9. `MoveFocus` 只基于 `sortedFocusOrder()` 返回的 tab stop 目标循环移动；排序规则是正 `TabIndex` 优先并按数值升序，其后按注册顺序稳定排序。证据：`internal/events.go:482`、`internal/events.go:785`、`internal/events.go:800`、`internal/events.go:805`、`internal/events.go:807`。
10. `KeyboardScope` 默认不注册 focus target；只有 `KeyboardScopeFocusable(true)`、`KeyboardScopeTabIndex(...)` 或 `KeyboardScopeAutoFocus(true)` 设置 `focusable` 后才注册。`KeyboardScopeDisabled(true)` 会直接布局 child 并跳过 listener、shortcut、focus target 注册。证据：`widget/keyboard_scope.go:55`、`widget/keyboard_scope.go:61`、`widget/keyboard_scope.go:67`、`widget/keyboard_scope.go:74`、`widget/keyboard_scope.go:139`、`widget/keyboard_scope.go:146`。
11. `KeyboardScopeAutoFocus(true)` 在第一次布局时调用 `RequestFocus(scopeCtx)`，并用 `keyboardScopeState.autoFocused` 防止每帧重复抢焦点。证据：`widget/keyboard_scope.go:149`、`widget/keyboard_scope.go:151`、`widget/keyboard_scope.go:162`。
12. `Button` 在 disabled 或 loading 时仍注册同一路径的 focus target，但标记 `FocusDisabled(true)`；可用时注册 `FocusActivate`，Enter/Space 默认激活会分发 click。证据：`widget/button.go:170`、`widget/button.go:172`、`widget/button.go:174`、`internal/events.go:1044`、`internal/events.go:1056`。
13. `Select` 的 field 和下拉 option row 都使用 runtime focus target：field disabled 时注册 disabled target，可用时注册 activate 打开/关闭；option row 在 select disabled 或 item disabled 时注册 disabled target，可用时注册 activate 选择并关闭。证据：`widget/selection.go:759`、`widget/selection.go:761`、`widget/selection.go:564`、`widget/selection.go:566`。
14. `RadioGroup` 等行项目也按行 context 注册 focus target，disabled 时注册 disabled target，可用时注册 activate；这说明列表型控件的 focus target path 是每个 row 的 PathID，而不是整个 group。证据：`widget/selection.go:159`、`widget/selection.go:161`。
15. `Input` 没有调用 `event.RegisterFocusTarget`。它使用 Gio `widget.Editor`、`key.FocusCmd` 和 `ctx.Gtx.Focused(editor)` 管理编辑器焦点；因此当前 Input 焦点不进入 FluxUI runtime `FocusedTarget()`/tab order 体系。证据：`widget/input.go:419`、`widget/input.go:421`、`widget/input.go:427`、`widget/input.go:442`、`rg widget/input.go` 未命中 `RegisterFocusTarget`。
16. `Dialog` 本体注册的是 portal 和 event boundary，而不是 focus target；对话框默认 action button 由内部 `TextButton` 注册 focus target，Dialog child 若需要焦点必须由其自身控件或 `KeyboardScope` 注册。证据：`widget/tabs_dialog_toast.go:1097`、`widget/tabs_dialog_toast.go:1103`、`widget/tabs_dialog_toast.go:1080`、`widget/button.go:170`。
17. focus 事件顺序已有单元测试覆盖：从 first 移动到 second 时依次发出 `blur:first`、`focusout:first`、`focus:second`、`focusin:second`，并校验 `RelatedTarget`。证据：`event/keyboard_test.go:12`、`event/keyboard_test.go:49`。
18. keydown 默认行为已有单元测试覆盖：可取消 Enter 激活和 Tab 移焦；未取消 Tab 时 focus 从 first 移到 second。证据：`event/keyboard_test.go:59`、`event/keyboard_test.go:96`、`event/keyboard_test.go:110`。

### focus target 注册矩阵

| 对象 | 注册路径 | disabled 规则 | hidden 规则 | tab index 规则 | path 对应结论 |
| --- | --- | --- | --- | --- | --- |
| `FocusManager` | 不注册，仅 facade | 由目标 entry 判定 | 由目标 entry 判定 | 由目标 entry 判定 | 使用 runtime 当前 focus target |
| `RegisterFocusTarget` | `ctx.PathID()` | `FocusDisabled(true)` 后不可 focus | `FocusHidden(true)` 后不可 focus | 默认 0；负数不进 tab order | 同时注册 event target，focus path 与 target path 同源 |
| `KeyboardScope` | `ctx.Scope("keyboard-scope")` | `KeyboardScopeDisabled` 跳过注册 | 无公开 hidden option | `KeyboardScopeTabIndex` 设置并启用 focusable | scope 自身是 target，child 是其后代 |
| `Button` | 按 button layout ctx | disabled/loading 注册 disabled target | 无显式 hidden option | 默认 0 | focus target 与点击 target 同一组件路径 |
| `Select field` | select field ctx | `SelectDisabled` 注册 disabled target | 无显式 hidden option | 默认 0 | field target 与 toggle 点击区域同一路径 |
| `Select option` | option row ctx | select disabled 或 item disabled 注册 disabled target | 无显式 hidden option | 默认 0 | 每个 option row 独立 target |
| `Dialog` | 本体不注册 focus target | 不适用 | closed 时不布局 overlay | 不适用 | portal/boundary 改写 event path，focus 来自内部控件 |
| `Input` | Gio editor tag | `ReadOnly = true`，不进 runtime focus | 无显式 hidden option | 不进 runtime tab order | Gio focus 与 FluxUI runtime focus 分离 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| `Input` 不进入 runtime focus 体系 | 高 | Input 使用 Gio editor focus，`FocusManager.Target()`、`MoveFocus(Tab)` 和 `KeyboardScope` 的 focused path 不会直接包含 Input；这会影响统一 tab order、focus event、shortcut scope 和 Dialog focus trap 的设计。 | A7.2、A7.3、A8、A13 |
| hidden 规则只有底层 API，组件层覆盖不足 | 中 | runtime 支持 `FocusHidden`，但 Button/Select/Input/Dialog 未暴露统一 hidden focus option；条件渲染会通过“不注册”清焦点，但可见性样式隐藏不一定同步为 hidden focus target。 | A7 可访问性、A13 收敛修复 |
| Dialog 没有显式初始焦点和焦点陷阱 | 中 | Dialog 只注册 portal/boundary，本体不注册 focus target；若内部没有可 focus 控件或没有 KeyboardScope，打开后 runtime focus 不一定进入 Dialog。 | A8 overlay、A7.3 focus trap |
| disabled target 仍进入 event target path | 低 | disabled 组件会注册 disabled focus target，同时仍是 event target；这是清理当前焦点所需，但后续事件逻辑不能把“是 event target”误解为“可 focus”。 | A5/A7 事件默认行为 |
| tab order 依赖本帧注册顺序 | 低 | 正 tab index 优先，其余按布局注册顺序；列表虚拟化、条件渲染和 portal 可能改变实际 tab 顺序，需要后续专项验证。 | A7.2、A8、A12 |

## 验收

- 已记录 runtime focus target 的创建、更新、清理规则：每帧重建 `focusTargets/focusOrder`，帧末对缺失、disabled、hidden 的当前焦点执行 blur。
- 已记录 disabled、hidden、tab index 的权威判定：`focusable = !Disabled && !Hidden`，`tabStop = focusable && TabIndex >= 0`。
- 已确认 focus target 与 event target path 的对应关系：`RegisterFocusTarget` 会同步注册 event target，focus 事件分发使用同一 event path。
- 已确认 `KeyboardScope` 的 focusable、disabled、autoFocus、tabIndex 行为，并标出 disabled scope 会跳过注册。
- 已确认 Button、Select 的 runtime focus target 注册路径，以及 Dialog/Input 的边界差异。
- 已标出 `Input` 使用 Gio editor focus 而非 FluxUI runtime focus，是 A7 后续任务的主要兼容风险。

## 后续依赖

- A7.2 键盘事件路由应重点处理 runtime focus 与 Gio editor focus 的分层：Input 文本编辑键、全局 shortcut、local shortcut 不能互相误吞。
- A7.3 tab order/focus trap 审查应验证 Dialog、Select popup、portal 内容是否需要独立 trap、restore focus、initial focus 规则。
- A8 overlay 审查应确认 Dialog/Menu/Select portal 的 event path 与 focus path 是否一致，尤其是关闭 overlay 后 restore focus 的语义。
- A12 测试审查应补充 disabled/hidden target 清焦点、负 tab index 可程序聚焦但不参与 Tab、KeyboardScope autoFocus 只执行一次、Input 不进入 runtime focus 的回归测试。
- A13 收敛修复应决定是否把 Input 纳入 FluxUI focus target，或在文档/API 中明确 Gio editor focus 是正式 escape hatch。
