# A5.4 PointerArea 影响范围审查

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：事件系统审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Event、Layout
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A5\\.4|PointerArea|PointerAreaElement|event system testbench|影响范围" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md`
  - `rg -n "PointerArea|PointerAreaElement|PointerOn|OnHover|Cursor|Hover" .`
  - `rg -n "PointerArea|HitRect|hit rect|整行|PointerAreaElement" widget examples docs/audits`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/pointer_area.go`
  - `widget/pointer_area_test.go`
  - `widget/spacer.go`
  - `widget/sizing.go`
  - `widget/flex.go`
  - `internal/render.go`
  - `examples/event_system_testbench/main.go`
  - `examples/event_system_testbench/FIX_ROADMAP.md`
  - `docs/audits/audit-a3.4-hit-layout-area.md`
- 关联能力：
  - PointerArea Gio tag 注册区域
  - PointerAreaElement 声明式包装边界
  - PointerArea 与 child layout dimensions 的关系
  - pass-through、disabled、captureOnPress 对命中范围的影响
  - event system testbench P2 手动验收入口

## 事实结论

### 实现链路

```text
PointerArea(child, opts...)
  -> 默认 passThrough = true
  -> Layout(ctx)
       if child == nil: return zero
       if disabled: return child.Layout(ctx.Child(0))
       state := pointerAreaStateFor(ctx)
       registerPointerAreaListeners(ctx, cfg)
       childDims := child.Layout(ctx.Child(0))
       registerPointerArea(ctx, state, childDims.Size, cfg)
       processPointerAreaEvents(ctx, state, cfg)
       return childDims

registerPointerArea(ctx, state, size, cfg)
  -> if size.X <= 0 || size.Y <= 0: no Gio event.Op
  -> if passThrough: pointer.PassOp{}.Push(...)
  -> clip.Rect(image.Rectangle{Max: size}).Push(...)
  -> gioEvent.Op(ctx.Gtx.Ops, state.tag)
```

结论：`PointerArea` 的 Gio 输入 tag 注册矩形直接使用 `childDims.Size`，并用 `clip.Rect(Max: size)` 限定。`PointerArea` 自身不绘制视觉，也不在注册时使用父窗口尺寸或整行尺寸。

### 注册区域和布局区域对照表

| 场景 | 布局返回区域 | Gio event.Op 注册区域 | 对兄弟控件影响 | 结论 |
| --- | --- | --- | --- | --- |
| `PointerArea(child)` 默认 | 返回 `child.Layout(ctx.Child(0)).Size` | `clip.Rect(Max: childDims.Size)` | 默认 `passThrough=true`，不阻断 sibling 接收 Gio pointer event；但命中仍只在 clip 内产生该 tag 的事件 | 默认边界来自 child 布局结果 |
| `PointerAreaDisabled(true)` | 返回 child layout size | 不注册 PointerArea listener/tag，不处理 pointer events | 不会因 PointerArea 产生额外命中 | 合格 |
| child 返回 `0x0` 或任一轴非正 | 返回 child size | `registerPointerArea` 直接返回 | 不会注册空/负尺寸目标 | 合格 |
| `PointerAreaPassThrough(false)` | 与 child 相同 | 与 child 相同 | 可改变 sibling 阻挡语义，但不扩大 clip rect | 不影响命中范围大小 |
| `PointerCaptureOnPress(true)` | 与 child 相同 | 与 child 相同 | press 后同 pointer 后续事件可派发到 capture owner；release/cancel 释放 | capture 改 dispatch target，不改注册区域 |
| 显式 `FillWidth` / `Fill` / `Expanded` 包裹 child | child 主动返回更大布局尺寸 | 注册区域随 child 返回尺寸扩大 | 这是调用方显式扩大 hit rect | 合法扩大 |
| 父级传入 exact/min 约束，child 会 `Constrain` 到 min | child 可能被约束放大 | 注册区域也随 childDims 放大 | 可能影响同一行空白或兄弟区域 | 当前主要风险 |

### PointerAreaElement 边界

`ui.PointerAreaElement(child, opts...)` 只是声明式 Element 对 `widget.PointerArea(child, opts...)` 的包装入口。它没有额外布局、clip 或事件命中策略，因此影响范围以底层 `widget.PointerArea` 的 child dimensions 为准。

### event system testbench 入口

`examples/event_system_testbench/main.go` 的 P2 `pointerWheelPanel` 使用：

```text
ui.PointerAreaElement(
  surface,
  ui.PointerCaptureOnPress(true),
  PointerOnOver/Enter/Leave/Out/Down/Move/Up/Cancel/Click/DoubleClick/AuxClick/ContextMenu/Wheel...
)
```

该入口是后续手动验收 PointerArea 影响范围的主要场景。历史 `FIX_ROADMAP.md` 已记录过 “P2 PointerArea 命中区域过大 / 影响整行” 的问题分类，并给出目标：默认 hit rect 应等于 child 实际布局尺寸；如果父布局分配更大空间，应通过显式 fill/passThrough 策略扩大。

### 测试覆盖现状

| 覆盖项 | 证据 | 结论 |
| --- | --- | --- |
| pointer、wheel、click-like 合成事件 | `TestPointerAreaDispatchesPointerWheelAndSyntheticEvents` | 覆盖 down/up/click/dblclick/aux/contextmenu/wheel 和 capture release |
| move 合并 | `TestPointerAreaCoalescesMovesToLatestSample` | 覆盖连续 move/drag 聚合到 latest sample |
| hit rect 使用 child size 的坐标级断言 | 未发现对应自动测试 | 缺口 |
| child 外同 row 空白不触发 PointerArea | 未发现对应自动测试 | 缺口 |
| sibling 控件不被 PointerArea 抢占 | 未发现对应自动测试 | 缺口 |
| testbench P2 手动验收 | `examples/event_system_testbench/main.go` | 可作为人工入口，但本次未启动 GUI 手动验证 |

## 风险

| 风险 | 严重度 | 说明 |
| --- | --- | --- |
| 父级 min/exact 约束间接放大命中区 | 高 | `PointerArea` 注册使用 child 返回尺寸，但没有在布局 child 前清空或放松 `Min` 约束。`Spacer` 等组件会用 `ctx.Gtx.Constraints.Constrain(size)`，因此在 exact/min 父约束下可能返回比视觉意图更大的尺寸。 |
| 缺少坐标级 hit rect 回归测试 | 高 | 当前测试证明事件字段和合成顺序，但不能证明 child 外、同 row 空白、兄弟控件区域不会触发 PointerArea。 |
| testbench 历史问题可能回归 | 中 | `FIX_ROADMAP.md` 已把 P2 归类为底层命中区域问题；若没有自动测试锁定，后续 layout 或 Element 包装调整可能重新扩大命中范围。 |
| pass-through 语义容易被误读 | 中 | `passThrough=true` 只影响是否阻断 sibling，不表示 PointerArea 命中整行；文档和测试需要区分“命中范围”和“事件透传”。 |
| capture target 容易掩盖真实 hit rect | 中 | `captureOnPress` 会让后续 pointerID dispatch 到 capture owner；验证 hit rect 时必须区分 press 初始命中区域和 capture 后续事件。 |

## 验收

- 已确认 `PointerArea` 的注册矩形来自 `childDims.Size`，且注册时使用 `clip.Rect(image.Rectangle{Max: size})`。
- 已确认 `PointerAreaElement` 没有额外扩大命中区域的包装逻辑。
- 已确认 disabled、zero-size、pass-through、captureOnPress 分别影响注册、阻挡或 dispatch target，但不直接扩大注册矩形。
- 已标出当前未满足的自动化验收缺口：缺少 child 外、同 row 空白、兄弟控件不受影响的坐标级测试。
- 已明确 event system testbench 的 P2 Pointer/Wheel 面板是后续手动验收入口。

本次审查未修改 runtime/widget/layout 代码；不把现有命中范围风险视为本次审查引入问题。

## 后续依赖

- 后续修复候选应优先补 `TestPointerAreaHitRectUsesRelaxedChildSize` 一类坐标级回归测试，断言普通 child 在父级 exact/min 约束下不会被 PointerArea 默认扩大为整行命中。
- A5 后续事件任务若审查 `KeyboardScope`、`EventBoundary` 或 overlay hit testing，应复用本任务的区分方式：注册区域、布局区域、dispatch target 分开记录。
- A3.4 hit/layout area 记录可作为对照；A5.4 的后续验证应补足 A3.4 已指出的坐标级测试缺口。
- testbench P2 手动验收建议固定操作：点击/滚动视觉面板内部、同一行面板外空白、相邻控件区域，分别确认日志只在视觉面板内部产生 PointerArea 事件。
