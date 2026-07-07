# A8.4 overlay focus 和 Escape 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph explore "A8.4 overlay focus Escape Dialog Popup Menu Select KeyboardScope focus restore Escape"`
  - `codegraph explore "widget/keyboard_scope.go KeyboardScope Layout autoFocus OnKeyboard Escape focusable RequestFocus FocusManager"`
  - `codegraph explore "widget/tabs_dialog_toast.go dialogWidget Layout Popup Escape OnKeyboard RequestFocus restore focus"`
  - `codegraph explore "widget/selection.go Select Escape keydown opened dropdown focus restore"`
  - `rg -n "Escape|KeyEscape|KeyNameEscape|OnKeyboard|KeyDown|RequestFocus|BlurFocus|FocusManager|Restore|opened|onOpen|outside|DropdownMenu|popupWidget|modalPanelInputShield|RegisterFocusTarget" widget internal event examples -g "*.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/keyboard_scope.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/material3_components.go`
  - `widget/selection.go`
  - `event/keyboard.go`
  - `internal/events.go`

## 事实结论

1. A8.4 的目标是审查 Dialog、Popup、Menu、Select、KeyboardScope 的打开时 focus、关闭时 restore、Escape 捕获规则；验收要求 modal 和非 modal 的 focus 行为有差异且可解释。证据：`docs/project-audit-task-breakdown.md:404`、`docs/project-audit-task-breakdown.md:408`、`docs/project-audit-task-breakdown.md:409`、`docs/project-audit-task-breakdown.md:410`。
2. KeyboardScope 是当前公开的通用键盘/focus 作用域：`KeyboardScopeFocusable`、`KeyboardScopeTabIndex`、`KeyboardScopeAutoFocus` 都会把 scope 变成 focus target；Layout 时先注册监听器，再按需注册 focus target，`autoFocus` 首次布局时调用 `RequestFocus(scopeCtx)`。证据：`widget/keyboard_scope.go:61`、`widget/keyboard_scope.go:67`、`widget/keyboard_scope.go:74`、`widget/keyboard_scope.go:143`、`widget/keyboard_scope.go:146`、`widget/keyboard_scope.go:148`。
3. KeyboardScope 的键盘事件只在当前 runtime focus target 位于该 scope 的 event path 内时从 Gio drain `key.Event`，然后转换为 FluxUI `KeyboardEvent` 并派发到 focused target；因此它是局部捕获入口，不是全局窗口级 Escape hook。证据：`widget/keyboard_scope.go:186`、`widget/keyboard_scope.go:190`、`widget/keyboard_scope.go:191`、`widget/keyboard_scope.go:195`、`widget/keyboard_scope.go:200`、`widget/keyboard_scope.go:201`。
4. Runtime 键盘默认行为只内建 Tab focus move 和 Enter/Space focus activation；`Escape` 只被 Gio key name 映射成 `"Escape"`，没有 runtime 默认关闭行为。证据：`internal/events.go:626`、`internal/events.go:1040`、`internal/events.go:1044`、`internal/events.go:1045`、`internal/events.go:1051`、`event/keyboard.go:299`、`event/keyboard.go:305`。
5. Runtime focus registry 每帧记录 focus target，并在帧结束时若当前 focus target 不存在或不可 focus，会 `changeFocus(..., 0)` 清空；存在 `RequestFocus`、`BlurFocus`、`MoveFocus`，但没有“打开 overlay 前记住旧 focus、关闭后 restore”的通用栈。证据：`internal/events.go:301`、`internal/events.go:306`、`internal/events.go:307`、`internal/events.go:310`、`internal/events.go:447`、`internal/events.go:467`、`internal/events.go:482`。
6. Dialog/Popup 的打开/关闭生命周期只处理 `onOpen`、`onOpened`、`onClose`、`onClosed`、`onOpenChange`、mask click 和 portal/boundary；Layout 内没有注册 `RegisterFocusTarget`、`RequestFocus`、`OnKeyboard` 或 Escape listener。Dialog portal 还注册 stop boundary，Popup 只注册 portal。证据：`widget/tabs_dialog_toast.go:984`、`widget/tabs_dialog_toast.go:1025`、`widget/tabs_dialog_toast.go:1043`、`widget/tabs_dialog_toast.go:1101`、`widget/tabs_dialog_toast.go:1102`、`widget/tabs_dialog_toast.go:1103`、`widget/tabs_dialog_toast.go:1886`、`widget/tabs_dialog_toast.go:1923`。
7. Dialog/Popup 内部 pointer 被 `modalPanelInputShield` 保护，但该 shield 只注册 Gio pointer tag 并消费 pointer/scroll 事件，不承担键盘 focus trap 或 Escape 处理。证据：`widget/tabs_dialog_toast.go:1104`、`widget/tabs_dialog_toast.go:1919`、`widget/tabs_dialog_toast.go:1640`、`widget/tabs_dialog_toast.go:1641`、`widget/tabs_dialog_toast.go:1644`。
8. DropdownMenu trigger 自身注册 focus target，并通过 focus activation 或 pointer click 切换 open；menu row 也注册 focus target 和 activation。打开后 outside click helper 只处理 pointer press；没有默认 auto-focus first item、focusout close、restore focus 或 Escape 关闭实现。相关配置字段已存在：`defaultFocus`、`stayOpenOnFocusout`、`skipRestoreFocus`。证据：`widget/material3_components.go:351`、`widget/material3_components.go:358`、`widget/material3_components.go:359`、`widget/material3_components.go:807`、`widget/material3_components.go:815`、`widget/material3_components.go:542`、`widget/material3_components.go:552`、`widget/material3_components.go:919`。
9. 普通 Menu 是纯 panel，不持有 open 状态；它的 row 可成为 focus target 并支持 Enter/Space activation，但 `MenuDefaultFocus`、`MenuStayOpenOnFocusout`、`MenuSkipRestoreFocus` 当前只是配置入口，没有在 `menuWidget.Layout` 内触发 focus/restore/Escape 逻辑。证据：`widget/material3_components.go:339`、`widget/material3_components.go:351`、`widget/material3_components.go:358`、`widget/material3_components.go:359`、`widget/material3_components.go:470`、`widget/material3_components.go:524`、`widget/material3_components.go:542`。
10. Select field 注册 focus target，Enter/Space 激活会切换 `state.opened`；option row 也注册 focus target，激活后选择并关闭。Select 的 popup outside helper 只处理 protected rect 外 pointer press；没有默认打开后 focus option、关闭后 restore field focus 或 Escape 关闭实现。证据：`widget/selection.go:729`、`widget/selection.go:737`、`widget/selection.go:740`、`widget/selection.go:752`、`widget/selection.go:520`、`widget/selection.go:542`、`widget/selection.go:545`、`widget/selection.go:706`。
11. 示例中有使用 KeyboardScope 手写 Escape stop 的入口，说明 Escape 可以由应用层/组件层注册局部 listener；但这不是 Dialog/Popup/Menu/Select 的默认 overlay 行为。证据：`examples/docs_browser/event_system_demo.go:186`、`examples/docs_browser/event_system_demo.go:187`、`examples/docs_browser/event_system_demo.go:188`、`examples/event_system_testbench/main.go:447`、`examples/event_system_testbench/main.go:449`。

### focus/Escape 行为矩阵

| 组件 | 打开时 focus | 关闭时 restore | Escape 捕获 | modal/非 modal 差异 |
| --- | --- | --- | --- | --- |
| `Dialog` | 无默认 focus target/auto-focus；内部控件可自行注册 focus | 无 restore 栈；关闭后若原 focus target 消失，Runtime 可能清空 focus | 无默认 Escape 关闭；只能由内部 KeyboardScope 或调用方处理 | modal 体现在 portal + stop boundary + mask/panel pointer shield，不体现在 focus trap |
| `Popup` | 无默认 focus target/auto-focus | 无 restore 栈 | 无默认 Escape 关闭 | 非 modal/轻量 portal：注册 portal，不注册 Dialog 的 stop boundary；focus 行为仍与普通树一致 |
| `DropdownMenu` | trigger 是 focus target；menu item 是 focus target，但默认不把 focus 移入菜单 | `skipRestoreFocus` 有配置入口，未见 restore 实现 | 无默认 Escape 关闭 | 非 modal，本地 op.Defer popup；focus 不被 trap |
| `Menu` | row 是 focus target/activation target；panel 自身不自动 focus | `skipRestoreFocus` 有配置入口，未见 restore 实现 | 无默认 Escape 关闭 | 纯 panel，无 overlay ownership |
| `Select` | field 是 focus target；option row 是 focus target，但打开后不自动 focus option | 无 restore；field focus 只由 Runtime 当前 focus 状态自然保留/清空 | 无默认 Escape 关闭 | 非 modal，本地 popup；focus 不被 trap |
| `KeyboardScope` | 可通过 `KeyboardScopeAutoFocus(true)` 首次布局请求 focus | 无 restore；可通过 FocusManager/调用方实现 | 可局部监听 Escape 并 stop/prevent/关闭外部状态 | 取决于包裹位置；自身不区分 modal/非 modal |

## 风险

1. 文档/API 暴露了 `MenuDefaultFocus`、`MenuStayOpenOnFocusout`、`MenuSkipRestoreFocus` 等配置入口，但本轮未发现对应实现；如果用户期待 Material/Web 菜单语义，会误以为打开菜单会自动 focus first item、focusout 自动关闭或关闭后恢复 trigger focus。
2. Dialog/Popup 已有 modal 事件边界，但没有 focus trap；键盘 Tab 默认行为仍按全局当前帧 tab order 移动，理论上可以越过 modal panel 到背景 focus target。
3. Dialog/Popup 的 `onCancelOnly` 注释包含 `Escape`，但当前代码只在 mask click 路径调用；没有默认 Escape 触发该 callback 的实现。
4. Select/DropdownMenu 关闭后 focus 依赖现有 focus target 是否仍存在，缺少显式 restore 会导致 keyboard 用户可能停留在已关闭 popup 的旧 option target 或被帧结束清空。
5. Escape 当前是普通 cancelable keydown，应用层可用 KeyboardScope 处理；但没有 overlay 栈或 topmost overlay 仲裁时，多 overlay 同时打开的关闭顺序无法由框架保证。

## 验收

- 已输出 Dialog、Popup、Menu、Select、KeyboardScope 的打开时 focus、关闭时 restore、Escape 捕获规则表。
- 已确认 modal 与非 modal 的差异主要在 event path/pointer 边界：Dialog 使用 portal + stop boundary + shield，Popup/Select/DropdownMenu 为 portal 或本地 popup，不提供默认 focus trap。
- 已确认当前 Runtime 键盘默认行为不包含 Escape；Escape 只作为 keydown detail 暴露给 listener。
- 已标出 `MenuDefaultFocus`、focusout、restore focus、Escape close 属于配置/API 或应用层能力，不应误判为当前 overlay 默认行为。

## 后续依赖

- A9 Ref 命令生命周期审查应关注 DialogRef/PopupRef/SelectRef 打开关闭命令是否需要记录 focus restore 来源。
- A10.6 Overlay 控件族审查应决定 Dialog/Popup 是否要补 focus trap、Escape close、restore focus，以及这些行为是否可取消。
- A7/A5 后续测试补强可增加 modal Dialog 中 Tab 不越界、Escape 关闭只作用 topmost overlay、Select/DropdownMenu 关闭后 focus 回到 trigger/field 的用例。
