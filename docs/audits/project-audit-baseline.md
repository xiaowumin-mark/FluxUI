<!-- fluxui-doc-meta
{
  "id": "project_audit_baseline",
  "title": "项目逻辑审查基线记录",
  "category": "工程审查",
  "order": 47,
  "summary": "记录 FluxUI 大项目逻辑审查 Batch 0~2 的基线事实，按任务拆分到独立子文件。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Context",
    "internal.Runtime",
    "ui.Element",
    "widget.Widget",
    "event.Event"
  ]
}
-->

# FluxUI 项目逻辑审查基线记录

本文件是审查基线的索引入口，各任务基线记录已拆分到独立子文件。

查看本文涉及的审查路线图和任务拆分，请参见：
- `docs/project-audit-roadmap.md`
- `docs/project-audit-task-breakdown.md`

---

## Batch 0：基线冻结

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A0.1 工作区状态冻结 | [audit-a0.1-workspace-freeze.md](audit-a0.1-workspace-freeze.md) | Done | 2026-07-06 |
| A0.2 版本和依赖快照 | [audit-a0.2-version-deps.md](audit-a0.2-version-deps.md) | Done | 2026-07-06 |
| A0.3 核心测试基线 | [audit-a0.3-core-tests.md](audit-a0.3-core-tests.md) | Done | 2026-07-06 |
| A0.4 示例 smoke 基线 | [audit-a0.4-example-smoke.md](audit-a0.4-example-smoke.md) | Done | 2026-07-06 |

## Batch 1：包边界和 runtime 基础

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A1.1 包依赖方向图 | [audit-a1.1-dependency-graph.md](audit-a1.1-dependency-graph.md) | Done | 2026-07-06 |
| A1.2 public API 所有权清单 | [audit-a1.2-public-api-ownership.md](audit-a1.2-public-api-ownership.md) | Done | 2026-07-06 |
| A1.3 escape hatch 边界审查 | [audit-a1.3-escape-hatch.md](audit-a1.3-escape-hatch.md) | Done | 2026-07-06 |
| A1.4 旧 API 兼容边界 | [audit-a1.4-legacy-api-compatibility.md](audit-a1.4-legacy-api-compatibility.md) | Done | 2026-07-06 |
| A2.1 Frame 生命周期图 | [audit-a2.1-frame-lifecycle.md](audit-a2.1-frame-lifecycle.md) | Done | 2026-07-06 |
| A2.2 PathID 和状态保存审查 | [audit-a2.2-pathid-state-retention.md](audit-a2.2-pathid-state-retention.md) | Done | 2026-07-06 |
| A2.3 Runtime registry 审查 | [audit-a2.3-runtime-registry.md](audit-a2.3-runtime-registry.md) | Done | 2026-07-06 |
| A2.4 redraw 和 invalidation 审查 | [audit-a2.4-redraw-invalidation.md](audit-a2.4-redraw-invalidation.md) | Done | 2026-07-06 |

## Batch 2：布局、渲染和样式稳定性

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A3.1 基础布局约束矩阵 | [audit-a3.1-basic-layout-constraints.md](audit-a3.1-basic-layout-constraints.md) | Done | 2026-07-06 |
| A3.2 滚动容器布局审查 | [audit-a3.2-scroll-container-layout.md](audit-a3.2-scroll-container-layout.md) | Done | 2026-07-06 |
| A3.3 横向内容溢出审查 | [audit-a3.3-horizontal-overflow.md](audit-a3.3-horizontal-overflow.md) | Done | 2026-07-06 |
| A3.4 hit area 和 layout area 对齐审查 | [audit-a3.4-hit-layout-area.md](audit-a3.4-hit-layout-area.md) | Done | 2026-07-06 |
| A4.1 Decoration 合并规则审查 | [audit-a4.1-decoration-merge-rules.md](audit-a4.1-decoration-merge-rules.md) | Done | 2026-07-06 |
| A4.2 交互态视觉来源审查 | [audit-a4.2-interaction-state-sources.md](audit-a4.2-interaction-state-sources.md) | Done | 2026-07-06 |
| A4.3 ripple 和 state layer 审查 | [audit-a4.3-ripple-state-layer.md](audit-a4.3-ripple-state-layer.md) | Done | 2026-07-06 |
| A4.4 cursor 策略审查 | [audit-a4.4-cursor-strategy.md](audit-a4.4-cursor-strategy.md) | Done | 2026-07-06 |
| A4.5 动画和布局隔离审查 | [audit-a4.5-animation-layout-isolation.md](audit-a4.5-animation-layout-isolation.md) | Done | 2026-07-06 |
| A5.1 Gio 原始输入映射表 | [audit-a5.1-gio-raw-input-mapping.md](audit-a5.1-gio-raw-input-mapping.md) | Done | 2026-07-06 |
| A5.2 EventTarget 分发顺序审查 | [audit-a5.2-eventtarget-dispatch-order.md](audit-a5.2-eventtarget-dispatch-order.md) | Done | 2026-07-06 |
| A5.3 旧事件桥接审查 | [audit-a5.3-legacy-event-bridge.md](audit-a5.3-legacy-event-bridge.md) | Done | 2026-07-06 |
| A5.4 PointerArea 影响范围审查 | [audit-a5.4-pointerarea-scope.md](audit-a5.4-pointerarea-scope.md) | Done | 2026-07-06 |
| A5.5 default action 可取消性矩阵 | [audit-a5.5-default-action-cancelability.md](audit-a5.5-default-action-cancelability.md) | Done | 2026-07-06 |
| A6.1 ScrollView 和 ListView offset 审查 | [audit-a6.1-scroll-offset.md](audit-a6.1-scroll-offset.md) | Done | 2026-07-06 |
| A6.2 wheel 分发和父子滚动审查 | [audit-a6.2-wheel-nested-scroll.md](audit-a6.2-wheel-nested-scroll.md) | Done | 2026-07-06 |
| A6.3 横向滚动策略审查 | [audit-a6.3-horizontal-scroll-strategy.md](audit-a6.3-horizontal-scroll-strategy.md) | Done | 2026-07-06 |
| A6.4 滚动后命中刷新审查 | [audit-a6.4-scroll-hit-refresh.md](audit-a6.4-scroll-hit-refresh.md) | Done | 2026-07-06 |
| A7.1 Focus target 注册审查 | [audit-a7.1-focus-target-registration.md](audit-a7.1-focus-target-registration.md) | Done | 2026-07-06 |
| A7.2 键盘事件和 shortcut 边界审查 | [audit-a7.2-keyboard-shortcut-boundary.md](audit-a7.2-keyboard-shortcut-boundary.md) | Done | 2026-07-06 |
| A7.3 键盘默认行为审查 | [audit-a7.3-keyboard-default-actions.md](audit-a7.3-keyboard-default-actions.md) | Done | 2026-07-07 |
| A7.4 文本输入事件审查 | [audit-a7.4-text-input-events.md](audit-a7.4-text-input-events.md) | Done | 2026-07-07 |
| A7.5 IME 和 submit 审查 | [audit-a7.5-ime-submit.md](audit-a7.5-ime-submit.md) | Done | 2026-07-07 |

## Batch 4：复杂边界和组件族

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A8.1 Overlay 挂载模型审查 | [audit-a8.1-overlay-mount-model.md](audit-a8.1-overlay-mount-model.md) | Done | 2026-07-07 |
| A8.2 Portal event path 审查 | [audit-a8.2-portal-event-path.md](audit-a8.2-portal-event-path.md) | Done | 2026-07-07 |
| A8.3 outside click 关闭规则审查 | [audit-a8.3-outside-click-close-rules.md](audit-a8.3-outside-click-close-rules.md) | Done | 2026-07-07 |
| A8.4 overlay focus 和 Escape 审查 | [audit-a8.4-overlay-focus-escape.md](audit-a8.4-overlay-focus-escape.md) | Done | 2026-07-07 |
| A9.1 Ref 命令生命周期审查 | [audit-a9.1-ref-command-lifecycle.md](audit-a9.1-ref-command-lifecycle.md) | Done | 2026-07-07 |
| A9.2 controlled value 和内部状态审查 | [audit-a9.2-controlled-value-internal-state.md](audit-a9.2-controlled-value-internal-state.md) | Done | 2026-07-07 |
| A9.3 OnChange 触发条件审查 | [audit-a9.3-onchange-trigger-conditions.md](audit-a9.3-onchange-trigger-conditions.md) | Done | 2026-07-07 |
| A10.1 基础交互控件族审查 | [audit-a10.1-basic-interactive-controls.md](audit-a10.1-basic-interactive-controls.md) | Done | 2026-07-07 |
| A10.2 表单选择控件族审查 | [audit-a10.2-form-selection-controls.md](audit-a10.2-form-selection-controls.md) | Done | 2026-07-07 |
| A10.3 文本输入控件族审查 | [audit-a10.3-text-input-controls.md](audit-a10.3-text-input-controls.md) | Done | 2026-07-07 |
| A10.4 滚动集合控件族审查 | [audit-a10.4-scroll-collection-controls.md](audit-a10.4-scroll-collection-controls.md) | Done | 2026-07-07 |
| A10.5 数值交互控件族审查 | [audit-a10.5-numeric-interaction-controls.md](audit-a10.5-numeric-interaction-controls.md) | Done | 2026-07-07 |
| A10.6 Overlay 控件族审查 | [audit-a10.6-overlay-controls.md](audit-a10.6-overlay-controls.md) | Done | 2026-07-07 |
| A10.7 拖放控件族审查 | [audit-a10.7-drag-drop-controls.md](audit-a10.7-drag-drop-controls.md) | Done | 2026-07-07 |
| A10.8 容器和装饰控件族审查 | [audit-a10.8-container-decoration-controls.md](audit-a10.8-container-decoration-controls.md) | Done | 2026-07-07 |

## Batch 5：示例、性能和冗余收敛

| 任务 | 文件 | 状态 | 日期 |
| --- | --- | --- | --- |
| A11.1 docs browser 组合场景审查 | [audit-a11.1-docs-browser-composition.md](audit-a11.1-docs-browser-composition.md) | Done | 2026-07-07 |
| A11.2 component_lab 审查 | [audit-a11.2-component-lab.md](audit-a11.2-component-lab.md) | Done | 2026-07-07 |
| A11.3 event_system_testbench 审查 | [audit-a11.3-event-system-testbench.md](audit-a11.3-event-system-testbench.md) | Done | 2026-07-07 |
| A11.4 示例覆盖矩阵 | [audit-a11.4-example-coverage-matrix.md](audit-a11.4-example-coverage-matrix.md) | Done | 2026-07-07 |
| A12.1 高频事件 benchmark 设计 | [audit-a12.1-high-frequency-event-benchmark.md](audit-a12.1-high-frequency-event-benchmark.md) | Done | 2026-07-07 |
| A12.2 大组件树 benchmark 设计 | [audit-a12.2-large-component-tree-benchmark.md](audit-a12.2-large-component-tree-benchmark.md) | Done | 2026-07-07 |
| A12.3 diagnostics 能力审查 | [audit-a12.3-diagnostics-capability.md](audit-a12.3-diagnostics-capability.md) | Done | 2026-07-07 |

---

## 关联能力总览

- 工作区状态冻结
- 审查产物记录
- Go 工具链复现
- Gio 运行时依赖复现
- 核心依赖版本冻结
- 核心包测试基线
- 既有失败识别
- 后续审查回归参照
- 示例 smoke 编译基线
- 包依赖方向
- runtime 所有权边界
- 反向依赖识别
- public API 所有权
- 旧 API 兼容边界
- runtime/event/widget 语义归属
- Gio context escape hatch
- Gio 原始事件桥接
- Widget 与 Element 兼容桥
- 事件系统迁移边界
- Runtime frame begin/end
- Context 根作用域和子作用域
- 每帧 event registry 重建
- 跨帧 PathID 和状态保存
- runtime frame 清扫边界
- PathID 身份规则
- Element keyed/unkeyed 状态保留
- Widget 列表位置状态风险
- Portal event parent 与 state path 边界
- Ref attach 命令目标边界
- Runtime event registry 生命周期
- 每帧 listener/focus/shortcut 重建
- pointer capture 跨帧 stale 风险
- Scroll wheel Gio tag 与 runtime registry 边界
- diagnostics current/last/pending 状态边界
- Runtime invalidator 与 Gio Window.Invalidate 绑定
- frame 内 op.InvalidateCmd 刷新
- redraw reason pending/current 语义
- 用户输入、动画、状态变更、程序命令 redraw 来源区分
- 动画 idle 停止与 redraw reason 基线
- 基础布局 constraints 输入/输出矩阵
- Row/Column Flex 有限和宽松约束语义
- Stack Stacked/Expanded 注释实现差异
- Center Gio Direction 约束边界
- Container/Padding inset 约束链路
- Fixed/Fill sizing 零 Max 和滚动上限风险
- ScrollView 主轴 content/viewport/offset 规则
- ListView/GridView 虚拟化滚动边界
- Grid 静态布局与 GridView 滚动语义区分
- ScrollRef 主轴命令与 offset clamp
- wheel 事件横纵轴过滤和嵌套滚动边界
- docs browser 代码块横向滚动边界
- Markdown 表格有限列宽策略
- chips row 横向 ScrollView 边界
- Tabs scrollable/fullWidth 宽度策略
- Menu/Select popup 宽度 clamp
- 普通 Row/Flex 非保护型横向布局边界
- PointerArea 子尺寸命中注册
- Pressable/ClickArea 无样式点击区域
- ContainerDecoration 视觉区域和 margin 命中边界
- Tabs item 等分/滚动命中边界
- Ripple 不改变 layout/hit size
- 空白区域不触发 hover target
- Decoration 字段级覆盖规则
- 交互态 Decoration 显式接管边界
- Material 3 默认样式 fallback
- disabled/hover/focus 视觉保护
- 状态 decoration 动画前清理
- 用户覆盖风险识别
- hover/pressed/focused 输入状态来源
- disabled/selected/error 组件语义状态来源
- Gio focus 与 runtime focus 语义边界
- Material state layer 状态消费
- focus indicator 状态消费
- 状态来源重复计算风险识别
- ripple primitive 绘制边界
- state layer 混色和圆形绘制边界
- ripple/state layer 绘制顺序
- hover/pressed 触发条件
- disabled 下交互反馈禁用
- touch target 与 layout size 区分
- component_lab cursor 异常入口定位
- Gio cursor op 设置边界
- Slider track cursor clip 策略
- disabled cursor 禁用边界
- cursor 清理和默认重置验证
- 单控件整窗 cursor 污染风险识别
- overlay 动画进度来源识别
- Dialog/Popup 普通布局与 overlay 布局隔离
- Toast/Snackbar 锚定 overlay 动画边界
- Select menu reveal 动画与 trigger 尺寸隔离
- Tabs indicator 视觉动画边界
- 动画 redraw 与布局尺寸关系识别
- Gio pointer 到 FluxUI PointerEvent/WheelEvent 字段映射
- Gio key 到 FluxUI KeyboardEvent 字段映射
- Gio Editor Change/Submit 到 FluxUI InputEvent best-effort 映射
- Gio transfer 到 FluxUI DragEvent/DropEvent 映射
- clipboard system API 与 event system 边界识别
- 不可用字段、best-effort 字段、后端依赖字段标注
- EventTarget capture/target/bubble 分发顺序
- listener priority、once、passive、stop、preventDefault 语义
- target path、portal、boundary 路径改写语义
- runtime keyboard default action 与 widget-local default action 分层
- Button/Pressable/ClickArea 旧 OnClick 与新 click 事件桥接顺序
- Input beforeinput/input/change 与旧 InputOnChange 桥接顺序
- ScrollView wheel default action 与旧 ScrollOnChange 触发顺序
- 旧 callback 防重复执行和可取消边界
- PointerArea Gio tag 注册区域与 child layout dimensions 对齐
- PointerAreaElement 声明式包装不额外扩大命中范围
- PointerArea pass-through、disabled、captureOnPress 影响范围区分
- event system testbench P2 Pointer/Wheel 手动验收入口
- PointerArea 坐标级 hit rect 自动测试缺口识别
- default action 可取消性矩阵
- PreventDefault 生效条件和 passive 限制
- click、keydown、wheel、beforeinput、dragover、drop 取消边界
- runtime keyboard default action 与 widget-local default action gate
- wheel/drop preventDefault widget 级测试缺口识别
- ScrollView offset 来源、单位、更新时机
- ScrollOnChange 旧回调触发和上报值边界
- 旧 ScrollOnChange 长期显示 0 的原因定位
- ScrollRef 命令消费与 offset clamp
- ListView Gio Position 和 reach-end 边界
- wheel target 注册区域和轴向过滤
- ScrollView wheel default action 和 PreventDefault gate
- 父子滚动命中优先级和横纵轴传递
- Input、Slider、DragSource wheel 截停边界
- Select/Menu/ListView 选项滚动入口
- docs content scroll 与代码块/表格嵌套滚动入口
- 横向 ScrollView 只消费 DeltaX 的策略
- shift-wheel 未定义/未实现边界
- touchpad 横向 delta 入口
- tabs/chips/table/code block 横向滚动验收入口
- 剩余 delta 未建模风险识别
- scroll 后 visual position 与 hit position 重建
- ScrollView/ListView 下一帧 ops/filter 命中刷新
- Select/Menu 下拉列表滚动命中入口
- pointer hover target 每帧重新收集边界
- 滚动后不移动鼠标点击新目标测试缺口
- runtime focus target 每帧注册和清理
- FocusManager facade 与 runtime focus 状态边界
- KeyboardScope focusable、disabled、autoFocus、tabIndex 规则
- Button/Select runtime focus target 注册路径
- Input Gio editor focus 与 FluxUI runtime focus 分离
- Dialog portal/boundary 与内部控件 focus 归属
- disabled/hidden/tabStop focus 判定规则
- keydown/keyup typed event 路由
- KeyboardScope Gio key event drain 边界
- OnShortcut 局部快捷键 scope 匹配和排序
- shortcut 与 runtime keyboard default action 顺序
- system global shortcut 与组件树 shortcut 分层边界
- Enter/Space runtime focus activation 默认行为
- Tab focus move 与 keyboard `PreventDefault` gate
- Button/Select/Menu/RadioGroup/Checkbox/Switch 键盘激活入口
- Escape 和 Arrow keys 默认行为缺口
- keyboard default 与合成 click default 双层可取消边界
- TextField beforeinput/input/change/submit 事件顺序
- InputRef SetText/Append/Clear programmatic source
- 用户输入、粘贴、删除、撤销/重做 best-effort 来源识别
- SearchBarInputOptions 高级文本事件转发边界
- composition synthetic 入口与真实 IME 自动桥接缺口
- IME compositionstart/update/end synthetic 顺序
- TextField Enter submit Gio SubmitEvent 边界
- InputEvent/KeyboardEvent IsComposing 未桥接风险
- runtime Enter 默认激活与 TextField submit 分层
- Dialog/Popup global-local mount API 与父约束边界
- Dialog portal stop boundary 与 Popup portal-only 差异
- Menu/DropdownMenu/Select 本地 op.Defer popup 挂载
- Tooltip child-local overlay 边界
- Toast/Snackbar anchored overlay 生命周期
- overlay z-order 依赖 Stack/op.Defer 顺序
- portal registration owner path 规则
- modal boundary event path 截断规则
- Popup portal-only 冒泡规则
- Dialog modal portal + stop boundary 规则
- Dialog/Popup maskClosable 遮罩关闭规则
- DropdownMenu/Select protected rect outside press 规则
- Menu 无独立 outside click 关闭边界
- 内部选择关闭与外部误关闭区分
- overlay 打开时 focus、关闭时 restore、Escape 捕获矩阵
- KeyboardScope 局部 Escape 处理入口
- Dialog/Popup modal 事件边界与 focus trap 缺口
- Menu/Select focus restore 配置入口和实现差异
- Escape 默认关闭行为缺口识别
- Ref 命令入队、redraw invalidator 和布局期消费规则
- ButtonRef/InputRef/ScrollRef/SliderRef/SelectRef/DialogRef/PopupRef 生效时机矩阵
- disabled/loading 下 Ref 命令 drain 后丢弃边界
- 未挂载/条件隐藏期间 Ref 命令保留和重新绑定消费风险
- Dialog/Popup/Select 受控入参与 Ref 命令覆盖边界
- Input/TextField controlled 判定依赖 `InputOnChange`
- Select/RadioGroup/Tabs 外部 value 与本帧 local current 值边界
- Checkbox/Switch 同帧多次 toggle 不累积风险
- Slider pressed 期间内部 progress 与外部 value 覆盖优先级
- SearchBar 继承 TextField 受控语义边界
- controlled value、内部 state、Ref/交互写入、OnChange 优先级矩阵
- InputRef programmatic mutation 与外部 value 重渲染回调差异
- Select/Tabs 当前项重复选择触发旧 OnChange 风险
- Checkbox/Switch 同帧多 toggle 重复 next 风险
- Slider 用户拖动和 Ref 设置 OnChange 触发差异
- ScrollView position diff 触发 ScrollOnChange 规则
- Button/Pressable/ClickArea 新旧 click 事件桥接边界
- Button 旧 OnHover 与 hover snapshot 变化过滤
- IconButton/FAB 直接旧 onClick 与新事件缺口
- 基础交互控件 pressed/focus/ripple/disabled 矩阵
- runtime focus target 与 Gio clickable focus 差异
- Checkbox/Switch click 和 keyboard 默认切换可取消性
- RadioGroup 当前项 no-op 与 Ref SetValue 边界
- Select trigger/option/popup/Ref 矩阵
- Select outside press 与 Ref 命令不可取消路径
- 表单选择控件 controlled value 与内部 open state 分层
- Input/TextField/SearchBar editor、focus、keyboard、IME、wheel pass-through 矩阵
- SearchBar 组合式 TextField 语义继承
- 输入控件不注册 FluxUI wheel listener 的父级滚动边界
- Gio Editor focus 与 runtime focus 分离在文本控件族中的验收入口
- IME composition 真实桥接缺口与 submit 风险
- ScrollView/ListView/Grid/GridView/Tabs/chips row 滚动集合矩阵
- ScrollView FluxUI wheel/default action/ScrollRef 职责边界
- ListView/GridView Gio list 虚拟化与 reach-end 边界
- TabsScrollable 继承横向 ScrollView 轴过滤和命中刷新
- chips row 组合责任与普通 Row 溢出风险
- Slider/RangeSlider pointer drag、capture、value clamp 矩阵
- Slider 不注册 wheel 的父级滚动 pass-through 边界
- Slider 键盘 step 缺口和 SliderRef StepBy 命令边界
- Progress 族只读 visual indicator 与 animation redraw 边界
- Dialog/Popup portal、modal、mask close 和 animation 边界
- Menu/DropdownMenu/Tooltip/Toast/Snackbar overlay 组件族矩阵
- DropdownMenu protected rect outside press 与普通 Menu 无 open 状态边界
- overlay focus trap、restore focus、Escape 默认关闭缺口
- DragSource/DropTarget payload、operation 和 Gio transfer 边界
- 拖放控件 hit area 与 child layout size 对齐
- 拖放组件不注册 wheel 的父级纵向滚动边界
- dragover/drop default action gate 与 payload 读取先后顺序
- Container/Card/Stack/Row/Column/Padding 组件族矩阵
- ContainerDecoration margin、padding、shadow 与事件命中边界
- Card bounded ripple overlay 与非 Button wrapper 边界
- 容器装饰 state decoration 不自动扩区的验收入口
- docs browser 左栏、右栏、搜索、分类 chips、代码块、表格和示例弹窗组合验收入口
- docs browser 横向区域与主文档纵向滚动互不污染的手动回归基线
- docs browser 示例弹窗打开、内部交互、遮罩关闭和重开行为验收表
- component_lab 主题、样式、cursor、hover、布局和 overlay smoke 入口
- component_lab 顶层虚拟化、perf diagnostics 和 idle redraw 自动基线
- component_lab CL-01 到 CL-17 视觉与交互手动验收表
- event_system_testbench P0-P7 操作说明、预期结果和复验点
- P1 EventTarget 分发、P3 Focus/Keyboard、P5 默认行为可取消性的手工判定口径
- event_system_testbench 历史问题与 R1-R5 修复记录的复验边界
- advanced_components、drag_drop_showcase、form_validation、horizontal_scroll、virtual_scroll 示例覆盖矩阵
- 示例到控件族、底层能力和核心能力真实入口的映射
- 示例 README 与 RunElement 入口实现漂移风险识别
- pointer move、wheel、hover、scroll 高频事件 benchmark 设计
- PointerArea 和 internal/perf 现有 benchmark 入口识别
- 高频输入 `allocs/op`、`input-ns/frame`、`hover-changes/frame` 预算口径
- ScrollView/ListView wheel 端到端 benchmark 缺口识别
- 大组件树 layout、paint、event registration benchmark 设计
- `internal/perf` 大树 benchmark 与 `FrameStats` 指标口径
- 大列表虚拟化 visible/culled 成本测量入口
- 大表格、真实 card grid、registration-only 对照 benchmark 缺口识别
- runtime/event/perf diagnostics 当前可见字段审查
- event path、target、defaultPrevented、redraw reason 输出边界
- 谁注册事件、谁取消默认行为、谁触发 redraw 的诊断字段缺口清单
