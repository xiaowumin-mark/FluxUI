# A6.1 ScrollView 和 ListView offset 审查
> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 3：滚动、手势和嵌套交互审查。

- 状态：Done
- 日期：2026-07-06
- 负责人：Codex
- 关注：Widget、Layout
- 输入命令：
  - `git status --short --branch --untracked-files=all`
  - `gopls go_workspace`
  - `gopls go_vulncheck ./...`
  - `rg -n "A6\\.1|ScrollView 和 ListView offset|ScrollOnChange|ScrollRef" docs/project-audit-roadmap.md docs/project-audit-task-breakdown.md docs/audits/project-audit-baseline.md`
  - `rg -n "type ScrollView|func ScrollView|ScrollOnChange|type ScrollRef|type ListView|func ListView|ListView\\(" widget ui internal event docs examples`
  - `rg -n "ScrollToOffset|ScrollBy|ScrollToEnd|ScrollRef|ScrollOnChange|lastY|lastX|off/1024|offset/1024|ScrollDistance|viewportFromListPosition|listViewState|Position\\.Offset" widget docs examples -g "*.go" -g "*.md"`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `widget/list_grid.go`
  - `widget/scroll_ref.go`
  - `widget/list_grid_test.go`
  - `widget/refs_test.go`
  - `docs/widgets/scroll_view.md`
  - `docs/audits/audit-a3.2-scroll-container-layout.md`
  - `docs/audits/audit-a5.3-legacy-event-bridge.md`
  - `docs/audits/audit-a5.5-default-action-cancelability.md`
- 关联能力：
  - ScrollView offset 来源、单位、更新时机
  - ScrollOnChange 旧回调触发和上报值边界
  - ScrollRef 命令消费与 offset clamp
  - ListView Gio Position 和 reach-end 关系
  - 旧 ScrollOnChange 长期显示 0 的原因定位

## offset 来源和单位矩阵

| 对象 | offset 来源 | 内部单位 | 更新时机 | 外部回调/观察 | 证据 |
| --- | --- | --- | --- | --- | --- |
| `ScrollView` wheel | Gio `pointer.Scroll` 经 `WheelEventFromGio` 转为 `WheelEvent`，未被 `PreventDefault` 取消后进入 `applyWheelDefault`。 | 当前主轴像素；纵向取 `DeltaY`，横向取 `DeltaX`；`wheelLeft` 保存小数残量。 | `Layout` 开始阶段先 `processWheelEvents`，再在 `layoutContent` 中 clamp、必要时重排 child。 | `ScrollOnChange` 在布局和 scrollbar 处理后比较 `First/Offset`，变化才调用。 | `widget/list_grid.go:321`、`widget/list_grid.go:338`、`widget/list_grid.go:164` |
| `ScrollView` `ScrollRef.ScrollToOffset` | `ScrollRef` 命令队列在 `ScrollView.Layout` 中 `drainCommands`。 | 参数是主轴绝对像素 offset；负值在入队前归零。 | 下一帧绑定 ref 的 ScrollView 消费命令；`layoutContent` 再按 content/viewport clamp。 | 若消费后 `Position.Offset` 变化，会触发旧 `ScrollOnChange`。 | `widget/scroll_ref.go:53`、`widget/list_grid.go:129`、`widget/list_grid.go:226` |
| `ScrollView` `ScrollRef.ScrollBy` | `ScrollBy(delta)` 入队为相对滚动命令。 | `delta` 不是像素；布局时换算为 `delta * contentMajor` 后四舍五入为像素。 | 命令消费时累计到 `pendingScrollBy`，测量出 content size 后换算并 clamp。 | 同上，回调观察最终 `Position.Offset`。 | `widget/scroll_ref.go:64`、`widget/list_grid.go:146`、`widget/list_grid.go:249` |
| `ScrollView` auto-to-end | `ScrollAutoToEnd` 或 `ScrollAutoToEndKey` 将 `Position.BeforeEnd=false`。 | 最终 offset 仍是像素，等于 `contentMajor - viewportMajor` 的 clamp 结果。 | 首次启用或 key 改变时在 `Layout` 前段设置；`layoutContent` 测量后滚到末尾。 | 末尾 offset 改变时触发 `ScrollOnChange`；不经过 wheel event。 | `widget/list_grid.go:150`、`widget/list_grid.go:253` |
| `ScrollView` scrollbar | Gio `widget.Scrollbar.ScrollDistance()` 给出归一化距离。 | 先是比例，再由 `scrollViewScrollByFraction` 换算到 `Position.Length` 像素。 | `layoutContent` 完成后 `drawScrollBar` 处理输入；变化后请求下一帧 redraw。 | `ScrollOnChange` 在 `drawScrollBar` 之后执行，会看到新的 `Position.Offset`。 | `widget/list_grid.go:436`、`widget/list_grid.go:447`、`widget/list_grid.go:386` |
| `ListView` virtualized | Gio `layout.List` 自己根据输入和 `Position` 维护滚动。 | Gio `layout.Position`：`First` 为首个可见项索引，`Offset` 为该项 leading edge 像素距离，`Length` 是估算内容总像素。 | `state.list.Layout` 内部更新；外部代码在布局后读取 `Position` 做 reach-end。 | 没有 `ScrollOnChange` 或 `ScrollRef`；公开观察点只有 `ListOnReachEnd`。 | `widget/list_grid.go:725`、`widget/list_grid.go:816` |
| `ListView` non-virtualized | 不使用 Gio `layout.List`，直接 Row/Column 排列全部 children。 | 无内部滚动 offset。 | 每次布局完整构建列表。 | 没有 offset 回调；大列表只记录非虚拟化诊断。 | `widget/list_grid.go:765` |

## ScrollOnChange 上报规则

当前 `ScrollView` 内部通过 `scrollViewSetPosition` 强制把 `Position.First` 设为 `0`，`Position.Offset` 保留主轴像素 offset，并写入 `Length`、`OffsetLast` 和 `BeforeEnd`。旧 `ScrollOnChange` 的回调值不是原始像素，而是：

```go
float32(first) + float32(off)/1024
```

由于 `ScrollView` 的 `first` 当前恒为 `0`，对外看到的 `x/y` 实际是 `offset / 1024`。例如一次 48 像素 wheel 默认滚动会变成 `0.046875`，如果示例用 `%.0f` 格式化，就会显示为 `0`。这解释了旧 `ScrollOnChange` 不应长期为 0 的核心边界：内部 offset 已经变化，测试也断言 `lastY > 0`，但当前兼容表达和显示格式会掩盖小滚动量。

## 事实结论

1. `ScrollView` 的持久状态保存在 `ctx.Memo("scroll")` 返回的 `scrollState` 中，包含 Gio `layout.List`、scrollbar、wheel tag、小数滚轮残量、上次回调位置和 auto-to-end key。
2. `ScrollView` 的真实滚动位置以 `state.list.Position.Offset` 保存，单位是当前主轴像素，最终由 `scrollViewClampOffset` 限制到 `[0, contentMajor - viewportMajor]`。
3. wheel 输入的默认滚动在 `processWheelEvents` 中先分发 cancelable `WheelEvent`；只有 dispatch allowed 时才调用 `applyWheelDefault` 修改 `Position.Offset`。
4. `applyWheelDefault` 会把 wheel delta 累加进 `wheelLeft`，只把整数像素部分写入 offset，小数部分跨事件保留。
5. `layoutContent` 会先用当前 offset 测量 child，再根据 `ScrollBy`、auto-to-end、content size 和 viewport size 计算最终 offset；offset 改变时会用最终 offset 再布局一次 child。
6. `scrollViewSetPosition` 把 `First` 固定为 `0`，所以 ScrollView 不使用 Gio List 的 item-index offset 语义，而是借用 `Position` 字段保存单个大内容的像素滚动。
7. `ScrollOnChange` 在 `layoutContent` 和 `drawScrollBar` 之后执行，只有 `First` 或 `Offset` 相对上次记录变化时才触发。
8. 纵向 ScrollView 调用 `ScrollOnChange(ctx, 0, value)`；横向 ScrollView 调用 `ScrollOnChange(ctx, value, 0)`。
9. 当前 `ScrollOnChange` 的 `value` 是 `First + Offset/1024`，不是像素 offset；对 ScrollView 来说等价于 `Offset/1024`。
10. `ScrollRef.ScrollToOffset` 的公开注释说明参数是像素；命令入队后在下一帧由绑定的 ScrollView 消费，并继续走 clamp 和回调路径。
11. `ScrollRef.ScrollBy` 的公开注释说明单位同 Gio List.ScrollBy；实际实现把 delta 乘以 content 主轴长度后作为相对像素位移。
12. `ScrollRef.ScrollToStart/Top` 和 `ScrollToEnd/Bottom` 都是主轴命令；绑定横向 ScrollView 时也会表现为左右滚动，不是只限垂直方向。
13. `ScrollAutoToEnd` 和 `ScrollAutoToEndKey` 通过 `BeforeEnd=false` 驱动布局阶段滚到底部，不经过 wheel event，也不受 `wheel.PreventDefault` 控制。
14. scrollbar 输入在 `drawScrollBar` 中处理，晚于内容布局；它会更新 `Position` 并请求 redraw，因此回调可在当帧观察到新 offset，而可视内容通常依赖下一帧按新 offset 重绘。
15. virtualized `ListView` 使用独立的 `listViewState` 和 Gio `layout.List.Position`，offset 语义是 Gio 的 item-index 加像素偏移，不是 ScrollView 的单内容像素 offset。
16. `ListView` 当前没有 `ScrollOnChange`、`ScrollAttachRef` 或公开 offset getter；滚动状态只被组件内部用于虚拟化窗口和 `ListOnReachEnd`。
17. non-virtualized `ListView` 不使用 Gio `layout.List`，因此没有可审查的滚动 offset，只在大列表场景记录性能诊断。
18. `TestScrollViewWheelScrollsContent` 已覆盖 wheel 后 `ScrollOnChange` 会触发且 `lastY > 0`，说明旧回调不是永久 0。
19. `TestHorizontalScrollViewRequiresHorizontalWheelDelta` 已覆盖横向 ScrollView 只响应横向 wheel delta，并把变化上报到 `x`。
20. `TestListViewWheelScrollsContent` 和嵌套滚动测试覆盖 ListView 能吃到 wheel 并触发 reach-end，但没有暴露或断言具体 offset 数值。

## 风险

| 风险 | 等级 | 说明 | 后续关联 |
| --- | --- | --- | --- |
| `ScrollOnChange` 单位和注释不一致 | 高 | 文档说“滚动偏移回调”，`ScrollRef.ScrollToOffset` 明确像素，但旧回调实际返回 `Offset/1024`。调用方容易把它误认为像素，示例用 `%.0f` 时也会持续显示 `0`。 | A11 docs/example、后续 scroll API 修复 |
| `ScrollBy` 单位不直观 | 中 | `ScrollBy(2.0)` 会换算为 `2 * contentMajor` 像素，而不是 2 像素；公开注释只写“单位同 Gio List.ScrollBy”，对 FluxUI 用户不够直接。 | A9 Ref 命令审查、A11 文档 |
| scrollbar 回调与视觉更新不同步 | 中 | scrollbar 在内容布局后更新 offset，`ScrollOnChange` 当帧可见新值，但内容已经按旧 offset 绘制，需要下一帧重绘。若外部逻辑假设回调值和当帧画面完全同步，可能出现一帧差。 | A6.2 嵌套/手势、A12 perf |
| ListView 无公开 offset 观察点 | 中 | virtualized ListView 内部有 Gio Position，但只有 `ListOnReachEnd`，不能像 ScrollView 一样观察连续 offset。若后续需要统一 scroll event，需要先定义 ListView 的公开单位。 | A6 后续滚动事件、A11 文档 |
| ScrollRef 应用缺少直接测试 | 中 | 当前只测试命令队列顺序，未断言 `ScrollToOffset/ScrollBy/ScrollToEnd` 消费后是否触发 `ScrollOnChange`、是否按像素/比例 clamp。 | A9 Ref 测试、A13 收敛 |
| auto-to-end 与回调缺少测试 | 低 | auto-to-end 会改变 offset 并可能触发回调，但现有自动测试没有直接覆盖 key 变化后的回调值。 | A6/A11 |

## 验收

- 已列出 ScrollView 的 offset 来源：wheel、ScrollRef、auto-to-end、scrollbar。
- 已明确 ScrollView 内部真实 offset 单位是主轴像素，最终由 content/viewport clamp。
- 已明确旧 `ScrollOnChange` 当前上报值是 `First + Offset/1024`，对 ScrollView 实际是 `Offset/1024`，不是像素。
- 已解释旧 `ScrollOnChange` 不应长期为 0：内部像素 offset 会变化，`widget.TestScrollViewWheelScrollsContent` 已断言回调值大于 0；长期显示 0 主要来自 `offset/1024` 表达和整数格式化。
- 已区分 ListView 的 Gio `layout.Position` 与 ScrollView 的单内容像素 offset：ListView 不暴露 `ScrollOnChange` 或 `ScrollRef`。
- 已标出当前测试覆盖和缺口：wheel/横向/嵌套/ListView reach-end 有覆盖，ScrollRef 应用、auto-to-end 回调、scrollbar 回调缺直接测试。

## 后续依赖

- A6.2 嵌套滚动和手势审查应复用本文的轴向过滤、wheel default action、ListView 与 ScrollView 状态差异。
- A9 Ref 审查应补充 `ScrollRef` 命令消费后的 offset、clamp 和 `ScrollOnChange` 触发测试建议。
- A11 示例和文档审查应决定是否把 `ScrollOnChange` 文档写成“兼容位置表达”或推动后续修复为真实像素 offset。
- 若后续修复 `ScrollOnChange` 为像素值，需要把 A1.4/A3.2/A5.3 中记录的旧 API 兼容边界一起更新，并补充迁移说明。
