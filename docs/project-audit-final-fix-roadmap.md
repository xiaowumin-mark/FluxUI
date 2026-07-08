<!-- fluxui-doc-meta
{
  "id": "project_audit_final_fix_roadmap",
  "title": "项目审查后最终修复路线图",
  "category": "工程路线图",
  "order": 48,
  "summary": "基于项目逻辑审查 A0-A13 的结果，给出 FluxUI 后续修复、规范化、回归验证和发布准入的实施路线图。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Runtime",
    "event.Event",
    "widget.ScrollView",
    "widget.Dialog",
    "widget.Input",
    "widget.Button"
  ]
}
-->

# FluxUI 项目审查后最终修复路线图

本文档是 `docs/project-audit-roadmap.md` 和 `docs/audits/audit-a13.4-final-fix-candidate-ranking.md` 之后的实施路线图。审查阶段已经完成事实、风险和候选排序；本路线图用于约束后续真正修改代码时的顺序、范围、验收和回滚策略。

## 总原则

- 先补可观测性和回归入口，再改跨层行为。
- 每个修复必须绑定至少一个审查产物、一个自动测试或手动 smoke 入口、一个回滚策略。
- 不允许以重构名义改变旧 API 签名或旧回调顺序，尤其是 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange` 和各类 `Ref`。
- 不允许一次 PR 同时修改滚动、overlay、focus、default action 和受控状态。跨系统修复必须拆成 staged PR。
- 每个功能完成后必须执行 `FEATURE_INTEGRATION_CHECKLIST.md` 中的关联性检查。

## 阶段路线

| 阶段 | 目标 | 对应候选 | 主要产物 | 退出条件 |
| --- | --- | --- | --- | --- |
| P0 | 修复准备与规范冻结 | F-08 | 本路线图、根目录开发规范、回归入口索引 | 规范文件存在，后续 PR 可按同一口径评审 |
| P1 | diagnostics 和回归底座 | F-05、F-08 | registry/event/redraw 最小诊断字段，高风险场景 smoke 表 | 默认关闭路径无明显性能回退，能定位事件取消和 redraw 来源 |
| P2 | 滚动与命中核心修复 | F-01 | `wheelPolicy`、scroll hit refresh 回归、ScrollView/ListView 边界适配 | 横向/纵向/嵌套滚动行为稳定，scroll 后命中新目标 |
| P3 | overlay 与 focus 规则修复 | F-02 | overlay close/focus/Escape 策略，topmost overlay 规则 | modal 与非 modal 差异清楚，内部点击/滚动不误关闭 |
| P4 | clickable 和 default action 收敛 | F-03 | click/focus/default action 窄 helper，旧回调桥接测试 | click、keyboard activation、`PreventDefault` 顺序一致 |
| P5 | 状态、Ref、OnChange、文本输入修复 | F-04、F-06 | 受控值优先级、Ref drain、IME/source/submit 规则 | 程序设置、用户输入和 Ref 命令同帧不互相覆盖 |
| P6 | layout/hit/style 收尾 | F-07 | hit area 测试、state layer/ripple/cursor 约束 | 视觉、布局、命中区域一致，无整窗 cursor 污染 |
| P7 | 示例、benchmark 和文档收敛 | F-08 | docs browser、component_lab、testbench、examples 回归矩阵 | 核心能力都有真实入口，文档与实现一致 |

## P1：diagnostics 和回归底座

### 范围

- Runtime registry：event target、listener、focus target、shortcut 的每帧数量。
- Event dispatch：last target/path、default prevented、stop propagation、portal/boundary 改写原因。
- Redraw reason：reason、source、owner path 的结构化记录。
- Perf：高频 pointer/wheel 默认关闭时不新增明显分配。

### 验收

- `event_system_testbench` 能辅助定位 P1/P3/P5 的事件路径和取消结果。
- `component_lab` 能观察 idle redraw、cursor、hover 和 overlay smoke。
- `go test ./event ./internal ./ui ./widget` 通过，若失败必须标注既有问题。

### 关联性

- 任何 diagnostics 字段不得成为组件逻辑依赖。
- 诊断开关默认关闭，禁止在默认路径保留大容量 event history。

### P1 实施记录（2026-07-07）

- 审查依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A2.3、A2.4、A5.2、A5.5、A7.1、A7.2、A8.2、A11.2、A11.3、A12.1、A12.3、A13.4 的事件、diagnostics、组件入口和风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 代码产物：`internal.FrameStats` 增加 registry 计数、last dispatch 取消/停止/路径改写字段和结构化 redraw reason；`Event` 内部记录 PreventDefault/StopPropagation 调用 target 与 phase；`Context.RequestRedrawReason`、`WindowInvalidate` 和 widget redraw invalidator 记录 reason、source、owner path。
- 回归入口：`event/event_test.go` 覆盖 registry 数量、last target/path、default prevented、stop propagation、passive prevent、portal/boundary path rewrite 和 redraw owner path；`examples/event_system_testbench` 的诊断面板指向 P1/P3/P5/P6 事件路径和取消结果；`examples/component_lab` 保留 idle redraw、slider cursor、hover/pressed 与 overlay smoke 入口。
- 验证：`go test ./event` 通过；`go test ./event ./internal ./ui ./widget` 通过；gopls diagnostics 对本轮修改文件无报错。
- 关联性约束：新增字段只由 diagnostics/perf 统计写入和测试读取，组件逻辑不依赖 diagnostics；诊断仍由 `EnablePerfDiagnostics` / `WithPerfDiagnostics` 默认关闭控制，默认路径不保留大容量 event history。
- 回滚策略：若 P2-P6 后续接入暴露性能或行为回退，可先回退 structured last-dispatch/redraw 字段和 testbench 文案，保留不改变行为的 registry 计数；若 registry 计数成本异常，可仅在 diagnostics enabled 时延迟统计。

## P2：滚动与命中核心修复

### 范围

- 统一 wheel 主轴、交叉轴、touchpad 双轴、shift-wheel 和剩余 delta 策略。
- 明确 `wheel.PreventDefault()` 只阻止对应 wheel default action，不影响 Ref、scrollbar、auto-to-end 等非 wheel 来源。
- 明确 `ScrollView` pixel offset 与 `ListView/GridView` virtual position 的适配关系。
- 建立 scroll 后 hit refresh 合同：滚动后不移动鼠标，当前坐标点击和 hover 应命中新视觉目标。

### 验收

- `examples/horizontal_scroll`：横向内容只消费明确横向输入，纵向滚动不误触横向滚动。
- `examples/virtual_scroll`：大列表/大网格滚动后 item 序号、命中和可见范围稳定。
- docs browser：主文档纵向滚动、代码块/表格横向区域、tabs/chips 横向区域互不污染。
- Select menu：内部滚动后 option hit 和 outside protected rect 同步更新。

### 回滚

- 保留 `ScrollView` 原 pixel-offset 路径和 `ListView/GridView` Gio list 路径。
- 新策略先通过内部 adapter 接入，异常时按组件回退。

### P2 实施记录（2026-07-07）

- 审查依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A3.2、A3.3、A3.4、A5.5、A6.1、A6.2、A6.3、A6.4、A8.3、A10.4、A11.1、A11.4、A12.1、A13.3、A13.4 的滚动、命中、default action、docs browser 和示例风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 代码产物：`ScrollView` wheel 处理收敛到内部 `wheelPolicy` adapter，统一 `pointer.Filter`、主轴 delta 判定和 default action gate；横向容器仍只接收 Gio `Scroll.X`，纵向容器仍只接收 Gio `Scroll.Y`，避免横向代码块、表格、tabs、chips 误吞普通纵向滚动；`wheel.PreventDefault()` 只阻止本次 wheel default scroll，不影响 `ScrollRef`、scrollbar、auto-to-end 等非 wheel 位置来源。
- 模型边界：保留 `ScrollView` pixel-offset 路径和 `ListView/GridView` Gio `layout.List.Position` 路径，不把 `ScrollRef` 或 `ScrollOnChange` 扩展到虚拟列表；同轴触边后的父级剩余 delta chaining 仍不在本轮建模，当前策略是主轴内 clamp、交叉轴通过严格 filter 透传给父级。
- shift-wheel / touchpad 策略：本轮不把 `Shift+DeltaY` 强转为横向滚动，因为 Gio `pointer.Filter` 当前不能按 modifier 条件接收 wheel；强行打开横向容器的 Y filter 会让 `TestVerticalWheelOverHorizontalScrollViewScrollsOuter` 回退。触控板横向能力以 Gio 后端明确提供 `Scroll.X` 为准，双轴输入中的交叉轴不由 `ScrollView` 转换为另一轴。
- 回归入口：`widget/list_grid_test.go` 新增 `TestScrollViewWheelPreventDefaultBlocksOnlyWheelDefaultAction`、`TestVerticalScrollViewClickAfterScrollUsesUpdatedVisualPosition`、`TestListViewClickAfterScrollUsesUpdatedVisibleItem`；既有 `TestHorizontalScrollViewRequiresHorizontalWheelDelta`、`TestVerticalWheelOverHorizontalScrollViewScrollsOuter`、嵌套滚动、虚拟化测试继续覆盖横向/纵向互不污染和 Gio list 可见窗口。
- 验证：定向 `go test ./widget -run "Test(ScrollViewWheel|HorizontalScrollView|VerticalScrollView|VerticalWheelOverHorizontal|NestedScroll|ListViewWheel|ListViewClickAfterScroll|ListViewVirtualization|GridViewVirtualization)"` 通过；`go test ./event ./internal ./ui ./widget` 通过；`go test ./examples/horizontal_scroll ./examples/virtual_scroll ./examples/docs_browser` 通过；gopls diagnostics 对本轮修改文件无报错。
- 关联性检查：未新增跨包依赖，未改变公开 `ScrollOption`、`ScrollRef`、`ScrollOnChange` 签名和回调顺序；diagnostics 默认路径未增加事件历史；Select/DropdownMenu protected rect 仍由当前帧 field/popup rect 计算，内部 ListView 滚动命中由新增 Gio list click-after-scroll 回归覆盖基础路径，outside close/focus 仍留在 P3。
- 回滚策略：如 `wheelPolicy` adapter 引发平台 wheel 路由异常，可回退 `processWheelEvents` / `applyWheelDefault` 到原手写轴向 filter；新增测试保留为回归口径。若未来要支持 modifier-aware shift-wheel 或同轴剩余 delta chaining，应先补新的 Gio 路由能力验证和父子 scroll chain 测试，再按组件接入。

## P3：overlay 与 focus 规则修复

### 范围

- Dialog：modal boundary、mask click、focus trap、Escape、restore focus。
- Popup：portal-only 或可配置 boundary，不默认变成 Dialog 语义。
- Select/DropdownMenu：trigger/popup protected rect、内部滚动、outside press、restore focus。
- Tooltip/Toast/Snackbar：不参与 modal close/focus trap，只保留 hover/timed overlay 语义。

### 验收

- Dialog modal 中 Tab 不越出面板，Escape 只关闭 topmost overlay。
- Popup 仍可按 portal-only 规则冒泡到 owner。
- Select/DropdownMenu 内部按钮、空白、滚动区域不误关闭；外部点击按规则关闭。
- 关闭 overlay 后 focus 回到 trigger、field 或明确清空。

### 回滚

- focus trap、Escape close、restore focus 分开接入。
- outside click 保留原 mask/protected rect 路径，统一 helper 只包裹判定，不吞掉组件差异。

### P3 实施记录（2026-07-07）

- 审查依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A7.1、A7.2、A7.3、A8.1、A8.2、A8.3、A8.4、A10.2、A10.6、A13.2、A13.4 的 overlay mount、portal path、outside press、focus target、keyboard default、Escape 和最终候选风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 代码产物：`internal.Runtime` 增加 `MoveFocusWithin`，按当前帧 tab order 和 event path 将焦点限制在指定 scope 内；`Dialog` 在 portal boundary 内接入 Tab focus trap、Escape cancel close、打开前 focus 记录和关闭后 restore/clear；新增 `widget/overlay_focus.go` 作为 overlay 键盘事件与 focus 小 helper；`Select` 和 `DropdownMenu` 在保留原 trigger/popup protected rect outside press 路径的前提下补充 Escape close、trigger/field focus restore 和 outside dismiss 后的明确清空。
- 语义边界：`Popup` 仍只注册 portal 到 owner，不默认注册 `BoundaryStopPropagation`、focus trap 或 Escape close；`Tooltip`、`Toast`、`Snackbar` 未接入 modal close/focus trap，继续保持 hover/timed overlay 语义。
- 回归入口：`event/keyboard_test.go` 新增 scope 内 focus wrap/进入测试；`widget/interactive_layout_test.go` 新增 Dialog Tab/Shift+Tab/Escape/restore focus、Select Escape restore field、DropdownMenu Escape restore trigger 测试；既有 modal 内部 press 与 outside mask close 测试继续覆盖 Dialog/Popup 指针路径。
- 验证：`go test ./event ./widget` 通过；`go test ./event ./internal ./ui ./widget` 通过；gopls diagnostics 对本轮修改文件无报错。`go_vulncheck ./...` 仅报告当前 Go 1.25.1 标准库及间接图像/系统包的既有漏洞项，本轮未新增或更新依赖。
- 关联性检查：未改变公开 `DialogOption`、`PopupOption`、`SelectOption`、`DropdownMenuOption` 签名；Dialog 的 mask click 仍走原 `fillWidget` 路径，Select/DropdownMenu outside press 仍走 `md3DismissOnOutsidePress` 和 protected rect 判定；新增 helper 只处理键盘 close/focus，不吞掉组件差异；Input/Gio editor focus、IME、ScrollView/ListView 滚动策略不在本轮扩展。
- 回滚策略：如 Dialog focus trap 异常，可先移除 portal `OnKeyDown` 的 Tab 分支并保留 restore focus；如 Escape close 异常，可单独移除 `md3RegisterEscapeClose` / Dialog Escape 分支；如 Select/DropdownMenu outside dismiss 与 focus 恢复冲突，可回退各自的 `RequestFocus` / `md3ClearFocusIfInside` 调用，保留原 protected rect outside press 判定。

## P4：clickable 和 default action 收敛

### 范围

- `drainClickableDefaultAction`：统一 click dispatch、disabled gate、旧回调/default action 顺序。
- `registerClickableFocusAction`：统一 Enter/Space keyboard activation。
- `clickableInteractionSnapshot`：统一 hover/pressed/focused 读取，减少重复状态计算。

### 验收

- Button、Pressable、ClickArea、Checkbox、Switch、RadioGroup、Select option 的 click 顺序一致。
- `PreventDefault` 能阻止状态切换或旧回调，但 disabled/loading 属于组件入口提前阻断。
- Tabs 若接入 helper，不改变已有 active 切换语义和 indicator 布局。

### 回滚

- 不合并 Button surface、Switch thumb/track、Tabs indicator、Select field visual。
- 单组件接入失败时回退原手写 loop。

### P4 实施记录（2026-07-07）

- 审查依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A1.4、A4.2、A5.2、A5.3、A5.5、A7.3、A9.2、A9.3、A10.1、A10.2、A13.1、A13.4 的 click/default action、keyboard activation、旧回调桥接、Ref、Select/Tabs 风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 代码产物：`widget/event_defaults.go` 增加 `drainClickableDefaultAction`、`registerClickableFocusAction`、`clickableInteractionSnapshot` 和内部 `runClickableDefaultAction`，保留 `dispatchClickDefault` 兼容旧调用；Button、Pressable、ClickArea、Checkbox、Switch、RadioGroup、Select trigger/option 的 pointer click、Enter/Space keyboard activation 和 interaction snapshot 逐步接入 helper。
- 语义边界：disabled/loading 仍在组件入口提前阻断，不通过 `PreventDefault` 表达；旧 `OnClick`、`OnChange`、状态切换和默认行为仍只在 cancelable click 被允许后执行；Button surface、Switch thumb/track、Tabs indicator、Select field visual 未合并到通用 helper；IconButton/FAB 不在本阶段验收清单内，未扩大范围。
- Tabs 边界：Tabs pointer click 接入 `drainClickableDefaultAction`，让 click listener 的 `PreventDefault` 能阻止 active 切换；未新增 Tabs focus target 或 Enter/Space activation，避免改变 A7.3 已记录的键盘语义缺口和既有 indicator 布局。
- 回归入口：`widget/interactive_layout_test.go` 新增 `TestClickablePreventDefaultBlocksPointerDefaultActions`，覆盖 Button、Pressable、Checkbox、Switch、RadioGroup、Select trigger、Select option、Tabs 的 `PreventDefault` 反例；新增 `TestTabsPointerClickStillActivatesWithoutPreventDefault` 和 `TestPressableKeyboardActivationUsesCancelableClick`，覆盖 Tabs 正常指针切换和 Pressable keyboard activation 的 cancelable click 路径。
- 验证：定向 `go test ./widget -run "Test(ClickablePreventDefaultBlocksPointerDefaultActions|TabsPointerClickStillActivatesWithoutPreventDefault|PressableKeyboardActivationUsesCancelableClick)$"` 通过；`go test ./widget ./event` 通过；`go test ./internal ./ui` 通过；`go test ./examples/event_system_testbench ./examples/component_lab ./examples/docs_browser` 通过；`go test ./examples/basic_components ./examples/material3_showcase` 通过；gopls diagnostics 对本轮修改文件无报错。
- 关联性检查：未改变公开 option/ref 签名，未新增跨包依赖，未改 Ref 调用路径的 cancelable 语义；Select/DropdownMenu outside close、focus restore、scroll 策略仍沿用 P3/P2 结果；event diagnostics 默认路径和 layout/hit size 不因 helper 接入扩大。
- 回滚策略：如单个组件接入 helper 后暴露顺序回退，可按 Button、Pressable/ClickArea、Checkbox/Switch、RadioGroup、Select、Tabs 的粒度回退到原手写 click loop；保留新增测试作为顺序合同。若 Tabs active 切换出现兼容问题，可先仅移除 Tabs 的 helper 接入，不影响其他 clickable 组件。

## P5：状态、Ref、OnChange、文本输入修复

### 范围

- controlled value、内部 state、Ref 命令、用户交互写入的优先级。
- `OnChange` 只在真实变化时触发，重复选择当前项不应触发真实变化回调。
- IME composition 期间不提前 submit，composition 状态和 Enter submit 规则清楚。
- `beforeinput`、`input`、`change`、`submit`、programmatic source 可区分。

### 验收

- `examples/form_validation`：快捷填充、用户编辑、提交验证一致。
- Input/TextField/SearchBar：用户输入、粘贴、删除、Ref SetText 的 source 明确。
- Slider/Checkbox/Switch/Tabs：同帧多次更新不互相覆盖。

### 回滚

- 先补 diff/gate 和测试，不改变公开 option/ref 签名。
- 真实 IME 桥接若不稳定，先保留 best-effort 并在文档中明确限制。

### P5 实施记录（2026-07-07）

- 审计依据：复核 `audit-a7.4-text-input-events.md`、`audit-a7.5-ime-submit.md`、`audit-a9.1-ref-command-lifecycle.md`、`audit-a9.2-controlled-value-internal-state.md`、`audit-a9.3-onchange-trigger-conditions.md`、`audit-a10.2-form-selection-controls.md`、`audit-a10.3-text-input-controls.md`、`audit-a10.5-numeric-interaction-controls.md`、`audit-a13.4-final-fix-candidate-ranking.md` 与 `audit-a11.4-example-coverage-matrix.md`，并按组件开发指南、代码风格指南和功能集成检查表约束执行。
- 代码落点：`widget/input.go` 为 programmatic/user mutation 增加真实 diff gate，同值 `InputRef.SetText` 和异常同值 user change 不再派发文本事件；过滤后未变值只保留可取消的 `beforeinput` 尝试，不再派发 `input`/`change`/`InputOnChange`；受 Gio 公共 IME API 限制，真实 IME composition 仍按 A7.5 记录为 best-effort，`Submit` 不伪造 composing 状态。
- 状态与 Ref：`widget/checkbox.go`、`widget/switch.go` 改为以本帧 `currentValue` 串行消费 Ref/用户默认动作，`Toggle`/`SetChecked` 同帧多命令按顺序累计并只对真实变化调用 `OnChange`；`widget/slider.go` 保持现有 Ref 顺序消费且 Ref 不作为 `OnChange` 来源，通过回归测试锁定。
- 选择控件：`widget/selection.go` 将 Select option 默认动作收敛到内部 `activateOptionValue`，重复选择当前值不触发 `SelectOnChange`，但仍保留关闭弹层/恢复焦点默认动作；`widget/tabs_dialog_toast.go` 对当前 active tab 点击增加 same-key gate，不再重复触发 `TabsOnChange`。
- 回归覆盖：新增 `TestInputRefSameValueDoesNotDispatchProgrammaticChange`、`TestSelectionControlsDoNotReportUnchangedCurrentItem`、`TestSameFrameRefCommandsAccumulate`，覆盖 Input/TextField programmatic source、Select/Tabs 当前项、Checkbox/Switch/Slider 同帧 Ref 命令和 Slider Ref 非 `OnChange` 来源。
- 验证：`gofmt` 已执行；`gopls go_diagnostics` 对修改文件无诊断；`go test ./widget` 通过；`go test ./examples/form_validation` 构建通过（该目录无测试文件）。本轮未修改 `go.mod`；`go_vulncheck ./...` 仍报告现有 Go 1.25.1 标准库与间接依赖风险，未引入新的依赖风险。
- 集成检查：未改变公开 option/ref 签名；保持 controlled value 外部回灌模型，Ref 只在组件布局帧内消费；事件默认动作仍经过 P4 click gate；SearchBar 继续复用 TextField/inputWidget 的 source 与 diff 规则。
- 回滚边界：如 programmatic 文本事件 gate 需要回退，可仅撤销 `input.go` diff gate 与对应测试；如选择/布尔控件同帧累计策略需回退，可按 `selection.go`、`tabs_dialog_toast.go`、`checkbox.go`、`switch.go` 分组件退回，不需要整体 revert。

## P6：layout、hit、style 收尾

### 范围

- PointerArea、Pressable、ClickArea、Container/Card、Tabs item 的视觉区域、布局区域、命中区域对齐。
- ripple/state layer 不改变 layout 和 hit size。
- cursor 设置、清理、继承和重置规则。
- decoration 合并不得丢失 disabled/hover/focus 表达。

### 验收

- 空白区域不触发 hover/click。
- Container decoration 不扩大交互区域。
- component_lab 中单个控件不会污染整窗 cursor。

### P6 实施记录（2026-07-08）

- 审计依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A3.4、A4.1、A4.2、A4.3、A4.4、A5.4、A10.8、A11.2、A13.4 的 layout/hit、decoration merge、interaction state、ripple/state layer、cursor 和 component_lab 风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 代码产物：`PointerArea` 使用放松 `Constraints.Min` 后的 child 尺寸注册 pointer hit rect，同时返回按父约束收束后的布局尺寸，避免父级 exact/min 约束把空白区变成 pointer 区；`LayoutClickArea` 的输入注册改为放松 Min 后测量，Pressable/ClickArea 默认只命中真实 child 区域；`LayoutRippleArea`/`LayoutRippleOverlayArea` 放松 Y 方向 inherited min，避免 ripple/click target 被整窗或整行高度放大，同时保留 Tabs equal-width 的 X 方向分配。
- 组件边界：ContainerDecoration/Card 继续以 surface 内容区注册交互，不把 margin、shadow 或 sibling 空白纳入 click；Tabs item 的 ripple/click target 不再逃逸到 item row 外部；ripple/state layer 仍只按 child size 绘制，不参与布局扩张；cursor 路径未新增全局状态，继续由 slider track 区域回归锁定。
- decoration/state：新增回归锁定 `resolveDecorationState` 和 `withDefaultStates`，确认 hover/pressed/disabled/focused 分支在合并和默认态补齐时不会被丢弃。
- 回归入口：`widget/pointer_area_test.go` 覆盖 exact 父约束下 PointerArea 空白区不触发 move；`widget/interactive_layout_test.go` 覆盖 Pressable、ContainerDecoration margin/shadow、ElevatedCard shadow/sibling、Tabs row 外空白点击；`widget/utils_test.go` 覆盖 decoration state 分支保留；既有 `TestSliderPointerCursorIsClippedToTrack` 作为 cursor 不污染整窗的自动化入口。
- 验证：`gofmt` 已执行；`gopls go_diagnostics` 对本轮修改文件无诊断；定向 `go test ./widget -run "Test(PointerAreaHitRectUsesRelaxedChildSize|PressableHitRectUsesChildSizeUnderExactParent|ContainerDecorationMarginAndSiblingBlankDoNotClick|CardRippleHitStaysInsideSurface|TabsPointerClickDoesNotEscapeItemRow|SliderPointerCursorIsClippedToTrack|ResolveDecorationStatePreservesInactiveStateBranches|WithDefaultStatesKeepsExplicitStateDecoration)$"` 通过；`go test ./widget` 通过；`go test ./event ./internal ./ui ./widget` 通过。`go_vulncheck ./...` 仍报告现有 Go 1.25.1 标准库与间接 `golang.org/x/image`、`golang.org/x/sys` 风险，本轮未修改依赖。
- 回滚策略：如放松 `LayoutClickArea` 的 Min 影响到少数组件的显式整宽命中，可优先在对应组件内部用 `FixedWidth`/`expandWidth` 包裹 child 或回退该组件调用点，不需要恢复全局旧行为；如 ripple Y 方向收窄影响特定控件的整高交互，可按 `LayoutRippleArea` 与 `LayoutRippleOverlayArea` 分开回退，并保留 P6 空白区反例测试作为准入基线。

## P7：示例、benchmark 和文档收敛

### 范围

- docs browser：文档页、搜索 chips、代码块、表格、示例弹窗。
- component_lab：样式、hover、cursor、overlay、idle redraw。
- event_system_testbench：P0-P7 操作说明和预期结果。
- examples：`advanced_components`、`drag_drop_showcase`、`form_validation`、`horizontal_scroll`、`virtual_scroll`。
- benchmark：高频 input 和大组件树。

### 验收

- 每个高风险修复至少有一个手动入口。
- 每个可自动化稳定路径至少有一个测试或 benchmark。
- README、docs browser 文档和示例入口不再漂移。

### P7 实施记录（2026-07-08）

- 审计依据：以 `docs/audits/project-audit-baseline.md` 为索引，复核 A11.1、A11.2、A11.3、A11.4、A12.1、A12.2、A12.3、A13.4 的 docs browser、component_lab、event_system_testbench、示例覆盖、benchmark 和最终候选风险记录，并对照 `COMPONENT_DEVELOPMENT_GUIDE.md`、`CODE_STYLE_GUIDE.md`、`FEATURE_INTEGRATION_CHECKLIST.md`。
- 文档产物：`README.md` 增加 P7 回归重点入口和 benchmark smoke 命令；`examples/docs_browser/README.md`、`examples/component_lab/README.md`、`examples/event_system_testbench/README.md` 补充 P7 或 P0-P7 手工验收表；`examples/advanced_components/README.md`、`examples/form_validation/README.md`、`examples/virtual_scroll/README.md` 更新为当前 `RunElement` 入口说明；新增 `examples/horizontal_scroll/README.md`；`examples/drag_drop_showcase/README.md` 增加平台能力区分的拖放 smoke；`docs/examples-inventory.md` 同步示例 runtime 状态，消除 legacy 入口漂移。
- 自动化产物：新增 `examples/docs_browser/p7_docs_test.go`，锁定五个 P7 重点示例的 `RunElement` 入口、README `go run` 命令、P7 smoke 编号和根 README / inventory 的回归入口；新增 `widget/scroll_bench_test.go`，提供 `BenchmarkWheelScrollViewVertical` 和 `BenchmarkHorizontalWheelDelta`，覆盖纵向 wheel offset 更新、横向 delta 消费和纵向 wheel 不误触发横向滚动的 benchmark 入口。
- 手动入口矩阵：docs browser 覆盖文档页、搜索/API 摘要、分类 chips、代码块、表格和示例弹窗；component_lab 覆盖主题/密度、hover/pressed、cursor、overlay、集合滚动和 idle redraw；event_system_testbench 覆盖 P0-P7 操作步骤、预期结果和诊断观察口径；专项 examples 覆盖 advanced components、drag/drop、form validation、horizontal scroll、virtual scroll。
- 验证：`gofmt` 已执行；`gopls go_diagnostics` 对新增 Go 文件无诊断；`go test ./examples/docs_browser`、`go test ./examples/component_lab`、`go test ./widget`、`go test ./examples/advanced_components ./examples/form_validation ./examples/horizontal_scroll ./examples/virtual_scroll ./examples/drag_drop_showcase` 通过；`go test ./widget -run '^$' -bench 'Benchmark(WheelScrollViewVertical|HorizontalWheelDelta)' -benchmem -benchtime=5x` 通过；`go test ./internal/perf -run '^$' -bench 'Benchmark(LayoutStaticTree|MouseMoveInteractiveTree|ListVirtualized|StaticSurfaceCache)' -benchmem -benchtime=3x` 通过。
- 关联性检查：本轮不修改 runtime、event、layout、widget 行为和公开 API 签名；新增 benchmark 只在测试路径构造 ScrollView fixture，不改变 `ScrollRef` / `ScrollOnChange` 语义；文档 smoke 明确 GUI、clipboard、drag/drop 外部后端和真实 IME 仍有平台依赖，避免把环境差异误判为框架回归。
- 仍保留风险：GUI 视觉结果、真实外部拖放、真实输入法组合和无剪贴板环境仍需要人工运行确认；短 benchtime benchmark 只证明入口可执行，不作为跨机器绝对性能预算；`go_vulncheck ./...` 在本会话开始时仍报告当前 Go 1.25.1 标准库与间接图像/系统包的既有公告，本轮未修改依赖。
- 回滚策略：文档变更可按示例 README、根 README、inventory 和路线图分组回退；`examples/docs_browser/p7_docs_test.go` 可独立移除而不影响产品行为；`widget/scroll_bench_test.go` 仅新增 benchmark，可独立回退且不影响运行时路径。

## PR 准入规则

| PR 类型 | 必须检查 | 建议测试 |
| --- | --- | --- |
| Runtime/Event | registry 生命周期、event path、default action、redraw reason、diagnostics 默认关闭成本 | `go test ./event ./internal` |
| Layout/Scroll | constraints、viewport、offset、axis、hit refresh、docs browser 横向区域 | `go test ./widget ./internal`，手动 `horizontal_scroll` |
| Widget | 旧 API、Ref、OnChange、focus、keyboard、style、examples | `go test ./widget ./event`，相关 example smoke |
| Overlay | portal/boundary、outside click、focus restore、Escape、z-order | `go test ./widget ./event`，手动 docs browser/component_lab |
| Text/Input | beforeinput、IME、submit、controlled value、programmatic source | `go test ./widget ./event`，手动 `form_validation` |
| Docs/Examples | 文档 meta、示例入口、README 与实现一致 | 示例构建或空跑 smoke |

## 完成定义

一个修复只有同时满足以下条件，才视为完成：

- 代码实现与对应审查结论不冲突。
- 关联组件检查表已执行并记录结果。
- 自动测试或手动验收覆盖了主要路径和至少一个反例。
- 文档、示例或迁移说明已更新。
- 回滚边界清楚，能按组件或 adapter 退回，不需要整体 revert。
