# A10.2 表单选择控件族审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Widget、Event
- 范围：Checkbox、Switch、RadioGroup、Select
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph explore "A10.2 Checkbox Switch RadioGroup Select click keyboard controlled value popup Ref default action cancelable"`
  - 局部源码行号抽取：`widget/checkbox.go`、`widget/switch.go`、`widget/selection.go`、`widget/component_refs.go`、`widget/event_defaults.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/selection.go`
  - `widget/component_refs.go`
  - `widget/event_defaults.go`
  - `event/keyboard.go`
  - `event/pointer.go`

## 事实结论

1. A10.2 的目标是审查表单选择控件族的 click default、keyboard、controlled value、`OnChange`、Ref 和 popup 边界，并明确每个选择控件默认行为的可取消性。范围为 Checkbox、Switch、RadioGroup、Select。证据：`docs/project-audit-roadmap.md:307`、`docs/project-audit-task-breakdown.md:444`、`docs/project-audit-task-breakdown.md:448`、`docs/project-audit-task-breakdown.md:449`、`docs/project-audit-task-breakdown.md:450`。
2. `dispatchClickDefault` 是本组控件用户 click/keyboard 默认行为的核心 gate。它会补齐 cancelable、bubbling 的 `click` 事件，调用 `event.DispatchPointerEvent`，只有返回 allowed 时才执行默认 action。因此经由该函数的选中、切换、打开、关闭默认行为可被 `PreventDefault` 阻止。证据：`widget/event_defaults.go:8`、`widget/event_defaults.go:13`、`widget/event_defaults.go:15`、`widget/event_defaults.go:17`、`widget/event_defaults.go:18`、`widget/event_defaults.go:31`、`widget/event_defaults.go:32`、`widget/event_defaults.go:33`。
3. `Checkbox` 是受控值控件：组件入参 `checked` 存入 `checkboxWidget.value`，用户激活只计算 `next := !c.value` 并调用 `CheckboxOnChange`，不在组件内部持久保存 next。外部必须用新的 checked 重渲染才会改变显示值。证据：`widget/checkbox.go:27`、`widget/checkbox.go:29`、`widget/checkbox.go:33`、`widget/checkbox.go:40`、`widget/checkbox.go:42`、`widget/checkbox.go:88`、`widget/checkbox.go:89`、`widget/checkbox.go:91`。
4. `Checkbox` 的 keyboard 和 pointer 默认行为都可取消：非 disabled 时注册 focus target，`FocusActivate` 调用 `dispatchClickDefault`；pointer 点击通过 `ClickedEvent` 进入同一个 gate。disabled 时注册 disabled focus target，不处理点击，Ref 命令 drain 后也跳过。证据：`widget/checkbox.go:94`、`widget/checkbox.go:97`、`widget/checkbox.go:98`、`widget/checkbox.go:103`、`widget/checkbox.go:104`、`widget/checkbox.go:122`、`widget/checkbox.go:126`。
5. `CheckboxRef` 支持 `SetChecked` 和 `Toggle`，布局期消费。`SetChecked` 与当前 value 相同不触发 `onChange`；`Toggle` 基于本帧原始 `c.value` 计算 next。同帧多个 Toggle 不累积，可能重复发出同一个 next，这延续 A9.2/A9.3 已记录风险。证据：`widget/component_refs.go:194`、`widget/component_refs.go:202`、`widget/component_refs.go:212`、`widget/checkbox.go:107`、`widget/checkbox.go:109`、`widget/checkbox.go:112`、`widget/checkbox.go:114`、`widget/checkbox.go:115`。
6. `Switch` 与 Checkbox 的事件和受控值模型一致：入参 `checked` 存入 `switchWidget.value`，激活时计算相反值并调用 `SwitchOnChange`；keyboard focus activate 和 pointer click 均通过 `dispatchClickDefault`，disabled 时 focus/click/Ref 都被跳过。证据：`widget/switch.go:35`、`widget/switch.go:40`、`widget/switch.go:49`、`widget/switch.go:151`、`widget/switch.go:152`、`widget/switch.go:154`、`widget/switch.go:157`、`widget/switch.go:160`、`widget/switch.go:189`。
7. `SwitchRef` 支持 `SetChecked` 和 `Toggle`，消费规则与 CheckboxRef 相同：disabled 时跳过；Set 同值不触发；Toggle 基于本帧原始 value，不累积本帧多命令。证据：`widget/component_refs.go:233`、`widget/component_refs.go:241`、`widget/component_refs.go:251`、`widget/switch.go:164`、`widget/switch.go:167`、`widget/switch.go:170`、`widget/switch.go:172`、`widget/switch.go:175`、`widget/switch.go:177`。
8. `RadioGroup` 是受控单选组，但布局期会有本帧局部 `currentValue`。Ref 命令先更新 `currentValue` 并调用 `RadioGroupOnChange`；渲染每个 item 时以该局部 current 判定 checked。外部仍需要用新的 value 重渲染来长期保持状态。证据：`widget/selection.go:41`、`widget/selection.go:48`、`widget/selection.go:57`、`widget/selection.go:113`、`widget/selection.go:114`、`widget/selection.go:123`、`widget/selection.go:145`。
9. `RadioGroup` 每个 item 都是独立 focus/click target。非 disabled 时 keyboard 和 pointer 都通过 `dispatchClickDefault`；如果点击当前已选 item，`activate` 直接 return，不触发 `onChange`。因此用户默认选择行为可取消，且同值用户选择有过滤。证据：`widget/selection.go:147`、`widget/selection.go:148`、`widget/selection.go:149`、`widget/selection.go:150`、`widget/selection.go:153`、`widget/selection.go:155`、`widget/selection.go:161`、`widget/selection.go:171`。
10. `RadioGroupRef` 只有 `SetValue`，无 open/popup 语义。布局期若 next 等于当前 `currentValue` 则跳过；disabled 时命令 drain 但不执行。该 Ref 命令属于程序写入，不经过 `dispatchClickDefault`，因此不可被事件 `PreventDefault` 取消。证据：`widget/component_refs.go:327`、`widget/component_refs.go:335`、`widget/selection.go:115`、`widget/selection.go:118`、`widget/selection.go:120`、`widget/selection.go:123`、`widget/selection.go:125`。
11. `Select` 是受控值 + 内部 popup open state 的混合控件。`currentValue := s.value` 每帧从外部 value 开始；popup open state 存在 `selectState` 中，由 trigger、option、outside press 或 `SelectRef` 改变。证据：`widget/selection.go:269`、`widget/selection.go:286`、`widget/selection.go:287`、`widget/selection.go:295`、`widget/selection.go:301`、`widget/selection.go:485`、`widget/selection.go:486`、`widget/selection.go:487`。
12. `Select` trigger 的 click/keyboard 默认行为可取消：`layoutSelectField` 非 disabled 时注册 focus target，focus activate 和 pointer click 均通过 `dispatchClickDefault` 执行 `state.opened = !state.opened` 并调用 `SelectOnOpenChange`。disabled 时关闭已打开 popup，并跳过 trigger 激活。证据：`widget/selection.go:488`、`widget/selection.go:750`、`widget/selection.go:752`、`widget/selection.go:753`、`widget/selection.go:755`、`widget/selection.go:758`、`widget/selection.go:761`、`widget/selection.go:771`。
13. `Select` option 的 click/keyboard 默认行为也可取消：每个 option row 注册 focus target；非 disabled option 的 focus activate 和 pointer click 通过 `dispatchClickDefault` 执行选值、`SelectOnChange`、关闭 popup 和 `SelectOnOpenChange(false)`。但 option 激活没有同值过滤，点击当前 active option 仍会调用 `onChange` 并关闭 popup。证据：`widget/selection.go:548`、`widget/selection.go:550`、`widget/selection.go:551`、`widget/selection.go:552`、`widget/selection.go:554`、`widget/selection.go:557`、`widget/selection.go:566`、`widget/selection.go:576`。
14. `Select` popup 通过 `md3OverlayProgress`、`ListView` 和 `op.Defer` 挂载。popup 仅在 `state.opened && len(s.options) > 0` 时可见；outside press 用 fieldRect 与 popupRect 作为 protected rect，外部点击关闭并调用 `SelectOnOpenChange(false)`。证据：`widget/selection.go:528`、`widget/selection.go:531`、`widget/selection.go:654`、`widget/selection.go:662`、`widget/selection.go:727`、`widget/selection.go:732`、`widget/selection.go:734`、`widget/selection.go:745`。
15. `SelectRef` 支持 `SetValue`、`Open`、`Close`、`Toggle`。这些命令在布局期直接改变 current/open 状态并调用 `onChange`/`onOpen`，不经过 `dispatchClickDefault`，所以不可由事件监听器取消。`SetValue` 对 `onChange` 有同值过滤，但仍会把 `currentValue = cmd.value` 写入本帧局部值。证据：`widget/component_refs.go:370`、`widget/component_refs.go:378`、`widget/component_refs.go:388`、`widget/component_refs.go:395`、`widget/component_refs.go:402`、`widget/selection.go:491`、`widget/selection.go:497`、`widget/selection.go:498`、`widget/selection.go:501`、`widget/selection.go:517`。

### 控件矩阵

| 控件 | click 默认行为 | keyboard 默认行为 | controlled value | popup | Ref | 可取消性 |
| --- | --- | --- | --- | --- | --- | --- |
| Checkbox | 点击切换为 `!value` 并调用 `CheckboxOnChange` | focus activate 同 click | 外部 checked 入参为权威值；内部不持久保存 next | 无 | `SetChecked`、`Toggle` | 用户 click/keyboard 可取消；Ref 不可取消 |
| Switch | 点击切换为 `!value` 并调用 `SwitchOnChange` | focus activate 同 click | 外部 checked 入参为权威值；内部不持久保存 next | 无 | `SetChecked`、`Toggle` | 用户 click/keyboard 可取消；Ref 不可取消 |
| RadioGroup | 点击不同 item 设为该 value；当前项 no-op | 每个 item focus activate 同 click | 外部 value 为权威值；本帧 `currentValue` 用于 Ref/点击后的局部一致性 | 无 | `SetValue` | 用户 click/keyboard 可取消；Ref 不可取消；同值用户选择过滤 |
| Select trigger | 点击切换 `state.opened` | trigger focus activate 同 click | 外部 value 为权威值；open state 内部保存 | 有，本地 `op.Defer` popup | `Open`、`Close`、`Toggle` | 用户 trigger click/keyboard 可取消；Ref open/close/toggle 不可取消 |
| Select option | 点击 option 设置 value、触发 `onChange`、关闭 popup | option focus activate 同 click | 本帧 `currentValue` 更新；长期依赖外部重渲染 | option 位于 popup ListView | `SetValue` | 用户 option click/keyboard 可取消；Ref set 不可取消；当前项重复选择未过滤 |

## 风险

1. Checkbox/Switch 的同帧多次 Ref `Toggle` 基于原始入参，不累积中间结果，可能重复发出同一个 next。
2. Select option 点击当前 active option 仍会触发 `SelectOnChange` 并关闭 popup；这与 RadioGroup 当前项 no-op 不一致。
3. Select 的用户默认行为可取消，但 SelectRef 和 outside press 关闭不经过 event default action，事件监听器无法 `PreventDefault`。
4. Checkbox/Switch/RadioGroup/Select 都依赖外部 value 重渲染长期确认值；如果调用方不更新 value，下一帧会回到旧值。
5. Select popup open state 是内部状态，value 是外部受控状态；Ref、outside press、trigger、option 会同时改变 open/value 两类状态，后续修复需要避免同一帧回调顺序不稳定。

## 验收

- 已输出 Checkbox、Switch、RadioGroup、Select 的 click、keyboard、controlled value、popup、Ref 矩阵。
- 已明确默认行为可取消性：Checkbox/Switch/RadioGroup/Select trigger/Select option 的用户 click 和 keyboard 默认行为都经过 `dispatchClickDefault`，可被 `PreventDefault` 阻止。
- 已明确不可取消路径：CheckboxRef/SwitchRef/RadioGroupRef/SelectRef 命令和 Select outside press 关闭不经过 default action gate。
- 已标出 Select 与 RadioGroup 的差异：RadioGroup 当前项选择 no-op，Select 当前 active option 会重复触发 `onChange`。
- 已确认本任务只做审查记录，没有修改运行时代码或组件行为。

## 后续依赖

- A10 后续修复应统一选择控件的同值选择语义，尤其是 Select 当前 active option 是否应与 RadioGroup 一样 no-op。
- 建议补充回归测试：`PreventDefault` 阻止 Checkbox/Switch/RadioGroup/Select trigger/Select option 默认行为，Ref 命令不受 `PreventDefault` 影响。
- Select 文档需要明确：`SelectOnOpenChange` 可由 trigger、option close、outside press、Ref open/close/toggle 触发；这些来源当前没有 source 字段。
