# A7.2 键盘事件和 shortcut 边界审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：键盘、焦点和可访问性审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "### A7\.|A7\.2|KeyboardScope|OnShortcut|shortcut" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "KeyboardScope|Shortcut|OnShortcut|keydown|keyup|KeyDown|KeyUp|global shortcut|RegisterShortcut|shortcut|key\." .`
  - `rg -n "RegisterShortcut|dispatchShortcut|Shortcut|keyboard|KeyDown|KeyUp|focusScope|focused" internal event ui widget examples/component_lab examples/event_system_testbench examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/keyboard.go`
  - `event/keyboard_test.go`
  - `event/input.go`
  - `internal/events.go`
  - `widget/keyboard_scope.go`
  - `ui/extended_types.go`
  - `system/global_shortcut.go`
  - `examples/event_system_testbench/main.go`
  - `examples/docs_browser/system_api_global_shortcut.go`
- 关联能力：
  - `keydown`、`keyup` typed event 路由
  - `KeyboardScope` Gio key event drain 边界
  - `OnShortcut` 局部快捷键注册、scope 匹配和排序
  - `PreventDefault`、passive、once、stop 对 shortcut/default action 的影响
  - system global shortcut 与组件树内 shortcut 的分层边界

## 事实结论

1. A7.2 的目标是审查 `keydown`、`keyup`、`OnShortcut`、global shortcut 和 `KeyboardScope`，输出局部快捷键和系统快捷键边界图，并确认局部快捷键只在 scope 内触发。证据：`docs/project-audit-task-breakdown.md:344`、`docs/project-audit-task-breakdown.md:348`、`docs/project-audit-task-breakdown.md:349`。
2. FluxUI 组件树内键盘事件类型由 `event.KeyDown`、`event.KeyUp` 暴露，实际类型来自 `internal.EventTypeKeyDown` 和 `internal.EventTypeKeyUp`。`KeyboardEvent` 包含 `Key`、`Code`、`Location`、`Repeat`、`Modifiers`、`IsComposing` 和 `Native`。证据：`event/keyboard.go:14`、`event/keyboard.go:15`、`internal/events.go:148`、`internal/events.go:157`。
3. `KeyboardEventFromGio` 将 Gio `key.Event` 映射为 FluxUI `KeyboardEvent`：`key.Release` 映射为 `keyup`，其他状态映射为 `keydown`；`key.NameReturn/key.NameEnter` 归一为 `Enter`，`key.NameSpace` 归一为 `Space`，`key.NameEscape` 归一为 `Escape`，`key.NameTab` 归一为 `Tab`。证据：`event/keyboard.go:279`、`event/keyboard.go:281`、`event/keyboard.go:299`、`event/keyboard.go:307`。
4. Gio modifiers 到 FluxUI modifiers 的映射由 `ModifiersFromGio` 负责：`Ctrl`、`Shift`、`Alt` 分别映射 Gio 对应 modifier，`Meta` 来自 `ModCommand` 或 `ModSuper`，`Shortcut` 来自 Gio `ModShortcut`。证据：`event/input.go:341`、`event/input.go:344`、`event/input.go:348`。
5. `OnKeyDown` 和 `OnKeyUp` 只是 `OnKeyboard` 的 typed helper；`OnKeyboard` 仍注册到通用 event target/listener 系统，所以 keydown/keyup 的 capture、target、bubble、priority、once、passive、stop 语义沿用 A5.2 的 EventTarget 规则。证据：`event/keyboard.go:197`、`event/keyboard.go:212`、`event/keyboard.go:217`、`internal/events.go:637`。
6. `DispatchKeyboardEvent` 的目标选择顺序是：显式 target 优先；未传 target 时用 runtime 当前 `FocusedTarget()`；仍为空时退到 root。随后先补全 keyboard defaults 字段和 key repeat 状态，再派发通用事件。证据：`event/keyboard.go:263`、`event/keyboard.go:273`、`internal/events.go:607`、`internal/events.go:616`、`internal/events.go:623`。
7. runtime 只在 `keydown` 且前置事件分发未被 `PreventDefault` 取消时运行 `dispatchShortcuts`；`keyup` 不触发 shortcut。证据：`internal/events.go:625`、`internal/events.go:626`、`internal/events.go:627`、`internal/events.go:632`。
8. runtime 的 key repeat 由跨帧 `keyDown` map 判断；`keydown` 记录 `Code` 优先、否则 `Key`，再次收到相同 identity 时设置 `Repeat=true`；`keyup` 删除该 identity。该 map 在 `beginEventFrame` 中只初始化，不清空，因此按键状态跨帧保留。证据：`internal/events.go:285`、`internal/events.go:889`、`internal/events.go:903`、`internal/events.go:906`、`internal/events.go:909`。
9. 局部快捷键的公开描述类型是 `Shortcut{Key, Code, Modifiers, ExactModifiers, Scope}`；`ShortcutKey` 按浏览器风格 key 匹配，`ShortcutCode` 按 code 匹配，`ShortcutExactModifiers` 要求没有额外 modifier，`ShortcutScope` 可绑定显式 scope。证据：`internal/events.go:174`、`event/keyboard.go:226`、`event/keyboard.go:231`、`event/keyboard.go:236`、`event/keyboard.go:242`。
10. `OnShortcut` 将 listener 注册到当前 `ctx.PathID()`；若 `spec.Scope` 为空，运行时把 listener target 作为 scope；若 `spec.Scope` 非空，运行时使用该显式 scope。注册时也会把当前 context 注册为 event target。证据：`event/keyboard.go:247`、`event/keyboard.go:250`、`internal/events.go:530`、`internal/events.go:541`、`internal/events.go:543`。
11. `dispatchShortcuts` 用当前 keyboard target 的 `eventPath(target)` 判断 scope：只有当 scope 存在于 focused target 的 path 内，且 shortcut spec 匹配当前 key event 时才会触发。该规则是“局部快捷键只在 scope 内触发”的权威边界。证据：`internal/events.go:934`、`internal/events.go:941`、`internal/events.go:947`、`internal/events.go:948`。
12. 多个 shortcut 命中时，排序规则是更靠近 target 的 scope 优先；同深度按 listener priority 降序；再按本帧注册序列升序。执行中尊重 `StopImmediatePropagation`、`StopPropagation`、`Passive` 和 `Once`。证据：`internal/events.go:957`、`internal/events.go:960`、`internal/events.go:963`、`internal/events.go:965`、`internal/events.go:973`、`internal/events.go:980`。
13. shortcut 匹配规则是：`Key` 非空时必须等于 event key，`Code` 非空时必须等于 event code；event modifiers 必须包含 spec modifiers；启用 `ExactModifiers` 时 event modifiers 必须与 spec 完全相等。证据：`internal/events.go:994`、`internal/events.go:998`、`internal/events.go:1001`、`internal/events.go:1004`、`internal/events.go:1007`。
14. `KeyboardScope` 是组件层入口：`KeyOnDown`、`KeyOnUp` 注册 key listener，`ShortcutOn` 注册局部 shortcut。`KeyboardScopeElement` 和 `ui.KeyboardScope` 只是把 widget 层能力暴露给 Element/Widget 两套 API。证据：`widget/keyboard_scope.go:81`、`widget/keyboard_scope.go:91`、`widget/keyboard_scope.go:95`、`widget/keyboard_scope.go:125`、`ui/extended_types.go:438`、`ui/extended_types.go:1092`。
15. `KeyboardScopeDisabled(true)` 会直接布局 child 并跳过 listener、shortcut、focus target 注册；因此 disabled scope 不会截获 keydown/keyup，也不会触发局部 shortcut。证据：`widget/keyboard_scope.go:55`、`widget/keyboard_scope.go:139`、`widget/keyboard_scope.go:140`。
16. `KeyboardScope` 的 Gio key event drain 发生在 child layout 之后，且只在 runtime 当前 focused target 非空并且 focused target 的 event path 包含该 scope 时执行。证据：`widget/keyboard_scope.go:156`、`widget/keyboard_scope.go:157`、`widget/keyboard_scope.go:190`、`widget/keyboard_scope.go:191`。
17. `processKeyboardScopeEvents` 当前使用 `ctx.Gtx.Event(key.Filter{})` 读取 Gio key events，没有显式 Gio focus tag filter；实际组件树边界依赖 runtime `FocusedTarget()` 与 `EventPathContains`，不是 Gio router 的 tag 过滤。证据：`widget/keyboard_scope.go:194`、`widget/keyboard_scope.go:195`、`widget/keyboard_scope.go:199`、`widget/keyboard_scope.go:201`。
18. runtime keyboard default action 在 shortcut 之后执行：`Tab` 移动 focus，`Shift+Tab` 反向移动，`Enter`/`Space` 激活当前 focus target。只要 keydown 或 shortcut handler 调用 `PreventDefault`，这些默认行为就被跳过。证据：`internal/events.go:626`、`internal/events.go:628`、`internal/events.go:1040`、`internal/events.go:1044`、`internal/events.go:1051`。
19. 现有单元测试已覆盖 keydown 可取消 Enter 激活和 Tab 移焦；也覆盖局部 shortcut 只在 focused target 所属 scope 内触发，切换 focus 后从 scope A 转到 scope B。证据：`event/keyboard_test.go:59`、`event/keyboard_test.go:76`、`event/keyboard_test.go:97`、`event/keyboard_test.go:115`、`event/keyboard_test.go:145`、`event/keyboard_test.go:158`。
20. event system testbench 提供手动入口：`KeyboardScopeFocusable(true)` + `KeyboardScopeAutoFocus(true)`，记录 keydown/keyup、Escape stop propagation、Tab preventDefault 和 `Ctrl+K` 局部 shortcut。证据：`examples/event_system_testbench/main.go:431`、`examples/event_system_testbench/main.go:445`、`examples/event_system_testbench/main.go:456`、`examples/event_system_testbench/main.go:459`。
21. system global shortcut 不进入组件树事件系统。`system.RegisterGlobalShortcut` 使用 `context.Context`、`GlobalShortcutSpec` 和 platform driver 注册系统级快捷键，返回 `GlobalShortcut` handle 和事件 channel；它不依赖 `internal.Context`、runtime focus target 或 `KeyboardScope`。证据：`system/global_shortcut.go:19`、`system/global_shortcut.go:33`、`system/global_shortcut.go:47`、`system/global_shortcut.go:63`、`system/global_shortcut.go:78`。
22. docs browser 的 global shortcut 示例通过 `system.RegisterGlobalShortcut` 注册 `Ctrl+Alt+Shift+F12`，回调中请求窗口显示和聚焦，并写入示例日志；这属于 OS/system 能力演示，不是 `OnShortcut` 或 `KeyboardScope` 的局部快捷键。证据：`examples/docs_browser/system_api_global_shortcut.go:20`、`examples/docs_browser/system_api_global_shortcut.go:47`、`examples/docs_browser/system_api_global_shortcut.go:53`、`examples/docs_browser/system_api_global_shortcut.go:55`。

### 局部快捷键和系统快捷键边界图

```text
Gio key.Event
  -> widget.KeyboardScope.processKeyboardScopeEvents
     gate: runtime FocusedTarget 在该 KeyboardScope 的 event path 内
  -> event.KeyboardEventFromGio
  -> runtime.DispatchKeyboardEvent(target = focused target)
     1. keydown/keyup capture -> target -> bubble listeners
     2. keydown 且未 PreventDefault: dispatchShortcuts
        gate: shortcut scope 在 focused target 的 event path 内
     3. keydown 且仍未 PreventDefault: runtime keyboard default action

system.RegisterGlobalShortcut
  -> platform driver / OS hotkey
  -> callback 或 GlobalShortcut.Events()
  -> 不进入 runtime event path、不依赖 FocusedTarget、不受 KeyboardScope 限制
```

### shortcut 边界矩阵

| 对象 | 触发来源 | 作用域判定 | keyup 行为 | cancel/stop 行为 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `OnKeyDown` | `DispatchKeyboardEvent` 通用事件分发 | 目标为 focused target 或显式 target 的 event path | 不适用 | `PreventDefault` 会阻止后续 shortcut/default action；`StopPropagation` 只影响传播 | 组件树内 keydown listener |
| `OnKeyUp` | `DispatchKeyboardEvent` 通用事件分发 | 同 keydown | 只派发 keyup，不触发 shortcut/default action | 可传播控制，但不触发 runtime keyboard default | 组件树内 keyup listener |
| `OnShortcut` | keydown 后置派生 | `spec.Scope` 或 listener target 必须在 focused target path 内 | 不触发 | 支持 passive、once、stop；`PreventDefault` 可阻止后续 default action | 局部快捷键权威入口 |
| `KeyboardScope` | Gio key event drain | `FocusedTarget` 必须位于 scope 内 | 转为 keyup listener，不派生 shortcut | disabled scope 跳过全部注册和 drain | 组件层局部键盘作用域 |
| `system.RegisterGlobalShortcut` | OS/platform driver | 无组件树 scope；由系统快捷键占用和平台权限决定 | 由系统 driver 决定事件模型 | 不使用 FluxUI `PreventDefault`/propagation | 系统级快捷键，和局部 shortcut 分层 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| Gio key drain 没有显式 focus tag filter | 中 | `processKeyboardScopeEvents` 使用空 `key.Filter{}`，组件树边界靠 runtime focus path gate。若 Gio router 中同一帧存在其他消费者或 editor focus，事件归属需要继续验证。 | A7.4、A7.5、A13 |
| Input Gio editor focus 与 runtime focus 分离 | 高 | A7.1 已确认 `Input` 不注册 runtime focus target；因此 `KeyboardScope` 的局部 shortcut 不一定覆盖正在 Gio editor 中编辑的文本输入，文本编辑键和局部 shortcut 的吞吐边界仍需专审。 | A7.4、A7.5、A13 |
| shortcut 在 keydown listener 之后运行 | 中 | 这是当前设计事实，允许 keydown listener 先 `PreventDefault`。但如果业务期望 shortcut 比普通 keydown 更早，需要明确优先级和文档语义。 | A7.3、A12 |
| `ShortcutKey` 匹配大小写敏感 | 中 | `ShortcutKey("k")` 与 Gio 字母 key 归一化之间依赖实际 `KeyboardEventFromGio` 输出；已有 testbench 用 `Ctrl+K` 手动验收，但自动测试里使用合成 event。 | A7.2 回归、A12 |
| global shortcut 与 local shortcut 命名相近 | 低 | `system.RegisterGlobalShortcut` 和 `ui.ShortcutOn` 分属不同包和不同事件域，但文档/API 使用者可能误以为二者共享 focus/scope/cancel 语义。 | Docs/API 指南 |
| key repeat map 跨帧保留但依赖 keyup 清理 | 低 | 如果平台缺失 key release 或窗口失焦时未补发 release，`Repeat` 可能保留为 true；Gio 注释说明 release 支持平台有限。 | A7.3、A12 |

## 验收

- 已记录 `keydown`/`keyup` 从 Gio 到 FluxUI typed event 的映射入口和 runtime dispatch 顺序。
- 已确认局部 shortcut 只在 `keydown` 且未被前置 listener 取消时运行，`keyup` 不派生 shortcut。
- 已确认局部 shortcut 的 scope 判定基于 focused target 的 event path；默认 scope 是 listener target，显式 scope 来自 `ShortcutScope`。
- 已记录 shortcut 多命中排序、once/passive/stop/preventDefault 的执行边界。
- 已明确 system global shortcut 是 `system` 包 OS-level 能力，不进入 runtime event path，不受 `KeyboardScope` 限制。
- 已标出 `Input` Gio editor focus、Gio key filter、大小写匹配和 release/repeat 的后续风险。

## 后续依赖

- A7.3 键盘默认行为审查应继续验证 shortcut、Tab 移焦、Enter/Space 激活和 `PreventDefault` 的最终优先级。
- A7.4 文本输入事件审查应处理 Input editor key event、beforeinput/input 和局部 shortcut 的冲突边界。
- A7.5 IME 和 submit 审查应确认 composing 状态下 shortcut 是否应被抑制或降级。
- A8 overlay 审查应验证 Dialog/Menu/Select portal 内的 KeyboardScope 与 focus path 是否满足局部 shortcut 边界。
- A12 测试审查应补齐真实 Gio router key event、`ShortcutKey("k")` 大小写、release 缺失/repeat 清理和 Input 内局部 shortcut 的自动测试。
