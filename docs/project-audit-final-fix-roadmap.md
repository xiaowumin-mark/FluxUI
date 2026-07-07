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
