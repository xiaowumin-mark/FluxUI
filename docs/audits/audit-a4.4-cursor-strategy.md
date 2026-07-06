# A4.4 cursor 策略审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 2：布局、渲染和样式稳定性。

- 状态：Done
- 日期：2026-07-06 19:50:00 +08:00
- 负责人：Codex
- 关注：Style、Event
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A4\\.4|cursor|Cursor" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "Cursor|cursor|pointer\\.Cursor|CursorPointer|CursorDefault|CursorText|CursorNotAllowed" internal event widget ui examples/component_lab -g "*.go"`
  - `rg -n -C 3 "SliderElement|RangeSliderElement|Slider|RangeSlider|SliderOnChange|SliderWidth|SliderLabeled|SliderTicks" examples/component_lab/main.go widget/slider.go ui/ui.go ui/extended_types.go`
  - `rg -n -C 4 "registerSliderPointer|processSliderPointerEvents|trackRect|rect :=|Slider\\(|func \\(s \\*sliderWidget\\) Layout|disabled" widget/slider.go`
  - `rg -n -C 3 "Cursor\\(\\)|CursorPointer|CursorDefault|TestSliderPointerCursorIsClippedToTrack|BlankArea|HoverTarget" widget internal event ui -g "*_test.go" -g "*.go"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `examples/component_lab/main.go`
  - `widget/slider.go`
  - `widget/slider_interaction_test.go`
  - `widget/interactive_layout_test.go`
  - `ui/ui.go`
  - `ui/extended_types.go`
- 关联能力：
  - component_lab cursor 异常入口定位
  - Gio cursor op 设置边界
  - Slider track cursor clip 策略
  - disabled cursor 禁用边界
  - cursor 清理和默认重置验证
  - 单控件整窗 cursor 污染风险识别

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main`，相对 `origin/main` ahead 20 |
| `git status --short --branch --untracked-files=all` | 仅输出分支行，无脏文件清单 |
| 判断 | A4.4 执行前工作区干净；本任务只新增 `audit-a4.4-cursor-strategy.md` 并更新索引，不修改 runtime/widget/style 源码。 |

## component_lab 入口

| 入口 | 位置 | cursor 相关性 | 结论 |
| --- | --- | --- | --- |
| `SliderElement` continuous | `examples/component_lab/main.go` controls 面板，`SliderWidth(280)`，绑定 `SliderOnChange` 和 `SliderAttachRef`。 | 进入 `widget.Slider`，最终调用 `registerSliderPointer`。 | component_lab 中最直接的 cursor 异常观察入口。 |
| `SliderElement` labeled + ticks | `examples/component_lab/main.go` controls 面板，开启 `SliderTicks(true)` 和 `SliderLabeled(true)`。 | 同一 slider cursor 路径；label 只影响上方显示空间，不改变 `trackRect` 的 cursor 注册宽度。 | 可用于观察 label 区域是否错误变 pointer。 |
| `RangeSliderElement` | `examples/component_lab/main.go` controls 面板，开启范围滑块和 ticks。 | 复用同一 `registerSliderPointer`，handle hover 区分 start/end，但 cursor 仍注册在 trackRect。 | 可用于观察范围滑块是否把整行或整窗污染成 pointer。 |
| Button、Chip、Tabs、ContainerDecoration 等普通交互组件 | component_lab 多处使用 clickable/onClick/onHover。 | 源码搜索未发现这些路径显式调用 `pointer.Cursor*.Add`。 | 当前不是 cursor 设置源；它们主要影响 hover/pressed 状态，不写 cursor op。 |

## cursor 设置规则

| 规则项 | 当前实现 | 结论 |
| --- | --- | --- |
| 设置来源 | `widget/slider.go` 的 `registerSliderPointer` 在非 disabled 且 state/tag 有效时执行。 | FluxUI 自身显式设置 pointer cursor 的来源目前集中在 Slider。 |
| 设置值 | `pointer.CursorPointer.Add(ctx.Gtx.Ops)`。 | slider track 鼠标悬停时使用 pointer cursor。 |
| 命中范围 | `clip.Rect(rect).Push(ctx.Gtx.Ops)` 包住 `gioEvent.Op` 和 cursor op；调用处传入 `trackRect := image.Rect(0, trackTop, width, trackTop+stateLayer)`。 | cursor op 被裁剪到 slider track/state-layer 高度范围，不覆盖整个组件外层或窗口。 |
| 事件注册 | 同一 clip 内同时注册 `gioEvent.Op(ctx.Gtx.Ops, state.tag)`，随后 `processSliderPointerEvents` 用同一 tag 和 rect 处理 pointer events。 | cursor 命中范围与 slider pointer target 使用同一个 rect。 |
| pass-through | `defer pointer.PassOp{}.Push(ctx.Gtx.Ops).Pop()` 包住 slider input op。 | slider input/cursor 不应阻断同级其它目标；这与 hover target 污染是不同问题。 |
| disabled | `registerSliderPointer` 遇到 disabled 直接 return；`processSliderPointerEvents` 也会清空 hover/pressed/active。 | disabled slider 不设置 pointer cursor，也不会保留旧 hover 状态。 |

## 清理、继承和重置规则

| 维度 | 规则 | 证据/说明 |
| --- | --- | --- |
| 清理 | FluxUI 没有自己的全局 cursor registry；cursor 生命周期跟随每帧 Gio ops。组件本帧不写 cursor op 时不会由 FluxUI 保留上帧 pointer。 | 搜索未发现 runtime/widget 中有 cursor 全局状态或 SetCursor 封装。 |
| 重置 | Gio `input.Router.Cursor()` 在没有命中 cursor op 的区域返回 `pointer.CursorDefault`。 | `TestSliderPointerCursorIsClippedToTrack` 先在 track 内断言 pointer，再移动到 track 外断言 default。 |
| 继承 | 没有 FluxUI 层面的 cursor 继承 API；父级 Decoration、hover、ripple、state layer 不会自动继承或派生 cursor。 | 搜索显示 Decoration/state layer/ripple 路径不调用 `pointer.Cursor*.Add`。 |
| 视觉与 cursor 分离 | slider label、state layer、handle 绘制不决定 cursor；cursor 由 `trackRect` input op 决定。 | A4.3 已记录 ripple/state layer 不改变 layout/hit size；A4.4 延续该边界。 |
| 事件与 cursor 分离 | hover/pressed 状态来自 pointer/clickable target；cursor 是 Gio ops 的命中结果，不由 `OnHover` 回调写全局状态。 | Button/Container/Tabs 等 hover 状态路径未写 cursor op。 |

## 测试覆盖

| 覆盖项 | 测试 | 结果 |
| --- | --- | --- |
| slider track 内 cursor | `widget.TestSliderPointerCursorIsClippedToTrack` | 轨道内 pointer move 后 `router.Cursor()` 为 `pointer.CursorPointer`。 |
| slider track 外重置 | `widget.TestSliderPointerCursorIsClippedToTrack` | 同一 ops 下移动到 y=60，`router.Cursor()` 回到 `pointer.CursorDefault`。 |
| 空白区域 hover target 不污染 | `widget.TestBlankAreaMovesDoNotTriggerHoverTarget` | 空白区域 pointer move 不触发 hover 回调，也没有 hover target。 |
| pointer move 合并后仍不产生空白目标 | `widget.TestPointerMoveCoalescingUsesLatestPosition` | 最新位置为空白区域时 hover target 为空。 |

## 事实结论

- `docs/project-audit-roadmap.md` 明确风险是“某个控件为了修交互，把全局鼠标 cursor 或 hover 状态污染到整窗”，并要求 component_lab 不出现整窗 cursor 异常。
- 当前 FluxUI 源码中显式 `pointer.CursorPointer.Add` 的业务组件路径只有 `widget.Slider`；`internal/hook_store.go` 等 `cursor` 命名是 hook cursor，不是鼠标 cursor。
- `registerSliderPointer` 同时把 Gio event tag 和 cursor op 放进 `clip.Rect(trackRect)` 内，因此 slider 的 pointer cursor 与 pointer event target 共享同一矩形边界。
- `trackRect` 高度使用 `stateLayer`，宽度使用 slider width；当 `SliderLabeled(true)` 产生上方 label 空间时，label 空间不在 trackRect 内，理论上不应让 label 区域变 pointer。
- disabled slider 不注册 cursor op，并且事件处理入口会清空 `hoverStart`、`hoverEnd`、`pressed`、`active`，没有 disabled 状态下保留 pointer cursor 的源码路径。
- Button、Pressable、ClickArea、ContainerDecoration、Tabs、Chip 等 component_lab 交互控件没有显式 cursor 设置；它们可能改变 hover/pressed 视觉，但不会直接把窗口 cursor 写成 pointer。
- cursor 清理由 Gio ops/router 语义承担：本帧未命中的 cursor op 不会被 FluxUI runtime 保存；既有测试已经覆盖 slider track 外回到 default。

## 风险

| 风险 | 等级 | 说明 |
| --- | --- | --- |
| cursor 策略未抽象成 FluxUI 统一 API | 中 | 目前只有 Slider 手写 `pointer.CursorPointer.Add`。后续其它控件若各自手写 cursor op，可能出现 clip 漏包、disabled 未跳过、或与 hit area 不一致。 |
| Slider cursor 区域是 trackRect，不是 20px handle hover 半径 | 中 | `updateSliderHover` 用 20px 半径判断 handle hover，cursor 则覆盖整个 track/state-layer 矩形。二者都在 slider 交互区域内，但不是完全同一几何规则；后续若要求 cursor 只在 handle 附近显示，需要单独调整策略。 |
| `pointer.PassOp` 与重叠组件优先级需继续观察 | 低 | Slider cursor/input op 使用 pass-through，便于同级继续接收事件；重叠场景下 cursor 最终优先级取决于 Gio router 的命中和 op 顺序，当前未建立 FluxUI 层测试矩阵。 |
| Text input 未设置 text cursor | 低 | 当前审查只关注整窗 pointer 污染；搜索未发现 `CursorText`。这不是 A4.4 验收失败，但后续若定义完整 cursor UX，需要把输入框纳入策略。 |
| component_lab 目前是手动观察入口，不是自动 cursor smoke | 低 | 已有 slider 单元测试覆盖最关键污染风险，但 component_lab 本身没有自动鼠标移动 smoke 断言。 |

## 验收

| 验收项 | 结果 | 证据 |
| --- | --- | --- |
| 输出 cursor 设置规则 | 通过 | 已记录 Slider 的 `registerSliderPointer`、`CursorPointer.Add`、`trackRect` clip 和 disabled return。 |
| 输出 cursor 清理规则 | 通过 | 已记录无 FluxUI 全局 cursor registry，清理由每帧 Gio ops/router 命中承担。 |
| 输出 cursor 继承规则 | 通过 | 已记录 Decoration/ripple/state layer/hover 不派生 cursor，当前无 FluxUI cursor 继承 API。 |
| 输出 cursor 重置规则 | 通过 | `TestSliderPointerCursorIsClippedToTrack` 覆盖 track 外回到 `CursorDefault`。 |
| 没有单个控件把整窗 cursor 污染成 pointer | 通过 | Slider cursor 被 `clip.Rect(trackRect)` 限定，且测试覆盖 track 外 default；其它 component_lab 交互控件未写 cursor op。 |
| 未修改运行时代码 | 通过 | 本任务只新增审计子文件并更新索引。 |

## 后续依赖

- A4.5 动画和布局隔离审查应继续确认动画 label、popup、overlay 不改变 cursor/input op 的几何边界。
- A5 事件系统审查需要把 cursor op 与 pointer target、hover target、capture target 区分记录：cursor 命中不应被误当成事件分发目标。
- A7 Widget 矩阵审查建议为 Slider、Input、DragSource/DropTarget、resizable/scrollable 类控件补一列 cursor 策略，避免后续新增 cursor op 时脱离统一规则。
- 后续若引入 `CursorText`、`CursorGrab`、`CursorNotAllowed` 等语义，应优先建立通用 helper，要求调用方必须同时声明 hit rect、disabled 行为和测试用例。
