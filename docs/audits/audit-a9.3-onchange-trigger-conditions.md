# A9.3 OnChange 触发条件审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 4：复杂边界和组件族。

- 状态：Done
- 日期：2026-07-07
- 负责人：Codex
- 关注：Widget、Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `go_workspace`
  - `go_vulncheck ./...`
  - `codegraph_explore "A9.3 OnChange trigger conditions form controls Input TextField SearchBar Select RadioGroup Checkbox Switch Slider Tabs ScrollView ScrollOnChange real change programmatic ref disabled same value"`
  - `codegraph_explore "widget/input.go inputWidget Layout InputOnChange dispatchTextChangeIfNeeded applyInputCommand TextField onChange user programmatic; widget/switch.go switchWidget Layout SwitchOnChange; widget/slider.go sliderWidget Layout SliderOnChange SliderRef drag; widget/tabs_dialog_toast.go tabsWidget Layout TabsOnChange TabsRef; widget/list_grid.go scrollWidget Layout ScrollOnChange call changed lastOffset"`
  - 局部源码行号抽取：`widget/input.go`、`widget/selection.go`、`widget/checkbox.go`、`widget/switch.go`、`widget/slider.go`、`widget/tabs_dialog_toast.go`、`widget/list_grid.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/input.go`
  - `widget/selection.go`
  - `widget/checkbox.go`
  - `widget/switch.go`
  - `widget/slider.go`
  - `widget/tabs_dialog_toast.go`
  - `widget/list_grid.go`
  - `widget/material3_components.go`
  - `widget/input_event_test.go`
  - `widget/slider_interaction_test.go`
  - `widget/list_grid_test.go`

## 事实结论

1. A9.3 的目标是审查表单控件、Slider、Tabs、ScrollView 的 `OnChange` 是否只在真实变化时触发，并明确程序设置和用户操作是否触发回调。该任务延续 A9 对声明式 value、内部 state 和命令型 Ref 边界的审查。证据：`docs/project-audit-task-breakdown.md:428`、`docs/project-audit-task-breakdown.md:432`、`docs/project-audit-task-breakdown.md:433`、`docs/project-audit-task-breakdown.md:434`。
2. `Input`/`TextField` 的用户输入有真实变化过滤：Gio `ChangeEvent` 后只有 `editor.Text() != state.syncedValue` 才调用 `commitUserInput`。但 Ref programmatic 命令没有同值过滤，`SetText`、`Append`、`Clear` 会进入 `commitProgrammaticInput`，随后 `finishInputMutation` 总是派发 `input/change` 并调用旧 `InputOnChange`。证据：`widget/input.go:407`、`widget/input.go:410`、`widget/input.go:414`、`widget/input.go:429`、`widget/input.go:431`、`widget/input.go:432`、`widget/input.go:568`、`widget/input.go:582`、`widget/input.go:600`、`widget/input.go:608`。
3. 外部 value 重渲染不是 `InputOnChange` 来源：初始化会写入 editor；后续只有 configured `onChange` 时才把外部 value 同步到 editor，并且这条同步路径不调用 `finishInputMutation`。因此程序设置分为两类：声明式 value 设置不触发旧回调，`InputRef` programmatic mutation 会触发旧回调。证据：`widget/input.go:375`、`widget/input.go:377`、`widget/input.go:382`、`widget/input.go:390`、`widget/input.go:391`、`widget/input.go:608`。
4. `SearchBar` 将 `SearchBarOnChange` 转发为内部 `InputOnChange`，所以继承 TextField 规则：用户真实输入变化触发，`InputRef` 类 programmatic mutation 触发，纯外部 value 重渲染不触发。证据：`widget/material3_components.go:2205`、`widget/material3_components.go:2233`、`widget/material3_components.go:2280`、`widget/material3_components.go:2281`、`widget/material3_components.go:2287`。
5. `RadioGroup` 对 Ref 和用户点击都以本帧 `currentValue` 去重。Ref 命令只有 `next != currentValue` 时触发；点击当前已选 item 会直接 return。禁用态注册 disabled focus target，且不处理点击。证据：`widget/selection.go:103`、`widget/selection.go:109`、`widget/selection.go:113`、`widget/selection.go:114`、`widget/selection.go:139`、`widget/selection.go:140`、`widget/selection.go:143`、`widget/selection.go:144`、`widget/selection.go:148`、`widget/selection.go:155`。
6. `Select` 的 Ref `SetValue` 有同值过滤，只有 `cmd.value != currentValue` 时触发 `onChange`；但 option 激活没有同值过滤，点击当前 active option 仍会调用 `onChange` 并关闭 popup。禁用的 Select 或禁用 option 不处理激活。证据：`widget/selection.go:493`、`widget/selection.go:498`、`widget/selection.go:529`、`widget/selection.go:530`、`widget/selection.go:531`、`widget/selection.go:532`、`widget/selection.go:535`、`widget/selection.go:542`、`widget/selection.go:549`。
7. `Checkbox`/`Switch` 的用户激活总是把当前入参取反后调用 `onChange`，因此单次点击是“真实变化”。Ref `Set` 命令有 `next != value` 过滤，Ref `Toggle` 也会产生相反值。但本帧多次 toggle 不累积，多个事件会重复基于原始 `value` 计算同一个 `next`，可能多次触发相同回调。证据：`widget/checkbox.go:88`、`widget/checkbox.go:89`、`widget/checkbox.go:90`、`widget/checkbox.go:103`、`widget/checkbox.go:107`、`widget/checkbox.go:112`、`widget/checkbox.go:114`、`widget/switch.go:151`、`widget/switch.go:152`、`widget/switch.go:153`、`widget/switch.go:166`、`widget/switch.go:170`、`widget/switch.go:175`、`widget/switch.go:177`。
8. `Slider` 的 `SliderOnChange` 不是外部 value 或 Ref 设置回调。Ref 命令先写入局部 cfg；未 pressed 时 state progress 立即从 cfg 重建，布局末尾比较 `endValue` 与 `cfg.value`，通常不会触发。用户拖动改变内部 progress 后，末尾用 epsilon `> 0.0001` 判断真实变化，非 disabled 才触发 `onChange`。证据：`widget/slider.go:235`、`widget/slider.go:241`、`widget/slider.go:244`、`widget/slider.go:255`、`widget/slider.go:259`、`widget/slider.go:290`、`widget/slider.go:292`、`widget/slider.go:297`、`widget/slider.go:298`。
9. `Tabs` 的 Ref 命令有同值过滤：命令 key 等于当前 `activeKey` 时 continue；不同才触发 `onChange`。用户点击路径没有同值过滤，点击当前 active tab 也会调用 `onChange`。disabled tab 不处理点击。证据：`widget/tabs_dialog_toast.go:188`、`widget/tabs_dialog_toast.go:191`、`widget/tabs_dialog_toast.go:192`、`widget/tabs_dialog_toast.go:232`、`widget/tabs_dialog_toast.go:233`、`widget/tabs_dialog_toast.go:234`、`widget/tabs_dialog_toast.go:235`、`widget/tabs_dialog_toast.go:236`。
10. `ScrollView` 的 `ScrollOnChange` 只在 Gio List 的 `Position.First` 或 `Position.Offset` 相对 `lastFirst/lastOff` 改变时触发。用户 wheel、Scrollbar、auto-to-end、`ScrollRef` 等只要最终导致 position 变化，都可能触发；纯重绘或 position 不变不会触发。回调值是 `first + off/1024` 的抽象 offset，不是像素。证据：`widget/list_grid.go:123`、`widget/list_grid.go:125`、`widget/list_grid.go:139`、`widget/list_grid.go:155`、`widget/list_grid.go:157`、`widget/list_grid.go:163`、`widget/list_grid.go:166`、`widget/list_grid.go:170`、`widget/list_grid.go:172`。
11. 现有测试覆盖不均：`widget/input_event_test.go` 覆盖 Input 事件和 `InputOnChange`；`widget/slider_interaction_test.go` 覆盖 Slider 拖动触发；`widget/list_grid_test.go` 覆盖 ScrollView 回调。未看到针对 Select/Tabs 当前项重复点击、InputRef 同值命令、Checkbox/Switch 同帧多 toggle 的直接回归测试。证据：`widget/input_event_test.go:48`、`widget/input_event_test.go:87`、`widget/slider_interaction_test.go:51`、`widget/list_grid_test.go:187`、`widget/list_grid_test.go:244`。

### OnChange 触发矩阵

| 组件 | 用户操作 | 程序设置/Ref | 外部 value 重渲染 | 同值过滤 | 结论 |
| --- | --- | --- | --- | --- | --- |
| Input/TextField | Gio ChangeEvent 后文本不同才触发 | `InputRef` Set/Append/Clear 会触发，即使同值也可能触发 | 不触发旧 `InputOnChange` | 用户输入有；Ref 无 | 需要区分声明式设置和 Ref programmatic mutation |
| SearchBar | 继承 TextField | 继承内部 TextField options | 不触发旧 `SearchBarOnChange` | 继承 TextField | 高级包装不改变触发规则 |
| RadioGroup | 点击不同项触发；点击当前项不触发 | Ref 不同值触发 | 不触发 | 有 | 符合“真实变化” |
| Select | 点击 option 触发，包括当前 active option | Ref SetValue 不同值触发 | 不触发 | Ref 有；用户 option 无 | 当前项重复选择会触发重复回调 |
| Checkbox | 单次点击/键盘激活触发取反 | Ref Set 同值不触发；Toggle 触发取反 | 不触发 | Set 有；Toggle 基于原入参 | 单帧多 toggle 可能重复同一 next |
| Switch | 同 Checkbox | 同 Checkbox | 不触发 | 同 Checkbox | 风险同 Checkbox |
| Slider | 拖动导致 stepped value 不同触发 | 一般 Ref 设置不作为 onChange 来源 | 不触发 | epsilon 比较 | 用户交互真实变化触发，程序设置不回调 |
| Tabs | 点击 tab 触发，包括当前 active tab | Ref 不同 key 触发 | 不触发 | Ref 有；用户点击无 | 当前 tab 重复点击会触发重复回调 |
| ScrollView | wheel/scrollbar 导致 position 变化触发 | ScrollRef/autoToEnd 导致 position 变化触发 | 不触发 | `First/Offset` 比较 | 回调代表最终滚动位置变化，不区分来源 |

## 风险

1. InputRef 同值 `SetText` 或空值 `Clear` 仍可能触发 `beforeinput/input/change/InputOnChange`，不满足“只在真实变化时触发”的严格定义。
2. Select 点击当前 active option 会重复触发 `SelectOnChange`，还会关闭 popup；这会让“重新确认当前值”和“值变化”在旧回调层面不可区分。
3. Tabs 点击当前 active tab 会重复触发 `TabsOnChange`；如果调用方在回调里做重载或埋点，会得到重复变化信号。
4. Checkbox/Switch 同帧多次 toggle 基于原始入参，不累积本帧 current 值，可能重复发出同一个 next。
5. Slider 的程序设置/Ref 设置通常不触发 `SliderOnChange`，而 InputRef programmatic mutation 会触发 `InputOnChange`；不同组件的 programmatic callback 规则不一致，需要文档化。
6. ScrollView 的 programmatic `ScrollRef` 和用户 wheel 都可能触发 `ScrollOnChange`，但回调签名没有 source 字段，调用方无法区分来源。

## 验收

- 已输出 Input/SearchBar、RadioGroup、Select、Checkbox、Switch、Slider、Tabs、ScrollView 的 `OnChange` 触发矩阵。
- 已明确用户操作、外部 value 重渲染、Ref/programmatic 设置三类来源是否触发旧回调。
- 已标出符合真实变化过滤的路径：RadioGroup、Slider 拖动、ScrollView position diff、Checkbox/Switch Ref Set、Select/Tabs Ref。
- 已标出不满足严格真实变化过滤的路径：InputRef 同值 programmatic mutation、Select 当前项点击、Tabs 当前项点击、Checkbox/Switch 同帧多 toggle。
- 已确认本任务只做审查记录，没有修改运行时代码或组件行为。

## 后续依赖

- A10 修复阶段应决定是否统一旧 `OnChange` 语义：programmatic Ref 是否触发、同值用户确认是否触发、重复同帧 toggle 是否合并。
- 建议补充回归测试：InputRef 同值 SetText/Clear、Select 当前项重复点击、Tabs 当前 tab 点击、Checkbox/Switch 同帧多 Toggle、ScrollRef 与用户 wheel 的 `ScrollOnChange` 来源差异。
- API 文档应明确：外部 value 重渲染不是 `OnChange` 来源；`InputRef` 和 `ScrollRef` 当前会产生 programmatic callback/position callback，SliderRef 通常不会。
