<!-- fluxui-doc-meta
{
  "id": "project_audit_roadmap",
  "title": "大项目逻辑审查路线图",
  "category": "工程路线图",
  "order": 45,
  "summary": "按功能、控件和底层能力的关联关系审查 FluxUI，先确认依赖链路，再逐项检查运行时、布局、事件、状态、样式、测试和文档，避免修复引入跨组件混乱。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Context",
    "internal.Runtime",
    "ui.Element",
    "widget.Widget",
    "event.Event",
    "event.PointerEvent",
    "event.KeyboardEvent",
    "widget.ScrollView",
    "widget.Input",
    "widget.Dialog",
    "widget.Popup"
  ]
}
-->

# FluxUI 大项目逻辑审查路线图

本文档用于一次系统级代码审查。目标不是马上修复某个单点问题，而是先确认每个功能、控件和底层系统之间的依赖关系，再逐层审查逻辑是否一致、职责是否清晰、是否存在重复实现、隐式耦合和互相覆盖的行为。

这次审查必须以“关联链路”为核心：一个控件出现问题时，不能只看控件文件本身，还要看它依赖的布局、事件注册、命中测试、滚动、焦点、样式、动画、状态保存、Ref、示例和文档。

阶段执行和任务量拆分见：`docs/project-audit-task-breakdown.md`。

## 审查原则

- 先审查，后修复。审查阶段只记录事实、风险和证据，不做行为改动。
- 先确定边界，再改代码。每个功能必须明确属于 runtime、event、ui、widget、style、theme、system 还是 examples。
- 先建立最小复现，再做底层修复。没有复现和验收条件的修复不进入实施。
- 运行时代码和核心 widget 变更必须小步推进，不能一次改多个交互系统。
- 旧 API 的兼容边界必须单独记录，例如 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`。
- 每个控件都要审查“默认行为”和“用户回调”的顺序，避免回调、事件系统和 Gio 原始事件互相抢输入。
- 每次修复必须有对应的回归测试或可手动验收示例。

## 产物

每个审查阶段完成后都应留下以下产物：

- 关联矩阵：该功能依赖哪些底层能力。
- 事实清单：当前 API、内部状态、事件路径、默认行为和测试覆盖。
- 风险清单：可能影响其他控件或示例的耦合点。
- 冗余清单：重复状态、重复事件处理、重复样式计算、重复 layout 逻辑。
- 修复建议：只描述建议，不在审查阶段直接改。
- 验收用例：自动测试、示例操作步骤、docs browser 手动验收项。

## 总体审查顺序

### A0 工作区基线

目标：确认审查从干净、可重复的状态开始。

审查项：

- `git status` 只允许保留明确的文档或审查记录变更。
- 记录当前 Go 版本、Gio 版本、主要依赖版本。
- 运行现有核心测试：`go test ./event ./internal ./ui ./widget`。
- 运行示例编译 smoke test，不做全量重型测试。
- 整理已知问题，不把历史问题混入新修复。

验收：

- 有一份“当前基线状态”记录。
- 明确哪些失败是既有问题，哪些是新问题。

### A1 包边界和依赖方向

目标：确认代码结构的所有权，避免高层控件反向污染底层运行时。

审查项：

- `internal`：frame、runtime、context、tree、interaction、render、path memory。
- `event`：公开事件类型、分发 API、Gio 数据转换、兼容旧事件。
- `ui`：Element、builder、reconciler、hooks、router。
- `widget`：组件实现、组件状态、Ref、默认行为。
- `layout`：约束、尺寸、轴向、滚动容器。
- `style` / `theme`：样式结构、默认值、Material 3 语义。
- `examples` / `docs`：真实使用方式和验收入口。

重点判断：

- 包依赖是否单向。
- 是否存在 widget 为了修某个场景直接修改 runtime 全局规则。
- 是否存在 event 和 widget 各自维护一套交互状态。
- 是否存在 ui Element 层和 widget 层重复表达同一行为。

验收：

- 输出包依赖图。
- 标注允许依赖和禁止依赖。
- 标注需要收敛的重复实现。

### A2 Runtime、Context 和 Frame 生命周期

目标：确认所有 UI、事件和状态都建立在稳定的 frame 生命周期上。

审查项：

- `Context` 中公开给 widget 的能力是否过宽。
- `Runtime` 的 frame 开始、frame 结束、invalidating、redraw reason 是否一致。
- `PathID` / target identity 是否稳定。
- 子树路径、portal 路径、overlay 路径是否有统一规则。
- 组件状态是否按路径保存，是否会因布局变化错配。
- 事件注册、focus target、pointer target、scroll target 是否在同一帧生命周期内清理。

风险点：

- 状态挂在错误路径导致控件之间串状态。
- 上一帧事件目标残留，导致鼠标不移动时点击失效。
- frame 内注册顺序变化导致事件命中不稳定。

验收：

- 有 runtime 生命周期图。
- 有路径身份规则。
- 有“上一帧状态可保留”和“每帧必须重建”的字段清单。

### A3 Layout、尺寸和约束链路

目标：确认布局系统不会被交互修复破坏。

审查项：

- 每个容器是否正确传递 Gio constraints。
- `Row`、`Column`、`Stack`、`Center`、`Container`、`Padding`、`FixedWidth/Height` 的尺寸语义是否一致。
- `ScrollView`、`ListView`、`Grid` 是否正确处理主轴和交叉轴。
- 横向滚动和纵向滚动是否有明确轴向边界。
- overlay、dialog、popup 是否影响正常文档流。
- 表格、代码块、chips、tabs 等横向内容是否会产生无限宽度。

风险点：

- 子组件为交互注册扩大了命中区域，同时也意外扩大布局尺寸。
- 滚动容器把无限约束传给普通内容，导致滚动条异常。
- overlay 使用普通布局占位，影响底层页面尺寸。

验收：

- 每个布局基础组件有约束输入/输出记录。
- 横向滚动内容有最大宽度策略。
- docs browser 的代码块、表格、tabs、chips 不阻挡页面纵向滚动。

### A4 渲染、样式和动画链路

目标：确认视觉样式、交互态和动画不互相覆盖。

审查项：

- decoration、background、border、shadow、ripple、state layer 的绘制顺序。
- hover、pressed、focused、disabled、selected、error 等状态的样式来源。
- Material 3 默认样式和用户传入 `Decoration` 的合并规则。
- 动画是否只影响视觉，不改变布局和命中区域。
- toast、snackbar、dialog、popup、select menu 的动画是否共享过多隐式状态。

风险点：

- 某个控件为了修交互，把全局鼠标 cursor 或 hover 状态污染到整窗。
- 动画期间控件逻辑状态和视觉状态不一致。
- 用户 decoration 覆盖默认样式时丢失必要的交互态。

验收：

- 每类状态有唯一来源。
- 动画不改变控件真实可点击边界，除非文档明确。
- component_lab 不出现整窗 cursor 异常。

### A5 事件系统和输入分发

目标：确认 Gio 原始事件、FluxUI 高级事件和旧 API 之间的关系。

审查项：

- Gio `pointer.Event`、`key.Event`、editor event 到 FluxUI 事件的转换。
- `event.Event` 的 capture、target、bubble、default action 是否有明确调用顺序。
- `StopPropagation`、`StopImmediatePropagation`、`PreventDefault` 是否只影响它们应该影响的层。
- `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange` 的兼容桥接。
- `PointerArea`、`KeyboardScope`、`EventBoundary` 是否只注册自身区域，不扩大到整行或整窗。
- synthetic dispatch 是否不会污染真实输入状态。

风险点：

- 旧事件和新事件各执行一次默认行为。
- `PreventDefault` 标记存在，但组件默认行为没有读取它。
- pointer 命中区域来自 layout 外层，而不是真实绘制/交互区域。

验收：

- 事件顺序有表格记录。
- 每个默认行为都有“可取消/不可取消”结论。
- P1/P2/P3/P5 的测试台说明能让使用者看懂。

### A6 滚动、手势和嵌套交互

目标：统一处理滚轮、触控板、拖拽、输入框、下拉菜单和嵌套滚动。

审查项：

- `ScrollView`、`ListView`、`Grid` 的滚动状态和 `ScrollOnChange` offset。
- 单行 `Input`、拖拽组件、slider、chips、tabs 是否会截停父页面纵向滚动。
- 横向滚动必须由横向 delta 或明确的 shift-wheel 策略触发，不能误用纵向滚动。
- 滚动后不移动鼠标，点击目标是否立即更新。
- 下拉菜单内部滚动是否保持点击和 hover 命中正确。
- nested scroll 是否有明确的事件传递和剩余 delta 策略。

风险点：

- 子控件注册 wheel 区域后吞掉父级滚动。
- scroll offset 更新了，但 hit-test cache 没有刷新。
- 横向滚动区域过宽，覆盖同一行其他控件。

验收：

- docs browser 文档页、代码块、表格、tabs、chips 都能正常滚动。
- `list_view_basic`、`scroll_view_basic` 可以滚动。
- 下拉选项滚动后无需移动鼠标即可点击当前项。

### A7 Focus、Keyboard 和 Text Input

目标：把焦点、键盘、文本输入、IME 组合拆清楚。

审查项：

- focus target 注册、tab order、disabled/hidden 处理。
- `focus/blur/focusin/focusout` 的顺序。
- `keydown/keyup`、局部 shortcut、global shortcut 的边界。
- Button、Select、Menu、Dialog 对 Enter、Space、Escape、Arrow keys 的默认行为。
- `Input` 的 beforeinput、input、change、submit 和 composition 事件。
- 程序调用 `InputRef.SetText` 与用户输入的来源区分。

风险点：

- keyboard scope 注册了，但没有连接到当前 focus path。
- Input 把 wheel 或 pointer 全部截停。
- IME 组合期间提前触发 submit 或 change。

验收：

- P3 测试台每项都有可观察反馈。
- 单行输入框上纵向滚轮仍能滚动父页面。
- 局部快捷键不污染系统全局快捷键。

### A8 Overlay、Portal、Modal 和 Popup 边界

目标：统一 dialog、popup、menu、select dropdown、tooltip、toast 的挂载和事件路径。

审查项：

- overlay 是全局挂载还是局部挂载。
- popup 的逻辑 owner 和布局 parent 是否一致。
- modal boundary 是否截断事件传播。
- 点击遮罩关闭和点击内容关闭是否严格区分。
- 打开弹窗的那一次点击是否会立刻触发外部关闭。
- Dialog、Popup、Menu、Select、Tooltip 是否复用同一套外部点击规则。

风险点：

- 使用全屏 Pressable 当遮罩导致内部按钮点击冒泡后关闭。
- portal 后 event path 丢失 owner，父组件无法观察或无法阻止。
- 弹窗内部滚动刷新后，命中目标仍停留在旧项。

验收：

- 点击弹窗内部按钮、空白、可滚动组件不会误关闭。
- 点击遮罩才按配置关闭。
- Popup 和 Dialog 的 modal/非 modal 行为有明确差异。

### A9 组件状态、Ref 和默认行为

目标：确认声明式输入、内部状态和命令式 Ref 不互相打架。

审查项：

- `ButtonRef`、`InputRef`、`ScrollRef`、`SliderRef`、`DialogRef`、`PopupRef`、`SelectRef` 的命令何时生效。
- controlled value 和内部状态的优先级。
- `OnChange` 是否只在真实变化时触发。
- Ref 命令是否触发事件、是否触发旧回调、是否触发 default action。
- 组件卸载后 Ref 是否安全。

风险点：

- 同一帧内外部 value 和 Ref command 互相覆盖。
- 程序设置值被误判为用户输入。
- callback 内再次修改状态导致重入。

验收：

- 每个 Ref 有状态机说明。
- 每个 `OnChange` 有触发条件说明。
- 程序行为和用户行为在事件里可区分。

### A10 组件族审查

目标：按控件族逐一审查，避免只修当前报错控件。

控件族和重点：

| 控件族 | 代表组件 | 必查底层能力 |
| --- | --- | --- |
| 基础交互 | Button、Pressable、ClickArea、IconButton、FAB | clickable、hover、pressed、focus、keyboard activation、ripple、disabled |
| 表单选择 | Checkbox、Switch、RadioGroup、Select | click default、keyboard、controlled value、OnChange、Ref、popup |
| 文本输入 | Input、TextField、SearchBar | editor、focus、keyboard、IME、wheel pass-through、programmatic source |
| 滚动集合 | ScrollView、ListView、Grid、Tabs、chips row | constraints、axis、wheel、hit-test refresh、virtualization、ScrollRef |
| 数值交互 | Slider、RangeSlider、Progress | pointer drag、capture、keyboard step、value clamp、visual state |
| Overlay | Dialog、Popup、Menu、Tooltip、Toast、Snackbar | portal、modal boundary、outside click、animation、focus trap |
| 拖放 | DragSource、DropTarget | Gio DnD、payload、pointer conflict、scroll conflict、drop default |
| 布局容器 | Container、Card、Stack、Row、Column、Padding | constraints、decoration、event area、state layer |
| 系统能力 | WindowDragArea、global shortcut、system events | OS boundary、component event boundary、platform fallback |

每个控件族都按同一模板输出：

- 公开 API 和旧 API。
- 内部状态。
- 依赖的 runtime 能力。
- 依赖的 layout 规则。
- 依赖的 event/focus/keyboard/scroll 规则。
- 样式和动画来源。
- Ref 行为。
- 自动测试覆盖。
- docs browser 或 example 手动验收。

### A11 示例和文档审查

目标：用真实示例验证 API，而不是只相信单元测试。

审查项：

- `examples/docs_browser` 是否覆盖核心控件真实组合。
- `examples/component_lab` 是否能暴露样式、cursor、hover 和布局问题。
- `examples/event_system_testbench` 是否说明每个事件功能怎么操作。
- `examples/drag_drop_showcase`、`examples/form_validation`、`examples/advanced_components` 是否仍符合当前 API。
- docs 中的 API 是否和代码一致。

验收：

- 每个核心控件至少有一个可手动验收示例。
- 每个复杂系统至少有一个组合示例。
- 文档不把 workaround 当成推荐用法。

### A12 性能、分配和诊断

目标：防止修复交互问题时引入每帧大量分配或全树重绘。

审查项：

- pointer move、wheel、hover、scroll 是否有分配热点。
- 大列表滚动时是否只处理可见区域。
- overlay 动画是否导致底层页面持续无意义重绘。
- diagnostics 是否能输出 event path、target、defaultPrevented、redraw reason。
- benchmark 是否覆盖高频输入和大组件树。

验收：

- 有 pointer/wheel/scroll benchmark。
- 高频事件不出现明显分配热点。
- 诊断输出能定位“谁注册了事件、谁取消了默认行为、谁触发了 redraw”。

### A13 冗余和收敛审查

目标：清理重复逻辑，降低后续修复风险。

审查项：

- 多个控件是否复制了 clickable、hover、pressed、disabled、ripple 逻辑。
- 多个 overlay 是否复制了 outside click、animation、portal 逻辑。
- 多个滚动容器是否复制了 offset、axis、wheel 逻辑。
- event default action 是否散落在组件内部。
- style 默认值是否散落在不同文件。

验收：

- 输出可合并 helper 列表。
- 输出暂不合并的理由，避免过度抽象。
- 每个 helper 的所有权明确。

## 控件关联审查模板

审查任意控件时使用以下模板：

| 维度 | 必答问题 |
| --- | --- |
| API | 公开 option、旧 callback、Ref、文档是否一致 |
| 状态 | controlled value、内部状态、路径状态、动画状态分别在哪里 |
| Layout | 输入 constraints、输出 size、命中区域是否等于视觉区域 |
| Render | 背景、边框、阴影、ripple、state layer 的绘制顺序 |
| Pointer | down/up/move/enter/leave/wheel 是否注册，是否扩大区域 |
| Keyboard | focus、keydown、keyup、shortcut、默认激活是否存在 |
| Text | 是否涉及 editor、IME、beforeinput、submit |
| Scroll | 是否截停父滚动，是否支持嵌套滚动，offset 是否正确 |
| Overlay | 是否涉及 portal、modal、outside click、z-order |
| Default action | 默认行为在哪里执行，是否可取消，取消后旧回调是否仍触发 |
| Event path | target、currentTarget、owner、portal path 是否清楚 |
| Ref | 命令何时消费，是否触发事件，卸载后是否安全 |
| Tests | 单元测试、交互测试、示例验收是否覆盖 |
| Docs | docs browser 文档和示例是否匹配真实行为 |

## 修复准入规则

审查后进入修复时，必须满足以下条件：

- 有明确问题编号。
- 有最小复现。
- 已确认受影响的控件族和底层能力。
- 已确认旧 API 兼容边界。
- 已确认修复不会改变无关控件的 layout、style 和 event 行为。
- 有自动测试或手动验收步骤。
- 修复只动一个底层能力或一个控件族，不能混改。

## 建议执行批次

### Batch 1：只审查不修复

- A0 工作区基线。
- A1 包边界和依赖方向。
- A2 Runtime、Context 和 Frame 生命周期。
- 产出第一版依赖图和风险地图。

### Batch 2：布局和视觉稳定性

- A3 Layout、尺寸和约束链路。
- A4 渲染、样式和动画链路。
- 优先确认 docs browser 和 component_lab 的布局/样式基线。

### Batch 3：输入和事件事实清单

- A5 事件系统和输入分发。
- A6 滚动、手势和嵌套交互。
- A7 Focus、Keyboard 和 Text Input。
- 只记录事实，不立即重写事件系统。

### Batch 4：复杂边界组件

- A8 Overlay、Portal、Modal 和 Popup 边界。
- A9 组件状态、Ref 和默认行为。
- A10 组件族审查。

### Batch 5：示例、性能和收敛

- A11 示例和文档审查。
- A12 性能、分配和诊断。
- A13 冗余和收敛审查。
- 输出后续修复路线图。

## 最终验收

整个审查完成时，应能回答以下问题：

- 每个核心控件依赖哪些底层能力。
- 每个底层能力影响哪些控件。
- 哪些逻辑是重复实现，哪些重复是有意保留。
- 哪些旧 API 必须保持兼容。
- 哪些问题应该在 runtime 修，哪些应该在 widget 修。
- 哪些示例是验收入口。
- 哪些测试能防止同类回归。
- 后续每个修复批次的风险边界在哪里。

审查完成后，才进入新的修复路线图。新的修复路线图必须从本审查产物中挑选问题，不能直接凭当前症状修改底层代码。
