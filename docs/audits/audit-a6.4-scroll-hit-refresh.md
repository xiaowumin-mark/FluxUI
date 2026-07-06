# A6.4 滚动后命中刷新审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：滚动、手势和嵌套交互审查。
- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Runtime、Event、Widget
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A6\.4|滚动后命中|hit-test|hover target|ScrollView|ListView|下拉菜单" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "QueuePointerMove|CoalescedPointerMove|PointerMove|pointer.*target|hoverTarget|Hovered|RegisterPointer|PointerArea|hit|Hit|Register.*Target|DispatchPointer|Scroll|Wheel|Delta" internal event widget -g "*.go"`
  - `rg -n "func .*ScrollView|type .*ScrollView|ListView|Select|Menu|ScrollRef|ScrollOnChange|pointer|hover|Hovered|Hit" widget ui internal event -g "*.go"`
  - `rg -n "WithPositionOffset|WithPointerPassThrough|Position\(|LayoutClickArea|LayoutRippleArea|LayoutRippleOverlayArea" internal -g "*.go"`
  - `rg -n "ClickAfterScroll|WheelScrollsContent|LargeMenuAndSelectUseVirtualized|NestedScrollPrefers|HorizontalScrollView" widget/list_grid_test.go`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `internal/interaction.go`
  - `internal/context.go`
  - `internal/clickable.go`
  - `internal/render.go`
  - `internal/ripple.go`
  - `event/pointer.go`
  - `widget/list_grid.go`
  - `widget/list_grid_test.go`
  - `widget/selection.go`
  - `widget/material3_components.go`
- 关联能力：
  - scroll 后 visual position 与 hit position 重建
  - ScrollView/ListView 当前帧 ops/filter 重建
  - Select/Menu 下拉列表滚动命中入口
  - pointer hover target 统计刷新
  - 不移动鼠标点击滚动后新目标的验收边界

## 事实结论

1. A6.4 的目标是确认滚动改变内容位置后，hit-test cache 或等效命中状态会更新；验收要求是滚动后不移动鼠标也能点击当前鼠标下的新目标。证据：`docs/project-audit-task-breakdown.md:328`、`docs/project-audit-roadmap.md:209`、`docs/project-audit-roadmap.md:215`。
2. 代码中没有 FluxUI 自己维护的通用 hit-test cache。命中区域主要由 Gio 每帧根据 ops、clip、transform 和 input filter 重建；FluxUI runtime 维护的 `hoverTarget` 是诊断/交互统计状态，不是事件分发命中缓存。证据：`internal/interaction.go:42`、`internal/interaction.go:64`、`internal/interaction.go:114`。
3. `ScrollView` wheel 默认行为在收到可取消且未取消的 wheel 后更新 `state.list.Position.Offset`，并调用 `ctx.RequestFrameRedrawReason("input.scrollview.wheel")` 请求下一帧。证据：`widget/list_grid.go:295`、`widget/list_grid.go:321`、`widget/list_grid.go:347`。
4. `ScrollView` 布局子内容时会按当前 offset 调用 `WithPositionOffset(scrollOffset)`，并在绘制时用同一个 `op.Offset(scrollOffset)` 回放子内容 ops；因此下一帧的视觉位置和 Gio input ops 位置同源。证据：`widget/list_grid.go:187`、`widget/list_grid.go:221`、`widget/list_grid.go:269`。
5. `ScrollView` 的 wheel target 是 viewport 大小的 Gio tag，并用 `pointer.PassOp` 注册，避免阻断子控件点击；它只负责 wheel 入口，不替代子控件自己的 click/hit 区域。证据：`widget/list_grid.go:267`、`widget/list_grid.go:284`、`widget/list_grid_test.go:116`。
6. `ListView` 直接使用 Gio `layout.List.Layout` 构建可见项；滚动后可见项和它们内部的 `Clickable` 区域由 Gio list 在下一帧按新 `Position` 重建。证据：`widget/list_grid.go:690`、`widget/list_grid.go:706`。
7. `Select` 下拉面板的选项列表使用 `ListView(..., ListVirtualized(true))`，选项行使用 `event.UseClickable` 和 `LayoutRippleArea` 注册点击区域；因此 Select 菜单滚动命中继承 ListView 与 Clickable 行为。证据：`widget/selection.go:527`、`widget/selection.go:654`、`widget/selection.go:560`、`widget/selection.go:634`。
8. `Menu` 在选项高度超过 `MenuMaxHeight` 时使用 `FixedHeight(maxHeight, ListView(..., ListVirtualized(true)))`；菜单行使用 `event.UseClickable`、`Snapshot` 和 `md3ActionSurface` 注册命中与 hover/pressed 状态。证据：`widget/material3_components.go:492`、`widget/material3_components.go:536`、`widget/material3_components.go:592`。
9. `Clickable.Hovered()` 和 `Clickable.Snapshot()` 会读取 Gio `widget.Clickable` 当前 hover/pressed/focus 状态，并通过 `ObserveInteractionSnapshot` 把当前帧 hover target 上报给 runtime。证据：`event/pointer.go:91`、`event/pointer.go:130`、`event/pointer.go:177`。
10. runtime 每帧开始时清空 `activeHover`，每帧结束时用本帧实际上报的 hovered clickable 生成 `hoverTarget`；上一帧 hover target 不会无条件残留为当前帧 active hover。证据：`internal/interaction.go:46`、`internal/interaction.go:64`。
11. 已有自动测试覆盖“横向 ScrollView 滚动后，不移动鼠标，在同一屏幕坐标点击会命中新视觉位置下的子项”：`TestHorizontalScrollViewClickAfterScrollUsesUpdatedVisualPosition` 先滚动 90px，再在 `(20,20)` press/release，期望点击 item 1。证据：`widget/list_grid_test.go:414`。
12. 已有测试覆盖 ScrollView wheel 后 offset/onChange 变化、ListView wheel 后可见/末尾状态变化、嵌套滚动目标优先级，以及 wheel target 不阻断子点击。证据：`widget/list_grid_test.go:169`、`widget/list_grid_test.go:480`、`widget/list_grid_test.go:534`、`widget/list_grid_test.go:601`、`widget/list_grid_test.go:116`。
13. 未发现专门测试覆盖纵向 ScrollView 滚动后同屏幕坐标点击新目标，也未发现 Select/Menu 下拉列表滚动后不移动鼠标点击当前新选项的回归测试。
14. 未发现专门测试覆盖滚动后不移动鼠标时 `hoverTarget` 或 hover 视觉态从旧子项切换到新子项；当前结论只能说明 runtime 不保留上一帧 `activeHover`，不能证明 Gio 在无 pointer move 场景一定产生新的 hover enter/leave。
15. 代码没有在 scroll default action 后主动合成 `pointermove` 或调用 FluxUI 层 hit-test refresh；滚动后的 hover 刷新依赖 Gio router 在 `router.Frame(&ops)` 后根据新 ops/filter 更新 hover 状态。

### scroll 后命中刷新规则表

| 场景 | offset 更新来源 | 命中区域刷新来源 | 已有自动覆盖 | 结论 |
| --- | --- | --- | --- | --- |
| `ScrollView` wheel | `applyWheelDefault` 更新 `Position.Offset` 并请求 redraw | 下一帧 `WithPositionOffset` + `op.Offset` 重建子 ops/filter | 横向点击新目标已覆盖 | 点击命中有正向证据 |
| `ScrollView` scrollbar | `Scrollbar.ScrollDistance` 更新 `Position` 并请求 redraw | 下一帧同 ScrollView offset 链路 | 未见点击新目标专项测试 | 需要补测试 |
| `ListView` wheel | Gio `layout.List` 更新 `Position` | Gio list 下一帧只构建可见项并注册子项 ops/filter | 只覆盖滚动/reach-end | 点击新目标需补测试 |
| `Select` 下拉菜单 | 内部 `ListView` | 选项行 `Clickable` 随 ListView 可见项重建 | 只覆盖大列表虚拟化 | 高风险入口，需补测试/手动验收 |
| `Menu` 下拉菜单 | 内部 `ListView` | 菜单行 `Clickable`/`Snapshot` 随 ListView 可见项重建 | 只覆盖大列表虚拟化 | 高风险入口，需补测试/手动验收 |
| pointer hover target | Gio `Clickable.Hovered()` 当前帧状态 | runtime 每帧 activeHover 重新收集 | 未见 scroll 后 hover 专项测试 | 依赖 Gio，证据不足 |

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| hover 刷新缺少直接回归测试 | 高 | 当前 runtime 不保留上一帧 active hover，但没有证明 scroll 后无 pointer move 时 Gio 一定让新目标 `Hovered()`。如果 Gio 后端只在 pointer move 时更新 hover，视觉 hover 和子菜单 hover-open 可能滞后。 | A8 overlay、A12 测试、A13 hit-test 冗余 |
| Select/Menu 滚动后命中未专项覆盖 | 高 | 下拉菜单是本任务明确输入，且包含 overlay、ListView、Clickable、hover submenu 多层组合；现有测试只证明大列表虚拟化，没有证明滚动后同坐标点击新选项。 | A8 overlay、A12 测试 |
| 纵向 ScrollView/ListView 点击新目标缺少对称测试 | 中 | 横向 ScrollView 已有正向测试，但纵向 ScrollView 和 Gio ListView 路径不同，不能完全用横向用例代替。 | A12 测试 |
| scrollbar 拖动后的命中刷新未覆盖 | 中 | scrollbar drag 更新 offset 的入口与 wheel 不同，虽然最终同样依赖下一帧布局，但没有 click-after-drag 测试。 | A6.5/A13 |
| FluxUI 无主动 hit refresh API | 中 | 当前没有“滚动后合成 pointermove/刷新 hover target”的明确机制；若未来需要独立于 Gio 后端保证 hover refresh，需要新增 runtime/widget 级设计。 | A13 收敛修复 |

## 验收

- 已记录 scroll 后命中刷新规则：`ScrollView` 和 `ListView` 都依赖下一帧 Gio ops/filter 重建，FluxUI runtime 没有独立 hit-test cache。
- 已确认 `ScrollView` wheel 会请求 redraw，滚动 offset 会在下一帧参与子内容视觉与输入区域位移。
- 已确认横向 `ScrollView` 已有“不移动鼠标，点击滚动后新视觉目标”的自动测试。
- 已确认 Select/Menu 下拉菜单都通过 `ListView` 和 `Clickable` 参与滚动命中，但缺少专项测试。
- 已确认 hover target 每帧重新收集，不会把上一帧目标无条件作为当前 active hover；但 scroll 后无 pointer move 的 hover 切换缺少直接证据，应作为后续测试/修复输入。

## 后续依赖

- A8 overlay 审查应重点验证 Select/Menu popup 内部滚动后，当前鼠标下的 hover、submenu open 和 click target 是否同步到新选项。
- A12 测试审查应补齐纵向 ScrollView、ListView、Select、Menu 的 scroll 后同坐标 click-after-scroll 测试。
- A12 或 A13 应补齐 scroll 后 hover target 刷新测试，至少覆盖 `Clickable.Hovered()`、runtime `LastInteractionStats().HoverTarget` 和菜单 hover submenu。
- A13 滚动和 hit-test 冗余审查应决定是否需要 FluxUI 层主动 hover refresh 或 synthetic pointer move，而不是继续隐式依赖 Gio 后端行为。
