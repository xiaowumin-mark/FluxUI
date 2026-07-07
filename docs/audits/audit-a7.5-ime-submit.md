# A7.5 IME 和 submit 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。
- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A7\.5|IME|composition|submit|SubmitEvent|compositionstart|compositionupdate|compositionend" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "compositionstart|compositionupdate|compositionend|Composition|SubmitEvent|submit|Submit|IsComposing|beforeinput|InputRef|TextField|SearchBar" event internal ui widget examples -g "*.go"`
  - `go test ./event ./internal ./widget ./examples/event_system_testbench ./examples/docs_browser`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `event/text.go`
  - `event/text_test.go`
  - `event/keyboard.go`
  - `internal/events.go`
  - `widget/input.go`
  - `widget/input_event_test.go`
  - `widget/keyboard_scope.go`
  - `examples/event_system_testbench/main.go`
  - `gioui.org@v0.9.0/widget/editor.go`
- 关联能力：
  - compositionstart/update/end typed event
  - TextField Enter submit
  - InputEvent.IsComposing / KeyboardEvent.IsComposing 边界
  - Gio Editor ChangeEvent / SubmitEvent 桥接
  - 组合输入期间 submit 提前触发风险识别

## 事实结论

1. A7.5 的目标是审查 `compositionstart/update/end` 与 Enter submit，并输出 IME 组合期间的事件顺序；验收重点是组合中不会提前提交，组合结束后的行为明确。证据：`docs/project-audit-task-breakdown.md:368`、`docs/project-audit-task-breakdown.md:371`、`docs/project-audit-task-breakdown.md:373`。
2. FluxUI 事件层已经定义 `compositionstart`、`compositionupdate`、`compositionend` 三段事件，并提供 `CompositionEvent`、`OnComposition`、`DispatchCompositionEvent`。composition 默认 `Bubbles=true`、`Cancelable=false`，默认类型为空时补为 `compositionupdate`。证据：`internal/events.go:24`、`event/text.go:10`、`event/text.go:61`、`event/text.go:115`、`event/text.go:137`。
3. composition 目前是事件层 synthetic 入口，不是 TextField 真实 IME 生命周期自动桥接。`widget/input.go` 的 editor loop 只消费 Gio `ChangeEvent` 和 `SubmitEvent`，没有处理 composition start/update/end。证据：`widget/input.go:427`、`widget/input.go:432`、`widget/input.go:437`。
4. Gio `widget.Editor` 在 v0.9.0 中有内部 `ime` 状态，但公开给调用侧的 editor event 仍以 `ChangeEvent` 和 `SubmitEvent` 为主；当前 FluxUI 没有读取 Gio 内部 IME 状态的公开入口。证据：`C:\Users\xiaow\go\pkg\mod\gioui.org@v0.9.0\widget\editor.go:76`、`C:\Users\xiaow\go\pkg\mod\gioui.org@v0.9.0\widget\editor.go:169`、`C:\Users\xiaow\go\pkg\mod\gioui.org@v0.9.0\widget\editor.go:173`、`C:\Users\xiaow\go\pkg\mod\gioui.org@v0.9.0\widget\editor.go:620`。
5. TextField 只有在配置了 `InputOnSubmit` 时才设置 `editor.Submit=true`；Gio 文档语义是把 carriage return 转成 `SubmitEvent`。证据：`widget/input.go:400`、`C:\Users\xiaow\go\pkg\mod\gioui.org@v0.9.0\widget\editor.go:53`。
6. TextField 的 submit 事件由 Gio `SubmitEvent` 直接进入 `dispatchSubmit`，生成 `InputType=insertLineBreak`、`Source=user`、`Trusted=true`、`BestEffort=false` 的 `InputEvent`。证据：`widget/input.go:437`、`widget/input.go:616`、`widget/input.go:621`。
7. `newInputEvent` 当前没有设置 `InputEvent.IsComposing`；因此用户编辑、programmatic 输入和 submit 在当前 TextField 路径里都会保持 `IsComposing=false` 的零值。证据：`event/text.go:47`、`widget/input.go:625`、`widget/input_event_test.go:193`。
8. `TestInputSubmitEvent` 只覆盖非 composition 场景，并断言 submit 的 `IsComposing=false`；没有真实 IME 组合期间按 Enter 的测试。证据：`widget/input_event_test.go:143`、`widget/input_event_test.go:190`、`widget/input_event_test.go:193`。
9. 普通 keyboard event 同样有 `IsComposing` 字段，但 `KeyboardEventFromGio` 未从 Gio key event 填充该字段；Enter/Return 被映射为 `Key="Enter"`。证据：`internal/events.go:156`、`event/keyboard.go:278`、`event/keyboard.go:301`。
10. runtime keyboard default action 对 `Enter`、换行、`Space` 会激活当前 focus target，且当前默认行为没有检查 `KeyboardEvent.IsComposing`。这条路径属于焦点控件激活，不是 TextField 的 Gio `SubmitEvent`，但同样受“组合中 Enter”语义影响。证据：`internal/events.go:629`、`internal/events.go:1040`、`internal/events.go:1051`。
11. event system testbench 的 P4 区域提供 composition synthetic 演示：手动依次派发 `compositionstart -> compositionupdate -> compositionend`，并明确说明真实 IME 生命周期取决于 Gio 后端。证据：`examples/event_system_testbench/main.go:85`、`examples/event_system_testbench/main.go:533`、`examples/event_system_testbench/main.go:555`、`examples/event_system_testbench/main.go:561`。
12. 当前可确认的真实 TextField 顺序为：普通用户编辑经 `ChangeEvent` 后触发 `beforeinput -> input -> change -> InputOnChange`；Enter submit 经 `SubmitEvent` 后触发 `submit`。composition synthetic 顺序可由调用方派发，但不会自动包围真实输入事件。证据：`widget/input.go:588`、`widget/input.go:594`、`widget/input.go:603`、`widget/input.go:607`、`widget/input.go:609`、`widget/input.go:616`。

### IME / submit 顺序矩阵

| 场景 | 入口 | 当前事件顺序 | IsComposing | submit 行为 | 结论 |
| --- | --- | --- | --- | --- | --- |
| synthetic composition 演示 | `DispatchCompositionEvent` | `compositionstart -> compositionupdate -> compositionend` | 不适用 `InputEvent` | 不触发 submit | 事件层正式 API，但不代表真实 IME |
| 普通用户文本编辑 | Gio `ChangeEvent` | `beforeinput -> input -> change -> InputOnChange` | false | 不触发 submit | 已覆盖基线 |
| 非组合 Enter submit | Gio `SubmitEvent` | `submit` | false | 触发 `InputOnSubmit` | 已有测试覆盖 |
| 真实 IME 组合中 Enter | Gio 后端/editor 内部状态 | 未桥接 composition 生命周期 | 当前无法置 true | 无 FluxUI 级 gate | 风险：无法从 FluxUI 层证明不会提前 submit |
| 组合结束后的提交 | Gio `ChangeEvent`/`SubmitEvent` 实际输出 | 取决于 Gio 后端输出 | 当前仍为 false | 若 Gio 输出 SubmitEvent 则提交 | 行为需要后端能力或手动测试确认 |
| runtime focused Enter 默认行为 | `KeyboardScope` Gio key event | `keydown` 后 default activation | 当前未填充 | 激活 focus target | 与 TextField submit 分层，缺少 composing gate |

## 风险

1. 当前 `InputEvent.IsComposing` 和 `KeyboardEvent.IsComposing` 在真实 Gio 输入路径中都没有被填充，后续如果把字段存在误读为“真实 IME 已桥接”，会高估可访问性和输入法兼容能力。
2. TextField 无法在 FluxUI 层识别“IME 组合中 Enter”，因此 A7.5 验收项“组合中不会提前提交”只能标为未完全满足；是否提前提交取决于 Gio 后端是否在组合阶段抑制 `SubmitEvent`。
3. runtime keyboard default action 对 Enter 没有 composing gate。若某个输入法后端同时产生 keydown Enter 与组合确认语义，非 TextField 的 focus activation 可能被提前触发。
4. synthetic composition API 可用于测试事件分发链路，但不能证明真实 IME 的 start/update/end 与 beforeinput/input/change/submit 顺序。
5. `submit` 虽然是 cancelable，但当前取消只影响事件分发返回值，没有额外默认行为可阻止；后续如引入表单提交、焦点移动或自动关闭菜单，需要重新绑定 `PreventDefault` 语义。

## 验收

- 已输出 IME composition 与 submit 的当前顺序矩阵。
- 已明确 composition 三段事件是正式事件层 API，但真实 Gio IME 生命周期尚未自动桥接到 TextField。
- 已明确当前 TextField submit 的非组合路径：`SubmitEvent -> submit`，`IsComposing=false`，已有测试覆盖。
- 已标出 A7.5 验收缺口：FluxUI 层目前无法证明“组合中不会提前提交”，只能依赖 Gio 后端行为或后续新增桥接。
- 已区分 TextField 的 Gio `SubmitEvent` 路径和 runtime keyboard default activation 路径，避免把两者混为一个 submit 行为。

## 后续依赖

- 若后续需要满足完整 IME 验收，应优先确认 Gio 是否暴露 composition 生命周期或 composing 状态；若无公开 API，需要评估平台后端事件或 Gio 升级。
- A7 后续可访问性任务若审查 Enter 默认激活，需要复用本文件的 composing gate 缺口，避免在组合输入中误触发按钮、选择器或表单动作。
- 自动测试建议补齐真实或模拟 IME 场景：composition active 时 Enter 不触发 `InputOnSubmit`，composition end 后的文本变更与 submit 顺序明确。
- event system testbench 可保留 synthetic composition 入口作为事件层验收，但文档应持续标注它不是后端 IME smoke 测试。
