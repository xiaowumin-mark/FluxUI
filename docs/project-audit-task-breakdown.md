<!-- fluxui-doc-meta
{
  "id": "project_audit_task_breakdown",
  "title": "大项目逻辑审查任务拆分",
  "category": "工程路线图",
  "order": 46,
  "summary": "将大项目逻辑审查路线图拆成可执行工作包，按阶段、依赖、产出、工作量和验收标准推进，避免审查和修复混在一起。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Context",
    "internal.Runtime",
    "ui.Element",
    "widget.Widget",
    "event.Event",
    "widget.ScrollView",
    "widget.Input",
    "widget.Dialog",
    "widget.Popup"
  ]
}
-->

# FluxUI 大项目逻辑审查任务拆分

本文档是 `docs/project-audit-roadmap.md` 的执行拆分。主路线图回答“审什么”，本文档回答“按什么顺序审、每一步产出什么、任务量多大、什么时候能进入下一步”。

本拆分默认只做审查和记录，不直接改运行时代码。若审查过程中发现必须立即修复的问题，需要先登记为风险，再进入单独的修复路线图。

## 工作量口径

工作量用点数估算，便于后续转换为 issue、milestone 或 sprint。

| 规模 | 点数 | 参考耗时 | 含义 |
| --- | ---: | --- | --- |
| XS | 1 | 0.5 天以内 | 单文件或单一事实确认 |
| S | 2 | 0.5-1 天 | 小范围审查，有明确输入输出 |
| M | 3 | 1-2 天 | 跨 2-4 个文件或一个小系统 |
| L | 5 | 2-4 天 | 跨包审查，需要矩阵和复现 |
| XL | 8 | 4 天以上 | 范围过大，应继续拆分后再执行 |

统计规则：

- 审查点数不包含修复实现。
- 写测试或补示例只作为“验收设计”记录，不在本阶段实际实现。
- 如果一个任务需要修改运行时代码才能验证，应拆成“审查任务”和“后续修复任务”。
- 任意 XL 任务不得直接开工，必须拆成 L 或更小。

## 角色维度

任务可按以下维度分配，不要求真实团队有这些角色，但每个任务需要明确主要关注点。

| 维度 | 关注内容 |
| --- | --- |
| Runtime | `internal`、frame、context、tree、path、state、redraw |
| Event | `event`、Gio input、pointer、keyboard、focus、default action |
| Widget | `widget` 组件实现、Ref、状态、默认行为 |
| Layout | constraints、axis、scroll、overlay size、hit area |
| Style | theme、decoration、state layer、ripple、cursor、animation |
| Docs/Test | examples、docs browser、testbench、回归测试设计 |
| Perf | benchmark、allocation、redraw reason、diagnostics |

## 执行总览

| 批次 | 阶段 | 主要目标 | 点数 |
| --- | --- | --- | ---: |
| Batch 0 | A0 | 基线冻结和记录模板 | 8 |
| Batch 1 | A1-A2 | 包边界、runtime、frame、path | 27 |
| Batch 2 | A3-A4 | 布局、尺寸、渲染、样式、动画 | 35 |
| Batch 3 | A5-A7 | 事件、滚动、焦点、键盘、文本输入 | 52 |
| Batch 4 | A8-A10 | overlay、portal、Ref、组件族 | 59 |
| Batch 5 | A11-A13 | 示例、文档、性能、冗余收敛 | 36 |
| 合计 | A0-A13 | 完整审查，不含修复 | 217 |

建议执行方式：

- 单人执行：按批次顺序推进，每个批次结束后停下复盘。
- 多人执行：Batch 0 和 Batch 1 不并行，之后 Layout、Event、Widget、Docs/Test 可以并行，但必须共享同一份事实清单。
- 每个批次完成后只输出审查结论和后续修复候选，不直接进入修复。

## Batch 0：基线冻结

目标：确认当前仓库状态、测试状态和记录方式，避免审查过程中混入新变量。

### A0.1 工作区状态冻结

- 规模：XS，1 点。
- 关注：Runtime、Docs/Test。
- 输入：`git status --short`、当前分支、未跟踪文件。
- 输出：工作区状态记录，明确哪些文件允许继续保留。
- 验收：审查前的脏文件和新增文件都有解释。

### A0.2 版本和依赖快照

- 规模：S，2 点。
- 关注：Runtime。
- 输入：`go version`、`go env GOPATH GOMOD`、`go list -m all`。
- 输出：Go、Gio、核心依赖版本清单。
- 验收：后续问题可以复现到同一版本环境。

### A0.3 核心测试基线

- 规模：M，3 点。
- 关注：Docs/Test。
- 输入：`go test ./event ./internal ./ui ./widget`。
- 输出：通过/失败清单，失败项标注是否为既有问题。
- 验收：没有把既有失败误判为审查引入问题。

### A0.4 示例 smoke 基线

- 规模：S，2 点。
- 关注：Docs/Test。
- 输入：docs browser、component lab、event system testbench、drag drop showcase 的构建或空跑测试。
- 输出：示例可运行状态表。
- 验收：明确哪些示例是后续手动验收入口。

## Batch 1：包边界和 runtime 基础

目标：先审查运行时和包依赖，后续所有控件审查都依赖这个结果。

### A1.1 包依赖方向图

- 规模：M，3 点。
- 关注：Runtime。
- 输入：`go list -deps`、包源码 import。
- 输出：`internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system` 的依赖方向图。
- 验收：标出允许依赖、禁止依赖、疑似反向依赖。

### A1.2 public API 所有权清单

- 规模：M，3 点。
- 关注：Runtime、Event、Widget。
- 输入：`go doc` 或 `gopls go_package_api` 输出。
- 输出：公开 API 按包归属的清单。
- 验收：每个公开 API 都能说明由哪个包维护语义。

### A1.3 escape hatch 边界审查

- 规模：S，2 点。
- 关注：Runtime、Event。
- 输入：`ctx.Gtx`、`FromWidget`、直接 Gio event 使用点。
- 输出：escape hatch 使用清单。
- 验收：标出哪些是临时兼容，哪些是正式 API。

### A1.4 旧 API 兼容边界

- 规模：M，3 点。
- 关注：Widget、Event。
- 输入：`OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`、各 Ref。
- 输出：旧 API 兼容矩阵。
- 验收：后续修复不得删除或改变已承诺签名。

### A2.1 Frame 生命周期图

- 规模：L，5 点。
- 关注：Runtime。
- 输入：`internal/runtime.go`、`internal/frame.go`、`internal/context.go`。
- 输出：frame begin、layout、event registration、paint、frame end 的生命周期图。
- 验收：明确每帧哪些状态重建，哪些状态保留。

### A2.2 PathID 和状态保存审查

- 规模：L，5 点。
- 关注：Runtime、Widget。
- 输入：path memory、tree、component state、Ref attach。
- 输出：路径身份规则和错配风险清单。
- 验收：能解释列表重排、条件渲染、portal 下状态如何保持。

### A2.3 Runtime registry 审查

- 规模：M，3 点。
- 关注：Runtime、Event。
- 输入：event listener、focus target、pointer target、scroll target、diagnostics registry。
- 输出：每帧注册表的创建、更新、清理规则。
- 验收：没有上一帧目标无条件残留的风险。

### A2.4 redraw 和 invalidation 审查

- 规模：M，3 点。
- 关注：Runtime、Perf。
- 输入：invalidate、redraw reason、animation tick、interaction update。
- 输出：触发 redraw 的来源清单。
- 验收：能区分用户输入、动画、状态变更、程序命令触发的 redraw。

## Batch 2：布局、渲染和样式稳定性

目标：确认布局和视觉系统稳定，避免事件修复时破坏页面结构。

### A3.1 基础布局约束矩阵

- 规模：L，5 点。
- 关注：Layout。
- 输入：Row、Column、Stack、Center、Container、Padding、Fixed/Fill sizing。
- 输出：每个布局组件的 constraints 输入/输出矩阵。
- 验收：能解释每个组件在有限/无限约束下的尺寸行为。

### A3.2 滚动容器布局审查

- 规模：L，5 点。
- 关注：Layout、Widget。
- 输入：ScrollView、ListView、Grid、ScrollRef。
- 输出：主轴、交叉轴、content size、viewport size、offset 规则。
- 验收：横向滚动和纵向滚动的约束边界明确。

### A3.3 横向内容溢出审查

- 规模：M，3 点。
- 关注：Layout、Docs/Test。
- 输入：docs browser 代码块、表格、tabs、chips、menu row。
- 输出：横向溢出行为和最大宽度策略清单。
- 验收：不会因为子内容无限宽而造成滚动条异常。

### A3.4 hit area 和 layout area 对齐审查

- 规模：L，5 点。
- 关注：Layout、Event。
- 输入：PointerArea、Pressable、ClickArea、Container decoration、Tabs item。
- 输出：视觉区域、布局区域、事件命中区域对照表。
- 验收：没有组件把整行或整窗错误注册成命中区域。

### A4.1 Decoration 合并规则审查

- 规模：M，3 点。
- 关注：Style、Widget。
- 输入：style.Decoration、组件默认 decoration、用户覆盖选项。
- 输出：默认样式与用户样式的合并规则。
- 验收：用户自定义不会无意丢失必要的 disabled/hover/focus 表达。

### A4.2 交互态视觉来源审查

- 规模：M，3 点。
- 关注：Style、Event。
- 输入：hover、pressed、focused、disabled、selected、error。
- 输出：状态来源表。
- 验收：每个状态只有一个权威来源，避免 widget/event 各算一份。

### A4.3 ripple 和 state layer 审查

- 规模：S，2 点。
- 关注：Style、Widget。
- 输入：ripple、state layer 绘制点。
- 输出：绘制顺序和触发条件。
- 验收：ripple 不改变 layout，不扩大 hit area。

### A4.4 cursor 策略审查

- 规模：M，3 点。
- 关注：Style、Event。
- 输入：component_lab 中 cursor 异常相关组件。
- 输出：cursor 设置、清理、继承和重置规则。
- 验收：没有单个控件把整窗 cursor 污染成 pointer。

### A4.5 动画和布局隔离审查

- 规模：M，3 点。
- 关注：Style、Layout。
- 输入：Dialog、Popup、Toast、Snackbar、Select menu、Tabs indicator。
- 输出：动画影响范围表。
- 验收：动画只影响视觉或明确的 overlay 位置，不意外改变普通布局。

## Batch 3：事件、滚动、焦点和文本输入

目标：审清输入链路。这个批次风险最高，只审查不改动。

### A5.1 Gio 原始输入映射表

- 规模：M，3 点。
- 关注：Event。
- 输入：pointer、key、editor、clipboard、drag/drop 相关 Gio event。
- 输出：Gio 字段到 FluxUI event 字段的映射表。
- 验收：标出当前不可用、best-effort、后端依赖字段。

### A5.2 EventTarget 分发顺序审查

- 规模：L，5 点。
- 关注：Event、Runtime。
- 输入：event dispatcher、listener registry、target path。
- 输出：capture、target、bubble、default action 顺序图。
- 验收：能解释 listener 顺序、once、passive、stop、preventDefault。

### A5.3 旧事件桥接审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：Button、Pressable、ClickArea、Input、ScrollView。
- 输出：新事件与旧 callback 的桥接顺序。
- 验收：不会出现默认行为执行两次或回调漏执行。

### A5.4 PointerArea 影响范围审查

- 规模：M，3 点。
- 关注：Event、Layout。
- 输入：PointerArea、PointerAreaElement、event system testbench。
- 输出：PointerArea 注册区域和布局区域对照表。
- 验收：测试区域不会影响整行或兄弟控件。

### A5.5 default action 可取消性矩阵

- 规模：L，5 点。
- 关注：Event、Widget。
- 输入：click、keydown、wheel、beforeinput、dragover、drop。
- 输出：每个默认行为的 cancelable 规则。
- 验收：`PreventDefault` 生效条件和不生效条件写清楚。

### A6.1 ScrollView 和 ListView offset 审查

- 规模：M，3 点。
- 关注：Widget、Layout。
- 输入：ScrollView、ListView、ScrollOnChange、ScrollRef。
- 输出：offset 来源、单位、更新时机、回调时机。
- 验收：能解释为什么旧 `ScrollOnChange` 不应长期为 0。

### A6.2 wheel 分发和父子滚动审查

- 规模：L，5 点。
- 关注：Event、Layout。
- 输入：Input、DragSource、Slider、Select menu、docs content scroll。
- 输出：wheel target、可滚动判断、剩余 delta、父级传递规则。
- 验收：子控件不应无条件截停父页面滚动。

### A6.3 横向滚动策略审查

- 规模：M，3 点。
- 关注：Layout、Event。
- 输入：horizontal scroll、tabs、chips、table、code block。
- 输出：横向 delta、shift-wheel、touchpad、纵向 wheel 的处理策略。
- 验收：纵向滚动不会误触发横向滚动，除非策略明确允许。

### A6.4 滚动后命中刷新审查

- 规模：L，5 点。
- 关注：Runtime、Event、Widget。
- 输入：下拉菜单、ListView、ScrollView、pointer hover target。
- 输出：scroll 后 hit-test cache 更新规则。
- 验收：滚动后不移动鼠标也能点击当前鼠标下的新目标。

### A7.1 Focus target 注册审查

- 规模：M，3 点。
- 关注：Event、Runtime。
- 输入：FocusManager、KeyboardScope、Input、Button、Select、Dialog。
- 输出：focus target 注册、disabled、hidden、tab index 规则。
- 验收：focus path 和 target path 能对应。

### A7.2 键盘事件和 shortcut 边界审查

- 规模：L，5 点。
- 关注：Event、Widget。
- 输入：keydown、keyup、OnShortcut、global shortcut、KeyboardScope。
- 输出：局部快捷键和系统快捷键边界图。
- 验收：局部快捷键只在 scope 内触发。

### A7.3 键盘默认行为审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：Button、Select、Menu、Dialog、Tabs、RadioGroup、Checkbox、Switch。
- 输出：Enter、Space、Escape、Arrow keys 默认行为表。
- 验收：可取消默认行为和不可取消行为分清。

### A7.4 文本输入事件审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：Input、TextField、SearchBar、InputRef。
- 输出：beforeinput、input、change、submit、programmatic source 规则。
- 验收：用户输入、程序设置、粘贴、删除、撤销重做可区分。

### A7.5 IME 和 submit 审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：compositionstart/update/end、Enter submit。
- 输出：IME 组合期间事件顺序。
- 验收：组合中不会提前提交，组合结束后行为明确。

## Batch 4：复杂边界和组件族

目标：审查最容易造成跨示例混乱的 overlay、Ref 和核心组件族。

### A8.1 Overlay 挂载模型审查

- 规模：L，5 点。
- 关注：Runtime、Layout、Widget。
- 输入：Dialog、Popup、Menu、Select dropdown、Tooltip、Toast。
- 输出：global/local mount、z-order、layout parent、logical owner 表。
- 验收：每种 overlay 的布局边界和事件边界明确。

### A8.2 Portal event path 审查

- 规模：L，5 点。
- 关注：Event、Runtime。
- 输入：portal registration、owner target、modal boundary。
- 输出：overlay/portal/dialog/modal 的 event path 规则。
- 验收：Popup 能按规则冒泡到 owner，Dialog modal 能按规则截断。

### A8.3 outside click 关闭规则审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：DialogMaskClosable、PopupMaskClosable、Select、Menu。
- 输出：外部点击、内部点击、打开点击、遮罩点击规则表。
- 验收：内部按钮、空白、可滚动组件不会误关闭弹窗。

### A8.4 overlay focus 和 Escape 审查

- 规模：M，3 点。
- 关注：Event、Widget。
- 输入：Dialog、Popup、Menu、Select、KeyboardScope。
- 输出：打开时 focus、关闭时 restore、Escape 捕获规则。
- 验收：modal 和非 modal 的 focus 行为有差异且可解释。

### A9.1 Ref 命令生命周期审查

- 规模：L，5 点。
- 关注：Widget、Runtime。
- 输入：ButtonRef、InputRef、ScrollRef、SliderRef、DialogRef、PopupRef、SelectRef。
- 输出：命令入队、消费、失效、卸载安全规则。
- 验收：每个 Ref 何时生效可预测。

### A9.2 controlled value 和内部状态审查

- 规模：M，3 点。
- 关注：Widget。
- 输入：Input、Select、RadioGroup、Checkbox、Switch、Slider、Tabs。
- 输出：外部 value、内部 state、OnChange 的优先级表。
- 验收：不会同一帧互相覆盖。

### A9.3 OnChange 触发条件审查

- 规模：M，3 点。
- 关注：Widget、Event。
- 输入：表单控件、Slider、Tabs、ScrollView。
- 输出：每个 OnChange 是否只在真实变化时触发。
- 验收：程序设置和用户操作是否触发回调有明确规则。

### A10.1 基础交互控件族审查

- 规模：L，5 点。
- 关注：Widget、Event、Style。
- 范围：Button、Pressable、ClickArea、IconButton、FAB。
- 输出：click、hover、pressed、focus、ripple、disabled 矩阵。
- 验收：旧 OnClick/OnHover 和新事件边界清楚。

### A10.2 表单选择控件族审查

- 规模：L，5 点。
- 关注：Widget、Event。
- 范围：Checkbox、Switch、RadioGroup、Select。
- 输出：click、keyboard、controlled value、popup、Ref 矩阵。
- 验收：每个选择控件的默认行为可取消性明确。

### A10.3 文本输入控件族审查

- 规模：M，3 点。
- 关注：Widget、Event。
- 范围：Input、TextField、SearchBar。
- 输出：editor、focus、keyboard、IME、wheel pass-through 矩阵。
- 验收：输入控件不会无条件阻挡父级滚动。

### A10.4 滚动集合控件族审查

- 规模：L，5 点。
- 关注：Widget、Layout、Event。
- 范围：ScrollView、ListView、Grid、Tabs、chips row。
- 输出：axis、wheel、virtualization、hit refresh、ScrollRef 矩阵。
- 验收：集合控件不会互相污染滚动和命中测试。

### A10.5 数值交互控件族审查

- 规模：M，3 点。
- 关注：Widget、Event。
- 范围：Slider、RangeSlider、Progress。
- 输出：pointer drag、capture、keyboard step、value clamp 矩阵。
- 验收：Slider 竖向父滚动和横向拖拽边界清楚。

### A10.6 Overlay 控件族审查

- 规模：L，5 点。
- 关注：Widget、Layout、Event。
- 范围：Dialog、Popup、Menu、Tooltip、Toast、Snackbar。
- 输出：portal、modal、outside click、animation、focus 矩阵。
- 验收：每个 overlay 组件有统一边界策略。

### A10.7 拖放控件族审查

- 规模：M，3 点。
- 关注：Widget、Event。
- 范围：DragSource、DropTarget。
- 输出：payload、pointer conflict、scroll conflict、drop default 矩阵。
- 验收：拖拽组件不会截停无关纵向滚动。

### A10.8 容器和装饰控件族审查

- 规模：M，3 点。
- 关注：Layout、Style、Event。
- 范围：Container、Card、Stack、Row、Column、Padding。
- 输出：constraints、decoration、event area、state layer 矩阵。
- 验收：容器装饰不会意外扩大交互区域。

## Batch 5：示例、性能和冗余收敛

目标：把审查结论落到真实示例、性能预算和后续修复路线图。

### A11.1 docs browser 组合场景审查

- 规模：L，5 点。
- 关注：Docs/Test、Widget。
- 输入：docs browser 左栏、右栏、代码块、表格、示例弹窗、搜索 chips。
- 输出：docs browser 手动验收表。
- 验收：文档页面滚动、示例弹窗、横向区域、搜索交互都有覆盖。

### A11.2 component_lab 审查

- 规模：M，3 点。
- 关注：Docs/Test、Style。
- 输入：component_lab 主要页面。
- 输出：样式、cursor、hover、布局异常清单。
- 验收：能作为视觉和交互 smoke 入口。

### A11.3 event_system_testbench 审查

- 规模：M，3 点。
- 关注：Docs/Test、Event。
- 输入：P0-P7 测试台。
- 输出：每个测试项的操作说明、预期结果和当前问题清单。
- 验收：用户能看懂 P1/P3/P5 等测试项在测什么。

### A11.4 示例覆盖矩阵

- 规模：M，3 点。
- 关注：Docs/Test。
- 输入：advanced_components、drag_drop_showcase、form_validation、horizontal_scroll、virtual_scroll。
- 输出：示例到控件族和底层能力的覆盖矩阵。
- 验收：每个核心能力至少有一个真实示例入口。

### A12.1 高频事件 benchmark 设计

- 规模：M，3 点。
- 关注：Perf、Event。
- 输入：pointer move、wheel、hover、scroll。
- 输出：benchmark 场景设计和预算。
- 验收：能衡量高频输入是否产生明显分配热点。

### A12.2 大组件树 benchmark 设计

- 规模：M，3 点。
- 关注：Perf、Layout。
- 输入：大列表、大表格、多卡片、多交互控件。
- 输出：layout、paint、event registration 成本测量方案。
- 验收：能区分 layout 成本和 event 成本。

### A12.3 diagnostics 能力审查

- 规模：L，5 点。
- 关注：Runtime、Event、Perf。
- 输入：interaction diagnostics、redraw reason、event path。
- 输出：诊断字段缺口清单。
- 验收：后续能定位谁注册事件、谁取消默认行为、谁触发 redraw。

### A13.1 clickable 和交互态冗余审查

- 规模：M，3 点。
- 关注：Widget、Event。
- 输入：Button、Pressable、ClickArea、Checkbox、Switch、Tabs、Select option。
- 输出：重复 hover/pressed/focused/ripple 逻辑清单。
- 验收：标出应该收敛到 helper 的部分和暂不收敛的理由。

### A13.2 overlay 逻辑冗余审查

- 规模：M，3 点。
- 关注：Widget。
- 输入：Dialog、Popup、Menu、Select、Tooltip、Toast。
- 输出：outside click、portal、animation、focus restore 重复逻辑清单。
- 验收：统一 overlay 基础设施的候选边界明确。

### A13.3 滚动和 hit-test 冗余审查

- 规模：M，3 点。
- 关注：Layout、Event。
- 输入：ScrollView、ListView、Grid、Select menu、tabs horizontal scroll。
- 输出：offset、wheel、axis、hit refresh 重复逻辑清单。
- 验收：后续可以围绕一个底层滚动模型规划修复。

### A13.4 最终修复候选排序

- 规模：L，5 点。
- 关注：全部。
- 输入：所有审查产物。
- 输出：按风险和影响面排序的修复候选列表。
- 验收：每个修复候选都有影响范围、依赖任务、验收示例和回滚策略。

## 阶段关卡

每个批次结束必须通过关卡，才能进入下一批次。

| 关卡 | 条件 |
| --- | --- |
| G0 | 基线可复现，测试和示例状态已记录 |
| G1 | 包边界、runtime 生命周期、PathID 规则已记录 |
| G2 | 布局、命中区域、样式状态来源已记录 |
| G3 | 事件、滚动、焦点、键盘、文本输入事实清单已记录 |
| G4 | overlay、Ref、组件族矩阵已记录 |
| G5 | 示例覆盖、性能预算、冗余清单和修复候选已记录 |

关卡不要求没有问题，而是要求问题已经被定位到明确的功能链路。

## 并行策略

可以并行：

- A3 布局审查和 A4 样式审查。
- A5 事件审查和 A7 文本输入审查，但两者必须共享 default action 规则。
- A11 示例审查和 A12 benchmark 设计。
- A13 冗余审查可以在对应控件族审查完成后局部开始。

不建议并行：

- A1 和 A2。包边界和 runtime 生命周期是后续所有审查的基础。
- A6 滚动审查和 A8 overlay 审查。下拉菜单、弹窗滚动、命中刷新有交叉，先完成 A6 更稳。
- A10 组件族审查和修复实现。组件族审查阶段不应夹带改代码。

## 任务记录模板

每个任务完成时建议用以下格式记录：

```md
### 任务编号：A6.2 wheel 分发和父子滚动审查

- 状态：Done
- 日期：
- 负责人：
- 输入文件：
- 关联能力：
- 事实结论：
- 风险：
- 冗余：
- 修复候选：
- 验收入口：
- 后续依赖：
```

## 下一步建议

从 Batch 0 开始执行，不要直接进入事件或滚动修复。建议第一轮实际产物为：

- `docs/audits/project-audit-baseline.md`
- `docs/audits/project-audit-dependency-map.md`
- `docs/audits/project-audit-runtime-lifecycle.md`

这三份产物完成后，再决定是否进入 Batch 2。这样可以在真正改代码前先建立稳定的事实基线。
