# A9.2 controlled value 和内部状态审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph explore "A9.2 controlled value internal state Input Select RadioGroup Checkbox Switch Slider Tabs Value OnChange State widget"`
  - `codegraph explore "widget/input.go Input TextField SearchBar inputWidget Layout controlled value internal state OnChange InputRef SetText beforeinput input change submit"`
  - `codegraph explore "inputWidget Layout syncExternalValue processInputRefCommands finishInputMutation dispatchInputEvent InputOnChange InputValue Priority"`
  - `codegraph explore "radioGroupWidget Layout switchWidget Layout sliderWidget Layout tabsWidget Layout Checkbox Layout onChange ref currentValue activeIndex controlled internal state"`
  - 局部源码行号抽取：`widget/input.go`、`widget/selection.go`、`widget/slider.go`、`widget/switch.go`、`widget/tabs_dialog_toast.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/input.go`
  - `widget/selection.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/slider.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/component_refs.go`
  - `widget/material3_components.go`
  - `ui/extended_types.go`

## 事实结论

1. A9.2 的目标是审查 Input、Select、RadioGroup、Checkbox、Switch、Slider、Tabs 的外部 value、内部 state、`OnChange` 优先级；验收要求同一帧不会互相覆盖。该任务属于 A9“声明式输入、内部状态和命令式 Ref 不互相打架”目标的一部分。证据：`docs/project-audit-roadmap.md:276`、`docs/project-audit-roadmap.md:281`、`docs/project-audit-task-breakdown.md:420`、`docs/project-audit-task-breakdown.md:424`、`docs/project-audit-task-breakdown.md:425`、`docs/project-audit-task-breakdown.md:426`。
2. 本批组件的主模型是“外部 value 是绘制基线，内部 state 只保存交互过程、打开状态、动画/焦点/编辑器状态，值变更通过 `OnChange` 通知调用方下一帧回填”。`ui/extended_types.go` 也把 `TextFieldElement`、`SliderElement`、`RadioGroupElement`、`SelectElement`、`TabsElement` 标为受控入口。证据：`ui/extended_types.go:474`、`ui/extended_types.go:487`、`ui/extended_types.go:497`、`ui/extended_types.go:502`、`ui/extended_types.go:660`。
3. `Input`/`TextField` 只有在配置了 `InputOnChange` 时才进入 controlled 同步：初始化总是 `editor.SetText(t.value)`；后续仅当 `controlled && t.value != state.syncedValue` 时才把外部 value 写回 Gio editor 并重置 input history。没有 `OnChange` 时，外部 value 在初始化后不再覆盖 editor，实际表现为 uncontrolled editor。证据：`widget/input.go:378`、`widget/input.go:380`、`widget/input.go:381`、`widget/input.go:385`、`widget/input.go:393`、`widget/input.go:395`。
4. `Input` 的内部 state 保存 Gio editor、`syncedValue`、focus、history 和 field size。Ref 文本命令先于 editor.Update 消费，调用 `commitProgrammaticInput` 发出 programmatic `beforeinput/input/change` 链路；用户输入来自 Gio `ChangeEvent`，在 `text != state.syncedValue` 时调用 `commitUserInput`；最后 `finishInputMutation` 更新 `syncedValue` 并触发旧 `InputOnChange`。证据：`widget/input.go:89`、`widget/input.go:408`、`widget/input.go:413`、`widget/input.go:426`、`widget/input.go:432`、`widget/input.go:435`、`widget/input.go:604`、`widget/input.go:608`、`widget/input.go:609`。
5. `RadioGroup` 以构造入参 `value` 作为 `currentValue` 基线；Ref 命令和用户点击都会更新本帧局部 `currentValue`，只有新值不同才触发 `onChange`。内部没有独立持久选中值，持久化依赖调用方下一帧传回新 value。证据：`widget/selection.go:48`、`widget/selection.go:103`、`widget/selection.go:104`、`widget/selection.go:109`、`widget/selection.go:113`、`widget/selection.go:114`、`widget/selection.go:139`、`widget/selection.go:143`、`widget/selection.go:144`。
6. `Select` 的值同样以外部 `value` 为每帧基线，布局内使用 `currentValue` 处理 Ref `SetValue` 和 option 激活；`state.opened` 是内部持久状态，disabled 时会强制关闭。选择 option 时本帧局部值先更新，再触发 `onChange`，随后关闭 popup 并触发 `onOpen(false)`。证据：`widget/selection.go:307`、`widget/selection.go:485`、`widget/selection.go:487`、`widget/selection.go:488`、`widget/selection.go:497`、`widget/selection.go:501`、`widget/selection.go:552`、`widget/selection.go:553`、`widget/selection.go:557`、`widget/selection.go:558`。
7. `Checkbox` 和 `Switch` 都以构造入参 `checked` 保存为 `value`，点击/键盘激活时计算 `next := !value` 并调用 `onChange`，组件自身不写入持久 checked state；Ref 命令也基于原始 `value` 计算并触发 `onChange`。证据：`widget/checkbox.go:33`、`widget/checkbox.go:40`、`widget/checkbox.go:89`、`widget/checkbox.go:90`、`widget/checkbox.go:107`、`widget/checkbox.go:112`、`widget/checkbox.go:114`、`widget/switch.go:40`、`widget/switch.go:149`、`widget/switch.go:150`、`widget/switch.go:168`、`widget/switch.go:173`、`widget/switch.go:175`。
8. `Slider` 每帧从外部 config 归一化出 `cfg.value/valueStart/valueEnd`。Ref 命令先应用到局部 cfg；当 thumb 未按下时，内部 progress state 会从 cfg 重建；拖动中则保留内部 progress，避免外部 value 在按压过程直接抢占正在拖动的视觉状态。布局末尾把 progress 还原为 stepped value，值不同且非 disabled 时触发 `onChange`/`onRangeChange`。证据：`widget/slider.go:85`、`widget/slider.go:232`、`widget/slider.go:234`、`widget/slider.go:241`、`widget/slider.go:244`、`widget/slider.go:259`、`widget/slider.go:290`、`widget/slider.go:292`、`widget/slider.go:297`。
9. `Tabs` 以外部 `active` 为每帧基线，Ref 和点击会更新本帧局部 `activeKey` 并调用 `onChange`。内部 `tabsRuntimeState` 只保存 active/previous index 和 indicator 动画进度，用于视觉动画；实际 active key 的持久化仍依赖外部下一帧回填。证据：`widget/tabs_dialog_toast.go:93`、`widget/tabs_dialog_toast.go:183`、`widget/tabs_dialog_toast.go:184`、`widget/tabs_dialog_toast.go:187`、`widget/tabs_dialog_toast.go:191`、`widget/tabs_dialog_toast.go:212`、`widget/tabs_dialog_toast.go:213`、`widget/tabs_dialog_toast.go:218`、`widget/tabs_dialog_toast.go:233`、`widget/tabs_dialog_toast.go:234`。
10. `SearchBar` 是高级包装，保存外部 `value` 和 `SearchBarOnChange`，并通过内部 input options 转发到 TextField；其受控语义应继承 TextField：有 onChange 才会持续同步外部 value。证据：`widget/material3_components.go:2200`、`widget/material3_components.go:2205`、`widget/material3_components.go:2212`、`widget/material3_components.go:2217`、`widget/material3_components.go:2222`。

### 优先级矩阵

| 组件 | 外部 value | 内部 state | Ref/交互写入 | OnChange 触发 | 同帧覆盖结论 |
| --- | --- | --- | --- | --- | --- |
| Input/TextField | 初始化总是写入；有 `OnChange` 时后续作为 controlled 基线 | Gio editor、`syncedValue`、focus、history | Ref 文本命令先于 Gio update；用户输入来自 Gio ChangeEvent | programmatic/user mutation 后触发；submit 走独立事件 | controlled 外部旧值会在下一帧覆盖未回填的 programmatic/user 值；同帧内按“外部同步 -> Ref -> Gio update”顺序 |
| Select | 每帧 `currentValue := value` | `opened`、outside tag、动画状态 | Ref value/open 先处理；option 激活更新本帧 `currentValue` | value 变化时触发；open 变化走 `onOpen` | 值不会持久存内部；open 是内部持久，value 需外部回填 |
| RadioGroup | 每帧 `currentValue := value` | 无持久选中值，只有交互/动画状态 | Ref/点击更新本帧 `currentValue` | 与本帧 current 不同才触发 | 本帧多次变更按 `currentValue` 去重，持久化依赖外部回填 |
| Checkbox | 每帧 `value := checked` | 交互/动画状态 | 点击和 Ref 基于原始 `value` 算 next | next 与原始 value 不同触发 | 多次同帧 toggle 不累积，存在重复同一 next 的风险 |
| Switch | 每帧 `value := checked` | 交互/动画状态 | 点击和 Ref 基于原始 `value` 算 next | next 与原始 value 不同触发 | 与 Checkbox 相同，多次同帧 toggle 不累积 |
| Slider | 每帧 cfg 从 value 归一化 | thumb progress、hover/pressed/capture | Ref 先改局部 cfg；拖动改 progress | progress 还原值与 cfg 不同且非 disabled 时触发 | 未 pressed 时外部 value 重建 progress；pressed 时内部拖动优先 |
| Tabs | 每帧 `activeKey := active` | active/previous index、indicator 动画 | Ref/点击更新本帧 `activeKey` | key 与本帧 activeKey 不同触发 | active key 不持久存内部；动画 state 跟随 active index |

## 风险

1. Checkbox/Switch 的本帧状态没有 `currentValue` 累积；同一帧连续 `Toggle` 或同一帧多个点击事件都会基于旧 `value` 计算，可能重复触发相同 `OnChange(next)`，而不是 true->false 或 false->true 的序列。
2. Input 的 controlled 判定依赖 `onChange != nil`。如果调用方期望“只传 value 不传 OnChange”也能持续受控，当前实现不会满足；初始化后 editor 会保持内部文本。
3. InputRef programmatic mutation 在 controlled 场景下会先触发事件和 `OnChange`，如果外部没有同步回填，下一帧会被旧外部 value 覆盖，表现为短暂写入后回弹。
4. Select/RadioGroup/Tabs 的值不持久保存在内部；所有用户选择都需要外部回填，否则下一帧回到旧入参。Select 的 `opened` 却是内部持久 state，value/open 的责任边界不完全一致，需要文档写清。
5. Slider pressed 期间内部 progress 优先，外部 value/Ref 对视觉 progress 的覆盖会被延后；这保护拖动体验，但也意味着外部强制值在拖动中不是立即可见。
6. SearchBar 作为 TextField 包装需要明确继承 TextField controlled 语义，否则高级组件调用方可能误判 SearchBar 总是受控。

## 验收

- 已输出 Input、Select、RadioGroup、Checkbox、Switch、Slider、Tabs 的外部 value、内部 state、Ref/交互写入、`OnChange` 触发和同帧覆盖矩阵。
- 已确认多数值组件采用“外部 value 为持久事实、内部 state 为交互/动画/打开状态”的受控模型。
- 已标出 Input 的特殊规则：是否持续同步外部 value 取决于是否配置 `InputOnChange`。
- 已标出 Checkbox/Switch 的同帧多次 toggle 不累积风险，这是 A9.2 验收中最需要后续修复或测试锁定的覆盖点。
- 已确认本任务只做审查记录，没有修改运行时代码或组件行为。

## 后续依赖

- A9.3 OnChange 触发条件审查应继续细化“真实变化”的定义，尤其是 Checkbox/Switch 多次同帧 toggle、Select/Radio/Tabs 的同值点击、Slider 拖动中的重复 value。
- A10 组件族修复前应补充 controlled value 回归测试：Input 无 OnChange/uncontrolled、有 OnChange/controlled、Ref 后未回填回弹、Checkbox/Switch 多 Toggle、Slider pressed 期间外部覆盖。
- 文档 API 说明应明确：Select/RadioGroup/Tabs/Checkbox/Switch/Slider 的 value 均由调用方负责持久化，内部 state 不会替调用方保存最终值。
