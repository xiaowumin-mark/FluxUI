# A13.4 最终修复候选排序

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A13.4 最终修复候选排序。

## 事实结论

A13.4 的目标是综合 A0.1 到 A13.3 的所有审查产物，按风险和影响面给出后续修复候选排序。排序口径如下：

1. 优先处理会造成错误命中、错误滚动、错误关闭、键盘越界或默认行为不可取消的用户可感知问题。
2. 优先处理跨 Runtime、Event、Layout、Widget 多层的边界问题，避免单点修复后继续在相邻组件中复发。
3. 对尚缺可观测性的高风险修复，先补 diagnostics 或回归入口，再推进结构性改动。
4. 对旧 API、Ref、OnChange、ScrollOnChange 等已承诺签名，修复必须保持兼容输出，不能以清理冗余为由改变外部语义。

### 修复候选排序

| 排序 | 候选 | 影响范围 | 依赖任务 | 验收示例 | 回滚策略 |
| --- | --- | --- | --- | --- | --- |
| F-01 | 统一滚动 wheel policy、default action gate 和 scroll 后 hit refresh 合同 | `ScrollView`、`ListView`、`GridView`、Tabs 横向滚动、chips/code/table 横向区域、Select/Menu 内部长列表、docs browser 主内容 | A3.2、A3.3、A6.1-A6.4、A10.4、A11.1、A11.4、A13.3 | `horizontal_scroll` 中横向 wheel 不污染纵向；`virtual_scroll` 滚动后 item 命中正确；Select menu 滚动后不移动鼠标点击当前坐标命中新 option；`wheel.PreventDefault()` 能阻止对应 default scroll | 保留现有 `ScrollView` pixel-offset 路径和 `ListView/GridView` Gio list 路径；通过 feature flag 或内部 adapter 逐组件接入；失败时仅回退 adapter 调用，不改公开 `ScrollRef` / `ScrollOnChange` 签名 |
| F-02 | 补齐 overlay outside click、focus restore、Escape、modal boundary 的统一策略 | Dialog、Popup、DropdownMenu、Select、Menu、Tooltip、Toast/Snackbar、KeyboardScope、portal event path | A8.1-A8.4、A10.6、A11.1、A13.2、A13.3 | Dialog modal 中 Tab 不越出面板；Escape 只关闭 topmost overlay；Select/DropdownMenu 内部点击和滚动不误关闭，外部点击按 protected rect 关闭；关闭后 focus 回到 trigger/field | 分层回滚：保留现有 mask click/protected rect；先把 focus/Escape 作为可关闭选项接入；若 topmost 仲裁异常，可回退到组件本地 KeyboardScope/旧 outside press 逻辑 |
| F-03 | 收敛 clickable/default action helper，统一 click、keyboard activation、hover snapshot 和 disabled gate | Button、Pressable、ClickArea、IconButton、FAB、Checkbox、Switch、RadioGroup、Tabs item、Select field/option | A1.4、A5.2-A5.5、A7.3、A10.1、A10.2、A13.1 | Button/Checkbox/Switch 的 pointer click 和 Enter/Space 激活都先派发 cancelable click；`PreventDefault` 阻止状态切换或旧 OnClick；disabled/loading 不触发 Ref click；Tabs 若接入新 helper，旧 tab 切换行为保持兼容 | 只抽窄 helper：`drainClickableDefaultAction`、`registerClickableFocusAction`、snapshot helper；不合并 Button surface、Tabs indicator、Select field visual；单组件接入失败时可回退到原手写 loop |
| F-04 | 稳定 controlled value、Ref 命令、OnChange 触发优先级 | Input/TextField/SearchBar、Select、RadioGroup、Checkbox、Switch、Slider、Tabs、ScrollView、Dialog/Popup/Select Ref | A9.1-A9.3、A10.2-A10.5、A1.4 | 程序设置和用户操作同帧不会互相覆盖；当前项重复选择不触发真实变化 OnChange；Ref 命令在 disabled/loading/未挂载时行为可预测；`form_validation` 快捷填充和用户编辑结果一致 | 保留旧 callback 签名和受控判定；优先增加内部 diff/gate 和测试；若 Ref 新 drain 规则异常，可按组件回退到旧命令消费路径 |
| F-05 | 补齐 event/redraw/registry diagnostics 归因字段和低成本 ring buffer | Runtime registry、event dispatcher、default action、redraw invalidator、state.Set、perf benchmarks、event_system_testbench | A2.3、A2.4、A5.2、A5.5、A12.1-A12.3 | 能输出每帧 target/listener/focus target 数量；能定位哪个 target/phase 调用了 `PreventDefault` 或 stop propagation；能区分 portal/boundary 改写原因；能定位 redraw source/path | diagnostics 默认关闭；新增字段放在可选开关或小容量 ring；如果性能回退，先关闭 registry dump 和 event history，只保留聚合计数 |
| F-06 | 修复文本输入 IME、beforeinput、submit 和程序化输入来源边界 | Input、TextField、SearchBar、InputRef、KeyboardScope、表单示例 | A5.1、A5.5、A7.4、A7.5、A10.3、A11.4 | IME composition 中 Enter 不提前 submit；真实用户输入、粘贴、删除、撤销/重做、Ref SetText 来源可区分；`beforeinput.PreventDefault` 不触发旧 `InputOnChange`；SearchBar 继承 TextField 语义 | 保留现有 best-effort rollback；先补 source 标记和测试，不改变 Gio editor 基础行为；若真实 IME 桥接不稳定，回退为 synthetic/event docs 明确限制 |
| F-07 | 收敛 layout area、visual area、hit area 和 decoration/state layer 边界 | PointerArea、Pressable、ClickArea、Container/Card、Tabs item、Checkbox/Switch/Radio、ripple/state layer、cursor | A3.1-A3.4、A4.1-A4.4、A10.8、A13.1 | 空白区域不触发 hover/click；Container decoration 不扩大交互区域；ripple 不改变 layout 或 hit size；单个控件不会把整窗 cursor 污染成 pointer | 不重写布局系统；优先补 hit area 测试和窄 helper；若视觉 helper 影响尺寸，回退到原组件绘制，仅保留诊断/测试 |
| F-08 | 建立示例 smoke、benchmark 和修复回归矩阵 | docs browser、component_lab、event_system_testbench、advanced_components、drag_drop_showcase、form_validation、horizontal_scroll、virtual_scroll | A0.3、A0.4、A11.1-A11.4、A12.1-A12.2 | 每个高风险修复都有至少一个手动入口和一个可自动化候选；高频 wheel/pointer benchmark 能区分 layout、event registration、input 分配热点 | 测试先以 docs/manual checklist 记录，不阻塞修复；自动化不稳定时回退为手动 smoke，但保留入口和预期结果 |

### 分组推进建议

| 阶段 | 建议内容 | 理由 |
| --- | --- | --- |
| 第一组 | F-05 diagnostics 最小字段 + F-01 滚动命中专项测试 | F-01 影响面最大，且现有 hit refresh 和 default action 归因不足；先补低成本观测和回归，降低误修风险。 |
| 第二组 | F-02 overlay focus/Escape/outside click + F-03 clickable default action helper | 这两组都直接影响事件路径、默认行为和焦点，适合共享 `PreventDefault`、KeyboardScope、portal boundary 的验收口径。 |
| 第三组 | F-04 controlled/Ref/OnChange + F-06 text input IME/source | 表单和输入语义高度相关，应一起冻结“用户输入、程序设置、Ref 命令、旧 callback”的优先级。 |
| 第四组 | F-07 hit/visual/style 边界 + F-08 示例/benchmark 扩展 | 作为前面修复的稳定性收尾，避免视觉 helper 或布局 helper 引入新命中范围问题。 |

## 风险

- F-01 是最高风险候选：`ScrollView` 和 `ListView/GridView` 当前分别由 FluxUI 和 Gio 管理滚动输入，强行一次性统一可能破坏虚拟化性能、旧 `ScrollOnChange` 输出或 wheel 可取消性。
- F-02 需要区分 modal、portal-only、deferred local popup、hover overlay 和 timed overlay；把所有 overlay 压成同一个 helper 会改变 Popup 冒泡、Tooltip hover 和 Toast 生命周期语义。
- F-03 若 helper 过大，会把 Button variant、Select field、Tabs indicator、Switch thumb/track 等真实差异压平；应只抽 click/focus/default action/snapshot 这类窄行为。
- F-04/F-06 涉及旧 callback 和受控语义，最容易形成同帧覆盖、重复 OnChange 或程序化输入错误触发用户回调。
- F-05 diagnostics 若默认开启细粒度 registry dump 或 event history，会在高频 pointer/wheel 场景引入分配热点；必须默认关闭并受 benchmark 约束。
- F-08 的手动 smoke 不能替代自动测试；但在 Gio GUI 行为难以完全自动化前，缺少手动入口会让修复回归不可复现。

## 验收

- 已按风险和影响面输出 F-01 到 F-08 的修复候选排序。
- 每个修复候选都包含影响范围、依赖任务、验收示例和回滚策略。
- 已明确首要候选为滚动 wheel/default action/hit refresh，其次为 overlay focus/Escape/outside click、clickable default action、controlled/Ref/OnChange、diagnostics、文本输入、hit/visual/style、示例/benchmark。
- 已给出分组推进建议，避免在缺少 diagnostics 和回归入口时直接做大范围行为重构。
- 本轮只记录最终排序，不修改 runtime、event、layout、widget 或 examples 行为。

## 后续依赖

- F-01 依赖 A6.1-A6.4 和 A13.3 的滚动模型边界，实施前应补 `wheel.PreventDefault`、scroll 后 hit refresh、Select menu 内滚动回归。
- F-02 依赖 A8.1-A8.4 和 A13.2 的 overlay 分类，实施前应先决定 Dialog/Popup/Select/DropdownMenu 的 focus restore 和 Escape 是否默认启用、是否可取消。
- F-03 依赖 A5.3/A5.5/A7.3/A13.1，实施时必须保持旧 `OnClick`、`OnHover`、选择类 `OnChange` 和 keyboard activation 的兼容顺序。
- F-04/F-06 依赖 A9.x、A7.4、A7.5 和 A10.3，实施时应同步更新 form/text/input 示例验收。
- F-05 依赖 A12.1-A12.3，新增诊断字段后必须跑高频事件和大组件树 benchmark，确认默认关闭路径无明显成本。
- F-08 应成为所有后续修复的复验索引，尤其覆盖 docs browser、component_lab、event_system_testbench、horizontal_scroll、virtual_scroll 和 drag_drop_showcase。
