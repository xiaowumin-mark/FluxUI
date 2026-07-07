# A7.3 键盘默认行为审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：键盘、焦点和可访问性审查。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "### A7\.|A7\.3|键盘默认" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "Default|PreventDefault|keydown|keyup|Enter|Space|Escape|Arrow|RegisterFocusTarget|FocusActivate" internal event widget examples -g "*.go"`
  - `go test ./event ./internal ./widget ./examples/event_system_testbench ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/events.go`
  - `event/keyboard.go`
  - `event/keyboard_test.go`
  - `event/input.go`
  - `event/pointer.go`
  - `widget/event_defaults.go`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/selection.go`
  - `widget/material3_components.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/keyboard_scope.go`
  - `examples/event_system_testbench/main.go`
- 关联能力：
  - `keydown` 默认行为 gate 和 `PreventDefault` 生效条件
  - `Tab` focus move、`Enter/Space` focus activation 规则
  - Button、Select、Menu、RadioGroup、Checkbox、Switch 的键盘激活入口
  - Escape、Arrow keys 当前默认行为缺口
  - runtime keyboard default 与 widget-local click default 分层

## 事实结论

1. A7.3 的目标是审查 Button、Select、Menu、Dialog、Tabs、RadioGroup、Checkbox、Switch 的键盘默认行为，并输出 `Enter`、`Space`、`Escape`、方向键的默认行为表；验收要求是分清可取消默认行为和不可取消行为。证据：`docs/project-audit-task-breakdown.md:352`。
2. FluxUI 的 runtime 只在 `DispatchKeyboardEvent` 中对 `keydown` 运行 shortcut 和 keyboard default action；`keyup` 只进入普通事件分发和 key state 清理，不运行默认行为。证据：`internal/events.go:609`、`internal/events.go:625`、`internal/events.go:626`、`internal/events.go:632`。
3. `applyKeyboardDefaults` 会把 keyboard event 统一设为 bubbles、cancelable，并补时间戳；因此 runtime keyboard default action 可以被任何非 passive `keydown` listener 或 shortcut handler 的 `PreventDefault` 取消。证据：`internal/events.go:876`、`internal/events.go:879`、`internal/events.go:880`、`event/keyboard_test.go:59`。
4. runtime keyboard default action 当前只识别 `Tab`、`Enter`、换行、`Space` 和空格字符：`Tab`/`Shift+Tab` 移动焦点，`Enter`/`Space` 调用当前 focus target 的 `Activate`。证据：`internal/events.go:1040`、`internal/events.go:1044`、`internal/events.go:1051`。
5. `Escape` 和 Arrow keys 没有 runtime 级默认行为。`keyNameForEvent` 只把 Gio Return/Enter、Space、Escape、Tab 归一化成字符串；方向键保持 Gio 原始名称，例如 `←`、`→`、`↑`、`↓`。证据：`event/keyboard.go:299`、`event/keyboard.go:301`、`event/keyboard.go:305`、`event/keyboard.go:307`。
6. `activateFocusTarget` 只在目标是已注册且未 disabled/hidden 的 focus target，并且存在 `Activate` 回调时执行；没有 `Activate` 的焦点目标对 `Enter/Space` 没有默认动作。证据：`internal/events.go:1056`、`internal/events.go:1065`、`internal/events.go:1071`。
7. Button 在非 disabled/loading 时注册 `FocusActivate`，`Enter/Space` 会派发 click；disabled/loading 时只注册 disabled focus target，不触发键盘激活。证据：`widget/button.go:172`、`widget/button.go:174`。
8. Pressable/ClickArea 注册 `FocusActivate` 并复用 legacy click dispatcher；键盘激活路径等价于 click 默认语义。证据：`widget/click_area.go:63`。
9. Checkbox、Switch、RadioGroup row 都在非 disabled 时注册 `FocusActivate`，激活动作通过 `dispatchClickDefault` 先派发 cancelable click，只有 click 未被取消才执行 toggle/select。证据：`widget/checkbox.go:95`、`widget/checkbox.go:97`、`widget/switch.go:158`、`widget/switch.go:160`、`widget/selection.go:159`、`widget/selection.go:161`、`widget/event_defaults.go:8`。
10. Select field 注册 `FocusActivate`，`Enter/Space` 会走 click default gate 后切换 open 状态；Select option row 也注册 `FocusActivate`，激活时选择对应 option。证据：`widget/selection.go:759`、`widget/selection.go:761`、`widget/selection.go:564`、`widget/selection.go:566`。
11. Menu item 对非 disabled 且无 submenu 的叶子项注册 `FocusActivate`，`Enter/Space` 触发 selection；disabled item 或带 submenu item 不注册 activation，submenu 打开主要来自 hover/focus/pressed snapshot，而非方向键默认行为。证据：`widget/material3_components.go:552`、`widget/material3_components.go:554`。
12. DropdownMenu trigger 注册 `FocusActivate`，`Enter/Space` 会切换 open 状态；这是 trigger 层默认行为，不等价于 menu 内方向键导航。证据：`widget/material3_components.go:807`。
13. Tabs item 当前使用 `UseClickable` 和 pointer click 更新 active key，审查范围内未发现 tab item 调用 `RegisterFocusTarget`/`FocusActivate`，也未发现 ArrowLeft/ArrowRight 或 Home/End 导航逻辑；因此 Tabs 的 `Enter/Space`、Arrow keys 键盘默认行为目前是缺口。证据：`widget/tabs_dialog_toast.go:189`、`widget/tabs_dialog_toast.go:230`、`widget/tabs_dialog_toast.go:312`。
14. Dialog 本体没有 runtime 级 `Escape` 默认关闭逻辑；Dialog 的 action button 通过 Button 获得 `Enter/Space` 激活，mask click 可以关闭，但注释中提到的 `Escape` 取消没有在审查代码路径中找到键盘默认实现。证据：`widget/tabs_dialog_toast.go:919`、`widget/tabs_dialog_toast.go:967`、`widget/tabs_dialog_toast.go:1077`。
15. event system testbench P3 手动覆盖了 `Escape` stop propagation、`Tab` preventDefault、`Enter/Space` focus activation；P5 文案提到 Dialog 的 Escape 边界，但主要作为手动验收提示，不构成自动实现证据。证据：`examples/event_system_testbench/main.go:445`、`examples/event_system_testbench/main.go:451`、`examples/event_system_testbench/main.go:473`、`examples/event_system_testbench/main.go:767`。

### 键盘默认行为表

| Key | runtime 默认行为 | 组件落点 | 可取消性 | 当前结论 |
| --- | --- | --- | --- | --- |
| `Enter` | `keydown` 后激活当前 focus target | Button、Pressable/ClickArea、Checkbox、Switch、RadioGroup row、Select trigger/option、Menu leaf item、Dropdown trigger | 可由 keydown listener/shortcut `PreventDefault` 取消；进入 widget 后，合成 click 也可被 click listener 取消 | 已实现通用 activation |
| `Space` / `" "` | 同 `Enter` | 同 `Enter` | 同 `Enter` | 已实现通用 activation |
| `Escape` | 无 runtime 默认行为 | Dialog/Popup/Menu/Select 未发现统一键盘关闭 default | 因无默认动作，`PreventDefault` 目前只影响用户自定义 listener/shortcut，不会阻止内建关闭 | 缺口，需后续明确 |
| Arrow keys | 无 runtime 默认行为 | Tabs、RadioGroup、Menu、Select 未发现方向键导航 default | 因无默认动作，`PreventDefault` 目前不会阻止内建方向键导航 | 缺口，需后续明确 |
| `Tab` / `Shift+Tab` | 移动焦点 forward/backward | 所有 tabStop focus target | 可由 keydown listener/shortcut `PreventDefault` 取消 | 虽非 A7.3 输出重点，但这是现有唯一非 activation 键盘默认行为 |

### 组件默认行为矩阵

| 组件 | `Enter/Space` | `Escape` | Arrow keys | cancelable 边界 |
| --- | --- | --- | --- | --- |
| Button | 派发 click，再执行 `OnClick` | 无 | 无 | `keydown` 可取消；click 可取消 |
| Select trigger | 切换 open | 无内建关闭证据 | 无选项导航证据 | `keydown` 可取消；click 可取消 |
| Select option | 选择 option | 无 | 无 | `keydown` 可取消；click 可取消 |
| Menu leaf item | 触发 `onSelect` | 无内建关闭证据 | 无菜单导航证据 | `keydown` 可取消；click 可取消 |
| Dialog | action button 可激活 | 无 runtime Escape 关闭证据 | 无 | action button 同 Button；Dialog 本体无 default action |
| Tabs | 无键盘 activation 注册证据 | 无 | 无左右导航证据 | pointer click 可工作；键盘默认缺口 |
| RadioGroup | 选中当前 row | 无 | 无组内方向键导航证据 | `keydown` 可取消；click 可取消 |
| Checkbox | toggle | 无 | 无 | `keydown` 可取消；click 可取消 |
| Switch | toggle | 无 | 无 | `keydown` 可取消；click 可取消 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| Tabs 缺少键盘 activation 和方向键导航 | 高 | Tabs 有 focus visual snapshot，但未注册 focus target activation；键盘用户可能无法用 `Enter/Space` 或左右方向键切换 tab。 | A12、可访问性修复 |
| Dialog/Menu/Select 的 `Escape` 关闭语义未统一 | 高 | 文案和常见可访问性预期都指向 Escape 关闭 overlay，但 runtime 和组件侧未发现统一实现；后续不能假设已有 default action。 | A8、A12、可访问性修复 |
| Menu/Select/RadioGroup 缺少方向键导航 | 中 | 当前 option/row 可以被 focus activation 选择，但未发现上下/左右方向键在组内移动或打开 submenu 的默认行为。 | A8、A12 |
| keyboard default 与 widget click default 双层可取消 | 中 | `Enter/Space` 先经过 cancelable `keydown`，再合成 cancelable `click`；这给用户两次取消机会，但也要求文档明确两层 default action 的顺序。 | A5.5、Docs/API |
| 方向键 key name 未归一化为 DOM 字符串 | 中 | Gio 方向键保留为符号名称，若后续实现默认行为或用户监听 `ArrowLeft` 风格 key，需要先确定兼容映射。 | A5.1、A7.4、A12 |
| Input 与 runtime focus 分离仍会影响键盘默认归属 | 中 | A7.1/A7.2 已确认 Input 使用 Gio editor focus；若文本编辑场景下触发 `Enter/Space/Escape/Arrow`，需要在 A7.4/A7.5 专审。 | A7.4、A7.5 |

## 验收

- 已明确 runtime 现有键盘默认行为：`keydown` 下 `Tab` 移动焦点、`Enter/Space` 激活 focus target。
- 已明确 `Enter/Space` 的两层 cancelable gate：先由 keyboard event `PreventDefault` 取消 runtime default，再由合成 click `PreventDefault` 取消 widget-local action。
- 已逐项记录 Button、Select、Menu、Dialog、Tabs、RadioGroup、Checkbox、Switch 的 `Enter`、`Space`、`Escape`、Arrow keys 行为。
- 已标出 `Escape` 和 Arrow keys 当前没有内建默认行为，不能误判为已有可取消 default action。
- 已运行核心测试，确认本次只新增审查文档，没有引入 runtime/widget 行为变更。

## 后续依赖

- A7.4 文本输入事件审查需要确认 Input editor 中 `Enter`、`Escape`、Arrow keys 与 runtime keyboard default/shortcut 的冲突边界。
- A7.5 IME 和 submit 审查需要确认 composing/submit 状态下 `Enter` 是否应降级或阻止 activation。
- A8 overlay 审查需要专门验证 Dialog、Popup、Menu、Select 的 Escape 关闭、焦点恢复和 portal/boundary 顺序。
- A12 测试审查应补齐 keyboard accessibility 自动测试：Tabs `Enter/Space/Arrow`、Menu/Select `Escape/Arrow`、RadioGroup arrow navigation、Dialog Escape。
- Docs/API 指南应说明键盘 default action 的两层可取消顺序，避免使用者把 `keydown` default 与合成 `click` default 混淆。
