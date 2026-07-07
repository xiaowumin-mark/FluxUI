# A11.3 event_system_testbench 审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A11.3 event_system_testbench 审查。

## 事实结论

审查范围覆盖 `examples/event_system_testbench` 的 `README.md`、`FIX_ROADMAP.md` 和 `main.go`。该 example 是独立于 docs browser 的事件系统全量手工测试台，启动入口为 `go run ./examples/event_system_testbench`，窗口标题为 `FluxUI 事件系统全量测试台`，并在 `main` 中开启 `ui.EnablePerfDiagnostics(true)` 与 `ui.LogEvents(true)`。因此它同时提供 GUI 内业务日志和控制台底层事件诊断，适合作为 Event 相关回归的手工入口。

| 测试项 | 代码入口 | 主要覆盖 | 操作说明 | 预期结果 | 当前问题/复验点 |
| --- | --- | --- | --- | --- | --- |
| P0 兼容 API 和 escape hatch | `compatPanel` | 旧 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`、`ctx.Gtx.Event` 说明 | 点击旧按钮、悬停按钮、编辑输入框、滚动 P0 内部滚动区，观察日志 | 日志分别出现旧 click、hover true/false、输入值变化和真实 `x/y` offset | `FIX_ROADMAP.md` 已记录 R1 修复 `ScrollOnChange` 长期为 0、单行 Input/DragSource 截停父滚动；仍需在测试台手工复验 |
| P1 EventTarget 和基础分发 | `dispatchPanel` | `TargetID`、事件路径、capture/target/bubble、`PreventDefault`、`StopPropagation`、`StopImmediatePropagation`、`Once`、`Passive`、synthetic dispatch | 切换 PreventDefault、StopPropagation 开关，点击合成事件按钮，并对比 GUI 日志顺序与控制台事件诊断 | 正常路径应体现父 capture、目标 listener、父 bubble；PreventDefault 后 `defaultAllowed=false`；StopPropagation 后父 bubble 不继续；Once 在同次双派发只出现一次；Passive 调 `PreventDefault` 返回 false | 现有界面仍偏 API 验证，`FIX_ROADMAP.md` 的 R6 标记为表达改进项：需要更可视化地展示三层事件路径和每个按钮的预期/实际 |
| P2 Pointer 和 Wheel | `pointerWheelPanel` | pointer down/up/move/enter/leave/over/out/cancel、click/dblclick/auxclick/contextmenu、wheel、pointer capture、coalesced move | 在蓝色区域内移动、拖动、双击、右键、中键、滚轮；观察蓝点位置、日志坐标、capture 状态和 coalesced 数量 | pointer 事件只发生在视觉面板区域；拖动期间 capture 生效；右键 contextmenu 可阻止默认；wheel 日志显示 `dx/dy`、修饰键和取消状态 | `FIX_ROADMAP.md` 已记录 R2 修复 PointerArea 命中区过大；仍需复验同一行空白区域不会抢事件或阻断滚动 |
| P3 Focus、Keyboard 和局部快捷键 | `keyboardFocusPanel` | `FocusManager`、focus/blur/focusin/focusout、keydown/keyup、repeat、key/code/modifiers、局部 `ShortcutOn`、Enter/Space/Escape/Tab 默认行为 | 点击焦点块，请求 forward/backward/blur，按 Tab/Shift+Tab、Ctrl+K、Escape、Enter、Space，并打开/关闭阻止 Tab 默认开关 | 点击后 GUI 日志出现 focus/focusin；键盘事件进入局部 scope；Ctrl+K 只在 scope 内触发并阻止默认；Escape 停止传播；阻止 Tab 默认时焦点不移动；Enter/Space 激活当前焦点块 | `FIX_ROADMAP.md` 已记录 R3 修复 focus/keyboard 桥接；仍需复验首帧/下一帧 focus 同步、disabled/hidden 跳过和局部快捷键作用域 |
| P4 Text Input、IME 和编辑事件 | `textInputPanel` | `beforeinput`、`input`、`change`、`submit`、`InputRef.SetText/Append/Clear/Focus/Blur`、composition synthetic 入口 | 在数字输入框输入普通字符和数字，按 Enter，使用程序化按钮，触发 synthetic composition start/update/end | 非数字输入可在 beforeinput best-effort 阶段被拦截；数字输入触发 input/change；Enter 触发 submit；程序化命令标记 source；composition 日志按 start/update/end 出现 | 历史反馈认为 P4 基本正常；真实 IME 仍依赖 P3 的键盘焦点桥接稳定性 |
| P5 组件默认行为统一迁移 | `defaultBehaviorPanel` | Button、Pressable、ClickArea、Checkbox、Switch、Radio、Select、ScrollView、Slider、Dialog、Popup 的默认行为和可取消性 | 先关闭阻止开关逐项点击/滚动/拖动；再打开阻止 click 默认或阻止 wheel 默认重复操作；打开 Dialog/Popup 并点击内部按钮、空白和遮罩 | 未阻止时旧回调或状态变化执行；父 capture 调 `PreventDefault` 后对应 click/wheel 默认行为不执行；Dialog/Popup 打开不卡死，内部点击不误关闭，遮罩仍按规则关闭 | `FIX_ROADMAP.md` 已记录 R4/R5 修复默认行为取消、Dialog/Popup 卡死和 Select/Menu 滚动后命中刷新；仍需完整手工复验 |
| P5/P6 DragSource、DropTarget 和 drag* 事件流 | `dragDropPanel` | 真实拖放组件、typed `DragEvent`、`dragover/drop` 默认行为、payload、operation | 拖动 DragSource 到 DropTarget，观察 active 状态和 drop 日志；点击 typed dragover 派发按钮 | DropTarget active/类型/错误日志准确；drop 后 payload 进入日志；typed dragover 可被 `PreventDefault`，dispatch 返回值反映默认行为是否允许 | 真实外部拖放依赖 Gio/系统后端；可先以内置 DragSource 到 DropTarget 做 smoke |
| P6 自定义事件、Portal 和边界策略 | `architecturePanel` | `Detail` payload、指定 target 派发、普通冒泡、stop boundary、redirect boundary、portal owner path、activation event | 点击 owner target、自定义事件、activation、boundary/portal 相关按钮，观察 owner capture/bubble 与 target listener 日志 | 自定义 detail 能到指定 target；普通事件按 owner 路径冒泡；stop boundary 截断；redirect/portal 按 owner path 改写；activation 给出 source/keyEquivalent | P6 受 P1 分发路径和 P5 overlay/portal 稳定性影响，修复相关底层后需要回归 |
| P7 性能、诊断和文档教学入口 | `diagnosticsPanel`、`logPanel` | `LogEvents`、事件路径、listener 耗时、取消状态、默认行为结果、高频 pointer/wheel 观察入口 | 在 P1/P2/P5 区域制造事件，再点击 P7 手动诊断标记，同时观察 GUI 日志和控制台输出 | GUI 日志显示业务侧结果；控制台输出事件 type、path、phase、listener 耗时、cancel/default 结果；高频 pointer/wheel 可用于观察 coalesced move 和诊断噪声 | GUI 诊断仍偏简略，`FIX_ROADMAP.md` 建议后续增加 event type、target、path、phase、defaultPrevented 的可视化诊断视图 |

### P1/P3/P5 教学口径

| 重点项 | 使用者应理解的测试目的 | 最小可执行步骤 | 判定方式 |
| --- | --- | --- | --- |
| P1 | 验证同一个事件如何从父级 capture 进入目标，再按 bubble 返回；同时验证 stop、once、passive、preventDefault 的语义差异 | 先无开关点击 synthetic dispatch，再分别打开 PreventDefault 和 StopPropagation 重复点击 | 看 GUI 日志顺序和 `defaultAllowed`；PreventDefault 只取消默认行为，不等同于停止传播；StopPropagation 停止后续路径；Passive 不能成功取消默认 |
| P3 | 验证组件树内局部键盘 scope 是否拿到真实键盘事件，并且默认键盘行为可被取消 | 点击焦点块，按 Ctrl+K、Tab、Escape、Enter/Space，再打开阻止 Tab 默认重复按 Tab | 看 focus/key 日志、当前焦点变化和 activation 日志；局部快捷键不应离开 scope 后仍触发 |
| P5 | 验证组件旧回调和状态变化是否统一挂在可取消的新事件默认行为之后 | 先点击每类控件确认能变更；打开阻止 click/wheel 后重复点击、滚动和拖动 | 阻止后不应执行旧 `OnClick/OnChange` 或滚动默认行为；overlay 内部点击不应被父级 outside/mask 规则误关 |

## 风险

- 本轮审查只建立测试台说明和验收基线，没有执行真实 GUI 手工点击；P0-P7 的交互结论仍需要人工运行窗口确认。
- `FIX_ROADMAP.md` 同时包含历史问题清单和 R1-R5 修复记录；本文件把 P0/P2/P3/P5 的旧失败归类为“已修复但需复验”，避免把历史问题误判为当前必现失败。
- P1 当前仍有可理解性风险：测试项把多个概念压在同一区域内，使用者需要对照日志才能理解 capture、bubble、stop、preventDefault 的区别。
- P7 诊断依赖控制台输出较多，GUI 只显示业务日志；在没有终端可见性的环境下，listener 耗时、完整 path 和 cancel/default 细节不够直观。
- P4 的真实 IME/composition 仍受平台、Gio 后端和键盘 focus 桥接影响；当前 synthetic composition 不能完全代表真实输入法行为。
- Drag/drop 受系统 transfer 后端能力影响；某些环境下外部 payload 行为可能不可复现，只能先用内部 DragSource/DropTarget 做 smoke。

## 验收

- 已建立 P0-P7 每个测试项的代码入口、覆盖范围、操作说明、预期结果和当前问题/复验点。
- 已单独说明 P1、P3、P5 的测试目的、最小操作步骤和判定方式，满足“不看源码也能理解在测什么”的验收目标。
- 已把 `FIX_ROADMAP.md` 中 R1-R5 的历史问题与修复记录纳入复验口径，明确 P0/P2/P3/P5 不应按旧失败直接判定当前状态。
- 已确认 `event_system_testbench` 是 Event 相关手工 smoke 入口，并与 A5/A6/A7/A8/A10 的事件、滚动、键盘、overlay、控件族审查结果建立对应关系。
- 已确认本轮只记录 Docs/Test 与 Event 审查事实，不修改 runtime、widget 或 example 行为。

## 后续依赖

- A5.2/A5.5：事件分发顺序或 default action 可取消性变更后，需要回归 P1 和 P5。
- A6.2/A6.4：wheel 分发、父子滚动、滚动后命中刷新变更后，需要回归 P0、P2、P5 和 P7 高频 wheel 观察。
- A7.1/A7.2/A7.3/A7.5：focus target、KeyboardScope、键盘默认行为、IME submit 变更后，需要回归 P3 和 P4。
- A8.2/A8.3/A8.4：portal event path、outside click、overlay focus/Escape 变更后，需要回归 P5 和 P6。
- A10.1/A10.2/A10.3/A10.4/A10.5/A10.6/A10.7：对应控件族修复后，需要在 P5 或 P5/P6 drag/drop 区域补做手工 smoke。
- A12.x：建议将 P1 分发顺序、P3 键盘 scope、P5 PreventDefault gate、P5 overlay 内部点击、P7 GUI 诊断视图拆成可重复的自动或半自动 smoke。
