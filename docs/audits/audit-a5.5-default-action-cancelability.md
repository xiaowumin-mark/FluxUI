# A5.5 default action 可取消性矩阵
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：事件系统审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A5\\.5|default action|PreventDefault|beforeinput|dragover|drop|wheel|keydown|click" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "PreventDefault|DefaultPrevented|Cancelable|cancelable|default action|beforeinput|dragover|drop|wheel|keydown|KeyDown|Click|OnClick|OnChange|DragOver|Drop" event internal widget examples -g "*.go"`
  - `rg -n --glob "*_test.go" "PreventDefault|DefaultPrevented|DispatchClick|beforeinput|BeforeInput|wheel|Wheel|dragover|DragOver|drop|Drop|KeyDown|Tab|Submit|passive|Passive" event internal widget examples`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/events.go`
  - `event/event.go`
  - `event/dispatcher.go`
  - `event/input.go`
  - `event/keyboard.go`
  - `event/text.go`
  - `event/drag.go`
  - `widget/event_defaults.go`
  - `widget/button.go`
  - `widget/click_area.go`
  - `widget/checkbox.go`
  - `widget/input.go`
  - `widget/list_grid.go`
  - `widget/drop_target.go`
  - `widget/drag_source.go`
  - `event/event_test.go`
  - `event/keyboard_test.go`
  - `event/text_test.go`
  - `widget/input_event_test.go`
  - `widget/list_grid_test.go`
  - `widget/drop_target_test.go`
  - `widget/drag_drop_integration_test.go`
- 关联能力：
  - `PreventDefault` 生效条件
  - runtime keyboard default action 可取消边界
  - widget-local click/wheel/beforeinput/drop default action gate
  - passive listener 取消保护
  - 事件可取消标记与实际默认行为的分层

## default action 可取消性矩阵

| 事件 | 默认 cancelable | 默认行为或兼容行为 | `PreventDefault` 生效条件 | 不生效条件 | 当前覆盖 |
| --- | --- | --- | --- | --- | --- |
| `click` | 是 | 旧 `OnClick` 回调；选择类控件的状态切换 default action；Button/Pressable/ClickArea Ref 和键盘激活最终也走 click dispatch。 | listener 在非 passive 状态下调用；调用方必须通过 `event.Dispatcher.DispatchClickEvent` 或 `widget.dispatchClickDefault` 读取 allowed。 | disabled/loading Button 不进入 click；绕过统一 helper 直接调用旧回调时不会受控；passive listener 调用返回 false。 | `event.TestDispatchClickBridgesLegacyHandler` 覆盖旧回调 gate；A5.3 已记录桥接顺序。 |
| `keydown` | 是 | runtime 键盘默认行为：`Tab` 移动焦点；`Enter`、`Space` 激活当前 focus target；local shortcut 位于普通 dispatch 之后、keyboard default 之前。 | `DispatchKeyboardEvent` 分发普通 `keydown` 后，若事件仍 allowed，才执行 shortcut 与 `runKeyboardDefault`；shortcut listener 也可继续 `PreventDefault` 阻止后续 keyboard default。 | `keyup` 当前也被标记 cancelable，但 runtime default 只处理 `keydown`；没有 runtime/focus target 时不会产生实际默认激活；passive listener 不能取消。 | `event.TestKeyDownCanCancelActivationAndFocusMove` 覆盖 Enter/Tab 取消。 |
| `wheel` | 是 | `ScrollView` 本地默认滚动：`processWheelEvents` dispatch 后，allowed 才调用 `applyWheelDefault` 修改 offset 并请求 redraw；位置变化后可能触发旧 `ScrollOnChange`。 | wheel listener 在目标或祖先路径非 passive 调用；事件来自 `ScrollView` 的 wheel tag 且轴过滤匹配；`DispatchWheelEvent` 返回 allowed。 | `ScrollRef`、auto-to-end、scrollbar 拖拽和其他非 wheel 位置变化不是此 default action；水平/纵向轴不匹配时 Gio filter 不会把事件交给该 ScrollView；passive listener 不能取消。 | `widget/list_grid_test.go` 覆盖 wheel 滚动和嵌套轴边界；缺少专门断言 `wheel.PreventDefault()` 阻止 `applyWheelDefault` 的 widget 测试。 |
| `beforeinput` | 是 | Input mutation gate：程序命令在 SetText 前 dispatch；用户输入在 Gio editor 已变更后 best-effort dispatch，取消时回滚 editor 文本；被接受后才发 `input/change/InputOnChange`。 | `InputOnBeforeInput` 或 EventTarget listener 非 passive 调用；`inputWidget.dispatchInputEvent` 返回 false 后，程序输入直接返回，用户输入 `editor.SetText(previous)`。 | `input`、`change` 不可取消，不能用它们阻止旧 `InputOnChange`；用户输入取消是 best-effort 回滚，复杂 IME/composition 仍需 A7 继续审查。 | `event/text_test.go` 覆盖 beforeinput 返回值；`widget/input_event_test.go` 覆盖用户输入取消回滚且不触发旧 onChange。 |
| `dragover` | 是 | `DropTarget` active 进入 gate：Gio `transfer.InitiateEvent` 后先发 `dragenter`，再发 `dragover`，两者 allowed 才 `setActive(true)`。 | `OnDrag(..., DragOver, ...)` 或通用 listener 非 passive 调用；`DropTarget.dispatchDropTargetEvent` 读取 `DispatchDragEvent` allowed。 | `DropTargetDisabled(true)` 或无 handler 时不注册/不处理；只取消 `dragover` 不会撤销已发生的 Gio initiate，只阻止 FluxUI active 状态和后续 active callback。 | `event.TestDispatchDragEventDefaultsAndPreventDefault` 覆盖 dragover 可取消；缺少 DropTarget active gate 的专门测试。 |
| `drop` | 是 | `DropTarget` 的 `onDrop`/`onError` gate：Gio `transfer.DataEvent` 读出 payload 后先发 `drop`，allowed 才调用旧 drop/error 回调。 | `drop` listener 非 passive 调用；`dispatchDropTargetEvent` 返回 false 后跳过 `onDrop` 和 `onError`，随后清理 active。 | 数据读取已经在 dispatch 前发生，因此取消不能避免 `dropEventFromTransfer` 读取 payload；disabled/无 handler 时不进入；passive listener 不能取消。 | `event` 层覆盖 drag 默认；`widget/drag_drop_integration_test.go` 覆盖正常 drop，但缺少 `drop.PreventDefault()` 阻止 `onDrop/onError` 的 widget 测试。 |

## 事实结论

1. `internal.Event.PreventDefault()` 的唯一生效条件是 `event.Cancelable == true` 且当前 listener 不是 passive；成功后设置 `DefaultPrevented=true` 并返回 `true`。
2. passive listener 中调用 `PreventDefault()` 返回 `false`，不会设置 `DefaultPrevented`；这条规则同时适用于普通 listener 和 shortcut listener。
3. `Runtime.DispatchEvent` 不执行通用 default action，只返回 `!(Cancelable && DefaultPrevented)`；具体默认行为由 `DispatchKeyboardEvent` 或 widget 调用方在 dispatch 后读取 allowed。
4. `event.DispatchPointerEvent` 会按事件类型重写 pointer event 的 `Bubbles` 和 `Cancelable`；`pointerenter`、`pointerleave`、`pointercancel` 不可取消，其余 pointer 类型可取消。
5. `click` 是特殊语义事件，`event.Dispatcher.DispatchClickEvent` 和 `widget.dispatchClickDefault` 都通过 `PointerEvent{Type: click, Cancelable: true}` 分发，并以 allowed 作为旧回调或状态切换 gate。
6. `Button` disabled/loading 时不会处理 Ref click、pointer click 或 focus activation；此时不是 `PreventDefault` 生效，而是组件自身提前阻断默认入口。
7. `Pressable`/`ClickArea` 没有 disabled gate；其旧 `onClick` 主要受 `click` 是否被取消控制。
8. `Checkbox` 等选择类组件通过 `dispatchClickDefault(ctx, source, activate)` 先发 cancelable `click`，未取消时才执行状态切换和旧 `OnChange`。
9. `DispatchKeyboardEvent` 在普通 keydown dispatch allowed 后才执行 local shortcuts；shortcut listener 可以继续 `PreventDefault()`，再阻止 `Tab` focus move 或 `Enter/Space` activation。
10. `runKeyboardDefault` 只处理 `keydown`；虽然 `KeyboardEventFromGio` 也把 `keyup` 标记为 cancelable，但当前没有 `keyup` default action。
11. `DispatchWheelEvent` 总是把 wheel 标记为 bubbles/cancelable；`ScrollView.processWheelEvents` 只在 allowed 时调用 `applyWheelDefault`。
12. `ScrollView` 的 wheel default action 只覆盖 Gio pointer.Scroll 进入该 wheel tag 且轴过滤匹配的情况；其他滚动来源不在 `wheel.PreventDefault` 控制下。
13. `beforeinput` 和 `submit` 在 `event/text.go` 中标记为 cancelable；`input`、`change`、composition 事件不可取消。
14. `InputRef.SetText/Append/Clear` 的程序输入在 mutation 前发 `beforeinput`，取消后不执行 `editor.SetText`、`input`、`change` 或旧 `InputOnChange`。
15. Gio 用户编辑是先产生 editor 变更再由 FluxUI 发 `beforeinput`；取消时通过 `editor.SetText(previous)` 回滚，因此属于 best-effort cancellation。
16. `dispatchSubmit` 当前只分发 cancelable `submit` 并把回调注册为 listener；返回值没有继续 gate 其他默认 submit 行为，因此 `submit.PreventDefault()` 当前只影响 dispatch allowed/diagnostics，不阻止额外 widget 行为。
17. `event.DispatchDragEvent` 对除 `dragleave`、`dragend` 外的 drag event 默认标记 cancelable，因此 `dragover` 和 `drop` 都可取消。
18. `DropTarget` 对 `dragenter/dragover/drop` 都读取 allowed：`dragover` 可阻止 active 进入，`drop` 可阻止 `onDrop/onError`。
19. `drop` 的 payload 读取发生在 dispatch `drop` 之前；因此 `PreventDefault()` 不能避免读取成本或读取错误产生，只能阻止后续旧回调执行。
20. `DragSource` 的 `drag`/`dragstart` 也会读取 allowed，但 `transfer.OfferCmd` 在 `drag` dispatch 之前已执行；这属于相邻风险，不是本任务主输入的 `dragover/drop` 行为。

## 风险

| 风险 | 级别 | 说明 | 建议归属 |
| --- | --- | --- | --- |
| default action 没有集中 registry | M | 通用 `DispatchEvent` 只返回 allowed；click、wheel、beforeinput、drop 的默认行为分散在 widget/helper 中，新增组件可能忘记读取 allowed。 | A5 后续 event API、Widget review checklist |
| `wheel.PreventDefault()` 缺少 widget 级回归测试 | M | 现有测试覆盖 wheel 滚动和嵌套轴过滤，但没有直接验证取消后 `applyWheelDefault` 不执行。 | A6.2 wheel 分发和父子滚动审查 |
| `drop.PreventDefault()` 不能阻止 payload 读取 | M | `DropTarget` 在 dispatch `drop` 前已经调用 `dropEventFromTransfer` 读数据；大 payload 或错误仍会发生，只能取消 `onDrop/onError`。 | A11.4 拖放示例/API 审查、system drag/drop |
| `dragover/drop` widget gate 测试不足 | M | event 层已有 dragover preventDefault 测试，DropTarget active/onDrop gate 缺少专门测试。 | Widget drag/drop tests |
| `submit` 标记可取消但无实际 widget default action | S | 当前 `PreventDefault` 可改变 dispatch 返回值，但没有后续 submit 默认行为可阻止，文档需避免暗示它会阻止 `InputOnSubmit` 自身。 | A7.4 表单提交和验证审查 |
| `keyup` 标记可取消但没有 runtime default | S | 这不造成错误行为，但文档需说明 cancelable 不等于一定存在默认动作。 | Event API 文档 |
| Gio 用户输入取消是 best-effort 回滚 | M | editor 已经变更后才获得 `ChangeEvent`，复杂 IME/composition 场景仍需后续专门审查。 | A7 text input |

## 验收

- 已列出 `click`、`keydown`、`wheel`、`beforeinput`、`dragover`、`drop` 的 cancelable 规则。
- 已明确 `PreventDefault` 的通用生效条件：
  - event 必须 `Cancelable=true`；
  - 当前 listener 不能是 passive；
  - 调用方必须读取 dispatch 返回的 allowed 才能阻止具体 default action。
- 已明确 `PreventDefault` 不生效或不控制的条件：
  - passive listener；
  - 不可取消的 `input/change/composition/dragleave/dragend`；
  - disabled/loading 等组件入口提前阻断；
  - Ref/auto-to-end/scrollbar 等非 wheel 滚动来源；
  - `drop` dispatch 前已经发生的数据读取。
- 已能解释 default action 分层：
  - runtime：`keydown` shortcut、focus move、focus activation；
  - event helper：`click` 到旧 `OnClick`；
  - widget-local：选择类 click 状态切换、ScrollView wheel、Input beforeinput、DropTarget drop。
- 已标出测试覆盖：
  - event 层覆盖 passive、click、keydown、beforeinput、dragover；
  - widget 层覆盖 input beforeinput rollback、wheel 滚动、drag/drop 正常流程；
  - 缺口是 wheel/drop preventDefault 的 widget 级 gate 测试。

## 后续依赖

- A6.2 wheel 分发和父子滚动审查应补充 `wheel.PreventDefault()` 阻止 `applyWheelDefault` 的测试或手动验收入口。
- A7 text input 审查应复用本文的 `beforeinput` mutation gate，并继续验证 IME/composition 下的 best-effort rollback 风险。
- A7.4 表单提交和验证审查应决定 `submit` 是否需要真实 default action；若需要，必须读取 `DispatchInputEvent` allowed。
- A11.4 拖放示例/API 审查应补足 `dragover.PreventDefault()` 和 `drop.PreventDefault()` 对 active/onDrop/onError 的验收。
- Event API 文档应明确：`Cancelable=true` 只表示事件允许取消，是否存在可阻止的默认行为取决于调用方是否实现并读取 allowed。
