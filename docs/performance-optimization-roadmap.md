<!-- fluxui-doc-meta
{
  "id": "performance_optimization_roadmap",
  "title": "性能优化路线图",
  "category": "工程路线图",
  "order": 40,
  "summary": "规划 FluxUI 在大型项目、密集组件和高频鼠标交互场景下的性能治理路线，明确优先优化 Flux 层，只有在 profiling 证明 Gio 底层成为瓶颈时才评估定制 Gio。",
  "example": { "id": "list_view_basic" }
}
-->

# FluxUI 性能优化路线图

本文档定义 FluxUI 后续性能治理的长期路线，重点解决大型项目、组件数量多、鼠标 hover/move 高频交互时 CPU 占用过高的问题。

核心判断：

- 短中期优先优化 FluxUI 自身的运行时、组件、事件、动画和布局策略。
- 不优先 fork Gio。只有 profiling 证明瓶颈位于 Gio 底层且 Flux 层无法绕开时，才进入定制 Gio 评估。
- 性能优化必须先建立可重复测量体系，再做分阶段改造，避免凭感觉重构。

## 背景

FluxUI 当前建立在 Gio immediate-mode 模型上。每个 frame 会重新执行 UI 构建、布局和绘制命令生成。这个模型简单直接，适合中小规模 UI，也便于实现自定义组件，但在大型应用中有明显风险：

- 鼠标移动会触发 hover 命中和交互状态更新。
- 大量可交互组件会在同一帧查询 hover、pressed、focus、animation 状态。
- 组件树较大时，每次操作都可能放大全树 layout/draw 成本。
- 浏览器有成熟的局部失效、合成层、样式缓存、命中测试缓存和虚拟化生态，因此同等 UI 规模下 CPU 通常更低。

这并不意味着必须修改 Gio 底层。当前观察到的主要问题更像 Flux 层缺少：

- 事件降噪。
- 组件级 dirty 标记。
- 可见区和交互区裁剪。
- 动画状态按需创建和释放。
- 大规模组件 benchmark 与性能预算。

## 目标

### 用户体验目标

- 鼠标在大型页面上移动时 CPU 不应持续飙高。
- 未发生视觉变化的鼠标 move 不应触发昂贵重绘。
- 大量不可见组件不应参与 layout、paint 和交互状态推进。
- 高频 hover/pressed 动画可按应用策略降级或关闭。
- 大型业务界面在空闲状态应稳定，不持续产生无意义 frame。

### 工程目标

- 建立稳定的性能基准和 profiling 流程。
- 将性能回归纳入 CI 或本地 smoke test。
- 优先在 Flux 层形成可维护的失效系统。
- 明确何时才值得 fork Gio，避免过早承担底层维护成本。

## 性能预算

以下预算用于桌面应用的初始目标，后续可根据真实产品调整。

| 场景 | 目标 |
| --- | --- |
| 空闲窗口 | 不持续产生 redraw；CPU 接近 0 |
| 1k 可见轻量组件鼠标移动 | 单帧 CPU layout+draw 低于 8ms |
| 5k 总组件、可见区裁剪后鼠标移动 | 单帧 CPU layout+draw 低于 16ms，允许降级 hover 动画 |
| 长列表 10k 项 | 只 layout 可见项和少量 overscan |
| hover target 未变化 | 不触发业务层状态更新，不启动新动画 |
| 大量 disabled 组件 | 不注册无意义 hover/pressed/ripple 动画 |
| overlay/menu/select | 只对打开 overlay 的可见项做交互动画 |

说明：预算中的 5k 指总组件规模，而不是 5k 个文本按钮同时可见并参与交互。2026-07-01 的基线显示 5k 全可见可交互文本按钮约 `390ms/frame`，该场景用于回归对比和压力测试；产品级目标应通过可见区裁剪、虚拟化和 dirty/cache 让每帧成本主要随可见节点增长。

## 测量体系

性能优化的第一阶段必须先补测量工具。没有 profile 的优化只能作为低风险局部清理，不能进入架构重构。

### 必需基准

新增 `benchmarks` 或 `internal/perf` 类测试场景：

- `BenchmarkLayoutStaticTree_1k_5k_10k`
- `BenchmarkMouseMoveInteractiveTree_1k_5k`
- `BenchmarkHoverTargetChange`
- `BenchmarkListVirtualized_10k`
- `BenchmarkTextHeavyList`
- `BenchmarkStaticSurfaceCache`
- `BenchmarkMaterialStateLayerIdle`
- `BenchmarkSelectMenuOpenLargeOptions`

基准需要记录：

- frame 总耗时。
- allocations/op。
- layout 耗时。
- widget state/memo map 访问次数。
- animation 状态推进数量。
- text shaping / text layout 耗时。
- input router / clickable 查询耗时。

### Profile 输出

提供本地命令：

```sh
go test ./examples/component_lab -run TestPerfScenario -cpuprofile out/perf/cpu.pprof
go tool pprof out/perf/cpu.pprof
```

建议补充：

- CPU profile。
- heap profile。
- trace。
- frame timeline 日志。
- redraw reason 日志。

### Redraw Reason

新增调试能力，记录每一帧为什么发生：

- `state.Set`
- `RequestFrameRedraw`
- animation running
- pointer hover/pressed/focus change
- pointer move interaction count（原始 move 只作为交互统计，不作为业务层 redraw reason）
- window event
- async complete
- explicit `WindowInvalidate`

示例输出：

```text
frame=153 duration=11.2ms reason=pointer.hover_changed layout=7.8ms draw=2.1ms animations=43 clickable=1280
```

## 优化原则

1. 先减少 frame 数，再减少每帧成本。
2. 先优化 Flux 层，再考虑 Gio 底层。
3. 先优化高频路径：mouse move、hover、scroll、typing、animation。
4. 可见区外组件默认不参与昂贵工作。
5. idle 状态不创建 transient 动画状态。
6. 业务回调不应在 render/layout 路径执行系统调用、IO 或长耗时计算。
7. 性能策略必须可配置，工具型应用允许牺牲部分 hover 动画质感换 CPU。

## 阶段路线

### Phase 0: 建立基线（已完成：2026-07-01）

目标：知道 CPU 高在哪里。

状态：已完成。Phase 0 只建立可重复测量和归因能力，不代表 CPU 高的问题已经完成优化。

任务：

- [x] 新增大型组件 benchmark。
- [x] 新增 pointer move / hover profile 场景。
- [x] 新增 redraw reason 日志。
- [x] 新增 frame 统计结构：layout、draw、animation、state、text、input。
- [x] 在 `examples/component_lab` 和 `examples/docs_browser` 增加性能 smoke 场景。

验收：

- [x] 可以稳定复现“大量组件下鼠标移动 CPU 高”的场景。
- [x] 可以判断热点来自 Flux layout、widget animation、Gio input router、text shaping 还是 paint/op。
- [x] 每次性能修改可以用同一命令对比前后结果。

完成记录：

- 已新增 `internal/perf` benchmark，覆盖 `BenchmarkLayoutStaticTree_1k_5k_10k`、`BenchmarkMouseMoveInteractiveTree_1k_5k`、`BenchmarkHoverTargetChange`、`BenchmarkListVirtualized_10k`、`BenchmarkTextHeavyList`、`BenchmarkMaterialStateLayerIdle`、`BenchmarkSelectMenuOpenLargeOptions`。
- 已新增 runtime frame stats 和 redraw reason：`layout`、`draw`、`animation`、`state`、`text`、`input` 都能输出 count/duration；`state.Set`、`animation.running`、`pointer.hover_changed`、`pointer.pressed_changed`、`pointer.focus_changed`、`window.event`、`async.complete`、`WindowInvalidate` 等路径能记录 redraw reason。Phase 2 后原始 `pointer.move` 改为 interaction 计数，避免无变化 move 被误判为业务层重绘原因。
- 已在 `examples/component_lab` 和 `examples/docs_browser` 增加 `TestPerfScenario`，用于性能 smoke 和 profile。`component_lab` smoke 已能输出类似 `reason=animation.running,pointer.hover_changed layout=10.6036ms text=2.4941ms input=5.5034ms pointer_moves=1 hover_changes=1` 的 frame 记录。
- 可重复命令：
  ```sh
  go test ./examples/component_lab -run TestPerfScenario -count=1 -cpuprofile out/perf/cpu.pprof
  go tool pprof -top -cum out/perf/cpu.pprof

  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=5s -benchmem -count=1 -cpuprofile out/perf/mousemove_1k_cpu.pprof
  go tool pprof -top -cum out/perf/mousemove_1k_cpu.pprof
  ```
- 2026-07-01 在 Windows / i5-1135G7 上的 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 基线：`frame-ns/frame=10043854`、`layout-ns/frame=9347792`、`input-ns/frame=6054720`、`text-ns/frame=2205134`、`draw-ns/frame=571977`、`animations/frame=5001`、`allocs/op=30837`。这些 section 是嵌套粗粒度计时，不应相加为 frame 总耗时。
- 2026-07-01 在同一机器上的 `BenchmarkMouseMoveInteractiveTree_1k_5k/5k` 基线：`frame-ns/frame=390899180`、`layout-ns/frame=383636770`、`input-ns/frame=349626860`、`text-ns/frame=315422550`、`draw-ns/frame=4460570`、`animations/frame=25001`、`allocs/op=1128204`。该结果稳定复现大量组件下鼠标移动 CPU 明显升高。
- `out/perf/mousemove_1k_cpu.pprof` 的累计热点显示：Flux layout 是主热点，`widget.(*flexWidget).Layout` 约 67.48%、`internal.(*Context).LayoutFlex` 约 67.03%、`widget.(*buttonWidget).Layout` 约 61.85%；Gio input/clickable 是次热点，`gioui.org/widget.(*Clickable).Layout` 约 37.27%、`gioui.org/io/input.(*Router).Frame` 约 14.44%、`Router.collect` 约 11.91%；text layout/shaping 明显但不是最大热点，`LayoutText` 约 17.31%、`LabelStyle.Layout` 约 16.32%、`Shaper.LayoutString` 约 3.64%、`Shaper.NextGlyph` 约 4.19%；paint/op 为较小热点，`clip.RRect.Push/Op/Path` 约 5%~6%、`DrawRipple` 约 4.41%、`fillRoundedRect` 约 3.53%；widget animation 不是当前最大 CPU 来源，`md3AnimateColor` 约 6.62%、`md3AnimateFloat` 约 2.09%、实际 `md3AnimatedFloatState.advance` 约 0.11%，但动画状态/key 访问会随组件数放大。
- 验收结论：Phase 0 完成的原因是现在已经有固定复现命令、同一命令可生成前后对比的 benchmark/profile，并且能把鼠标移动 CPU 拆到 Flux layout、Gio input/clickable、text shaping/layout、paint/op、widget animation 几类上；当前首要优化方向应优先落在减少全树 layout/input/text 遍历和减少每帧高频状态/key 访问上。

### Phase 1: 低风险热路径降本（已完成：2026-07-01）

目标：不改架构，先降低 pprof 主链路中的 layout、input、text、key/map 和 animation bookkeeping 浪费。

状态：已完成。Phase 1 已降低 1k 可见可交互按钮鼠标移动场景的热路径成本；它不包含事件降噪、虚拟化或 text/static cache，因此 2k/5k 全可见文本按钮仍会被 layout/input/text 线性放大，后续继续由 Phase 2/3/4 处理。

已开始的方向：

- [x] Runtime legacy memory 按 frame 活跃 key 清理，避免组件和 transient 状态只增不减。
- [x] MD3 state layer 动画按需创建，idle 组件不创建 hover 动画状态，退出后释放。

后续任务：

- [x] disabled 组件跳过 hover/pressed/focus/ripple 查询和动画路径。
- [x] target 为 0 且没有历史动画状态时，所有轻量动画直接返回；保留 pressed/focus 的必要反馈。
- [x] 减少 `ctx.TreePath()` 字符串拼接、`Context.Child/Scope` 临时字符串和 `md3MotionKey` 构造次数。
- [x] 减少 Runtime memory、state、memo、animation map 的高频访问次数，优先批量读取或局部缓存。
- [x] 减少每个 Button/Container 每帧重复调用 `Clicked/Hovered/Pressed/Focused`，形成一次性 interaction snapshot。
- [x] 对 `ctx.Theme().Colors`、Shapes、Density 等高频读取做局部变量缓存。
- [x] 组件内部避免每帧创建临时 slice/map，优先复用 backing array。
- [x] 高频 row/item 组件避免在未 hover/selected/focused 时创建 decoration animation state 和 text/style 临时对象。

验收：

- [x] `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 的 `frame-ns/frame`、`layout-ns/frame`、`input-ns/frame`、`text-ns/frame` 和 `allocs/op` 有可测下降。
- [x] 大型 idle 页面鼠标移动 CPU 有可测下降，且不引入持续 redraw。
- [x] allocations/op 明显下降，尤其是 path/key/map 和组件临时对象相关分配。
- [x] 所有现有视觉和组件测试通过。

完成记录：

- 新增 `ClickableState.Snapshot` / `event.InteractionSnapshot`，`Button`、`ContainerDecoration`、`Checkbox` 在同一帧只批量读取一次 hover/pressed/focus。`BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 的 `input-ops/frame` 从 `8000` 降到 `3000`，原因是每个按钮不再重复调用 `Hovered/Pressed/Focused`。
- disabled 组件现在跳过 hover/pressed/focus 查询、ripple 注册和 disabled state-layer 动画路径；`LayoutButton` 对 disabled 或 nil clickable 不再进入 Gio clickable layout。
- `md3AnimateFloat` / `md3AnimateFloatDirectional` 在 target 为 0 且没有历史动画状态时直接返回，并在退出动画结束后释放状态；`Runtime.memoryValue` 只在 key 命中时标记 active，避免 missing animation key 写入 active map。
- `Button` 的 idle background/foreground/border color animation 改为按需创建：未 hover/pressed/focused 且没有历史动画状态时直接返回目标色，有历史状态时继续完成退出动画并释放。该改动把 1k mouse move 的 `animations/frame` 从 `5001` 降到 `4`，保留 state-layer/ripple/focus 反馈作为必要交互反馈。
- `Row` / `Column` 在构造时预生成 `internal.FlexChild`，避免每帧重建一层 `layout.FlexChild` slice 和闭包；`Text` 向 `LayoutText` 传递已规范化字体，避免文本热路径重复 `TrimSpace/Normalize`。
- 可重复对比命令：
  ```sh
  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=5s -benchmem -count=1
  ```
- 2026-07-01 本轮 Phase 1 开始前同机基线：`frame-ns/frame=12877589`、`layout-ns/frame=12086568`、`input-ns/frame=7760607`、`text-ns/frame=2583978`、`draw-ns/frame=690168`、`animations/frame=5001`、`input-ops/frame=8000`、`B/op=2988276`、`allocs/op=30837`。
- 2026-07-01 完成后同机结果：`frame-ns/frame=8262546`、`layout-ns/frame=8090885`、`input-ns/frame=5948744`、`text-ns/frame=2098454`、`draw-ns/frame=606477`、`animations/frame=4`、`input-ops/frame=3000`、`B/op=2658907`、`allocs/op=23838`。相对本轮基线：frame 约下降 35.8%，layout 约下降 33.1%，input 约下降 23.3%，text 约下降 18.8%，allocs/op 约下降 22.7%。
- 大型 idle/压力场景的验证结论：`BenchmarkMaterialStateLayerIdle` 完成后 `animations/frame=0`，说明 idle animation bookkeeping 不再持续创建 redraw；`go test -tags visual ./examples/material3_showcase -run "Idle|Redraw|Animation" -count=1` 通过，验证 settle 后不引入持续 redraw。2k/5k 全可见文本按钮的 frame 时间仍主要由 text shaping/layout 与全树 layout 线性放大，原因是 Phase 1 没有做 text cache 和可见区裁剪，这部分继续进入 Phase 3/4。
- 已通过验证：`go test ./widget ./internal ./event ./layout`、`go test ./state ./anim ./ui ./router ./app`、`go test ./examples/component_lab ./examples/docs_browser`、`go test -tags visual ./ui -run ^$ -count=1`、`go test -tags visual ./examples/component_lab -count=1`、`go test -tags visual ./examples/docs_browser -count=1`、`go test -tags visual ./examples/material3_showcase -count=1`。

### Phase 2: 交互事件降噪（已完成：2026-07-01）

目标：鼠标 move 不等于一定重绘。

状态：已完成。Phase 2 已把原始 pointer move 和 hover/pressed/focus 语义变化分开：Gio input router 仍会处理 pointer move，但 Flux 层只在交互目标或状态发生变化时记录业务层 redraw reason，并把高频 move 合并为最新位置。

任务：

- [x] 建立 hover target 跟踪。
- [x] 当 pointer move 后 hover target 不变时，不触发业务层 redraw。
- [x] hover enter/leave 只影响旧 target 和新 target。
- [x] 为组件提供 `OnHover` 的 change-only 语义，避免每帧重复回调。
- [x] 区分 pointer move、hover changed、pressed changed、focus changed。
- [x] 对高频 pointer move 做 coalescing，只处理最新位置。

需要注意：

- Gio 本身仍需要 input router 处理事件，但 Flux 层可以避免把无变化 move 扩散成全树状态变化。
- 如果动画仍在运行，仍需 frame redraw，但只应由动画驱动，而不是 move 本身驱动。

验收：

- [x] 鼠标在同一组件内部移动，不应触发业务状态更新。
- [x] 鼠标在空白区域移动，不应触发昂贵 hover 动画。
- [x] hover target 改变时，只启动相关 target 的视觉变化。

完成记录：

- 新增 `internal/interaction.go`，Runtime 每帧维护 hover target、pointer move coalescing 队列和交互统计：`pointer_moves`、`hover_changes`、`pressed_changes`、`focus_changes`、`hover_target`。
- `event.Clickable` 现在绑定 runtime/target 并缓存上一帧 hover/pressed/focus；`HoverChanged` 和组件 `OnHover` 改为 change-only 语义，不再因为鼠标仍停留在同一 target 内部而重复触发业务回调。
- `ContainerDecorationOnHover` 与按钮 hover 回调保持一致，只在 enter/leave 语义变化时触发；空白区域 pointer move 不会创建 hover target，也不会启动昂贵 hover animation。
- `FrameStats` 和 `FormatFrameStats` 已区分原始 pointer move 计数与 `pointer.hover_changed`、`pointer.pressed_changed`、`pointer.focus_changed` redraw reason；动画仍可用既有 `animation.running` 单独驱动 frame redraw。
- `examples/component_lab` 和 `examples/docs_browser` 的性能 smoke 改为通过 `QueuePointerMoveF32` / `CoalescedPointerMove` 只处理最新位置，不再手动把 raw `pointer.move` 记录为业务 redraw reason。
- 新增/更新 benchmark：`BenchmarkPointerMoveSameTargetCoalesced`、`BenchmarkPointerMoveBlankAreaCoalesced` 和 `BenchmarkMouseMoveInteractiveTree_1k_5k` 会输出 pointer/hover/pressed/focus change 指标，可用同一命令对比事件降噪前后：
  ```sh
  go test ./internal/perf -run ^$ -bench "Benchmark(PointerMoveSameTargetCoalesced|PointerMoveBlankAreaCoalesced)$" -benchtime=5s -benchmem -count=1
  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=5s -benchmem -count=1
  ```
- 2026-07-01 同机验证结果：`BenchmarkPointerMoveSameTargetCoalesced` 为 `pointer-moves/frame=1`、`hover-changes/frame≈0.001969`、`frame-ns/frame=8419609`、`text-ns/frame=0`；`BenchmarkPointerMoveBlankAreaCoalesced` 为 `pointer-moves/frame=1`、`hover-changes/frame≈0.002342`、`frame-ns/frame=11084149`、`text-ns/frame=0`。这说明高频 move 被合并为最新位置，同一 target 内移动和空白区域移动不会扩散成持续 hover change。
- 2026-07-01 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 验证结果：`frame-ns/frame=11515403`、`layout-ns/frame=11251153`、`input-ns/frame=8026498`、`text-ns/frame=2841764`、`allocs/op=23838`、`pointer-moves/frame=1`、`hover-changes/frame=1`、`pressed-changes/frame=0`、`focus-changes/frame=0`。该场景有意每帧切换 target，因此 hover change 为 1；同一 target 和空白区域场景则验证了不变 move 不触发业务状态更新。
- 已通过验证：`go test ./widget -run "TestHoverCallbacksAreChangeOnly|TestBlankAreaMoves|TestPointerMoveCoalescing|TestContainerDecorationOnHover" -count=1`、`go test ./internal ./event ./widget ./layout -count=1`、`go test ./state ./anim ./ui ./router ./app -count=1`、`go test ./examples/component_lab ./examples/docs_browser -count=1`。

### Phase 3: 可见区虚拟化和布局裁剪（已完成：2026-07-02）

目标：组件数量大时，CPU 随可见项增长，而不是随总项增长。

当前 pprof 说明 5k 全可见可交互文本按钮会把 layout/input/text 线性放大到不可接受水平，因此虚拟化和可见区裁剪应早于全局交互质量策略。

状态：已完成。Phase 3 已把真实大列表、Grid、Menu/Select options 和 docs_browser/component_lab 的长内容路径收口到可见项布局；`BenchmarkMouseMoveInteractiveTree_1k_5k/5k` 保留为全可见压力测试，不再作为产品预算目标。

任务：

- [x] List/Grid 默认使用虚拟化。
- [x] 对所有滚动容器暴露 viewport。
- [x] 长列表只构造、layout、paint 可见项和 overscan 项。
- [x] overlay menu/select 对超大 options 做虚拟化。
- [x] docs/browser 示例中所有长内容区域强制虚拟化。
- [x] 为业务方提供“非虚拟化大列表”诊断 warning。
- [x] 可交互区域也要跟随可见区裁剪，不可见 item 不注册 pointer/key handler。

验收：

- [x] 10k 行列表滚动和 hover 的 CPU 与 100 行同量级。
- [x] Select/Menu 大量选项打开时不全量 layout。
- [x] 不可见 item 不创建 clickable、动画和 text layout 状态。
- [x] `BenchmarkMouseMoveInteractiveTree_1k_5k/5k` 不再作为产品预算目标，只作为全可见压力测试；真实大列表场景以虚拟化 benchmark 为准。

完成记录：

- `ListView` 现在真正尊重 `ListVirtualized(false)`，默认仍走 Gio `layout.List` 虚拟化；虚拟化路径统计实际 builder 调用数量，非虚拟化大列表会记录 `nonvirtual_warnings`。因此可见区外 item 不会执行 row builder，也不会创建 clickable、hover 动画状态或 text layout。
- `GridView` 已按可见行布局并统计实际可见 cell；固定 `Grid` 和显式关闭虚拟化的大列表在超过阈值时输出诊断 warning，提醒业务方避免非虚拟化大列表。
- `ScrollView`、`ListView`、`GridView` 会向子 context 暴露 viewport，并通过 `FrameStats.Virtualization` 输出 `virtual_items`、`virtual_culled`、`virtual_containers`、`viewports`、`nonvirtual_warnings`，用于同一命令对比前后性能。
- `Menu`、`DropdownMenu` 和 `Select` 的大 options popup 改为 lazy `ListView`，打开 1000 项时只布局可见项和少量 overscan，不再先构造全部 option row。
- `examples/component_lab` 的性能 smoke 改成固定高度虚拟化大列表；`examples/docs_browser` 的右侧正文列表和左侧文档菜单都使用 `ListViewElement(..., ListVirtualized(true))`，并在 smoke 中断言 `VisibleItems < TotalItems` 和 `CulledItems > 0`。
- 新增单测覆盖：`TestListViewVirtualizationBuildsVisibleItemsOnly`、`TestGridViewVirtualizationBuildsVisibleCellsOnly`、`TestLargeMenuAndSelectUseVirtualizedOptionLists`、`TestNonVirtualizedLargeListReportsWarning`。
- 可重复对比命令：
  ```sh
  go test ./internal/perf -run ^$ -bench BenchmarkListVirtualized_100_10k -benchtime=3s -benchmem -count=1
  go test ./internal/perf -run ^$ -bench "Benchmark(MenuOpenLargeOptions|SelectMenuOpenLargeOptions)$" -benchtime=3s -benchmem -count=1
  go test ./examples/component_lab ./examples/docs_browser -run TestPerfScenario -count=1
  ```
- 2026-07-02 Windows / i5-1135G7 上的 `BenchmarkListVirtualized_100_10k` 验证结果：100 行为 `frame-ns/frame=214100`、`layout-ns/frame=209335`、`input-ns/frame=125342`、`text-ns/frame=23173`、`virtual-visible/frame=13`、`allocs/op=904`；10k 行为 `frame-ns/frame=204242`、`layout-ns/frame=200012`、`input-ns/frame=123902`、`text-ns/frame=24647`、`virtual-visible/frame=13`、`allocs/op=904`。两者同量级，且 10k 场景每帧裁剪约 9987 项。
- 2026-07-02 `BenchmarkMenuOpenLargeOptions`：1000 options 打开时 `virtual-visible/frame=19`、`virtual-culled/frame=981`、`frame-ns/frame=288264`、`allocs/op=901`；`BenchmarkSelectMenuOpenLargeOptions`：`virtual-visible/frame=13`、`virtual-culled/frame=987`、`frame-ns/frame=288222`、`allocs/op=783`。这证明 overlay options 没有全量 layout。
- 已通过验证：`go test ./widget -run "TestListViewVirtualization|TestGridViewVirtualization|TestLargeMenuAndSelectUseVirtualizedOptionLists|TestNonVirtualizedLargeListReportsWarning" -count=1`、`go test ./internal ./widget ./event ./layout ./ui -count=1`、`go test ./examples/component_lab ./examples/docs_browser -count=1`。`govulncheck ./...` 可完整运行；本轮没有新增依赖，报告项来自 Go 标准库和既有 `x/image` / `x/sys` 依赖链。

### Phase 4: Text/静态绘制 cache 和轻量 dirty（已完成：2026-07-02）

目标：先缓存低风险的昂贵静态工作，让鼠标移动不重复做文本布局和纯绘制。

状态：已完成。Phase 4 已优先缓存不会注册 pointer/key handler 的静态文本和静态 surface paint 宏；交互组件仍每帧注册 input handler，不会因缓存丢事件。更大粒度的 subtree dirty/cache 继续留给 Phase 6。

任务：

- [x] 对静态 text shaping / text layout 做 scoped cache，key 包含内容、字体、字号、约束、locale、line-height。
- [x] 对纯绘制 surface、icon、shape path、静态 decoration 做 cache。
- [x] 对无 input handler 的静态 Element/Widget 增加轻量 dirty 判断。
- [x] 对有 input handler 的节点先不缓存 handler 注册本身，避免丢事件或命中错误。
- [x] 输出 text cache / static paint cache 命中率到 perf diagnostics。

风险：

- [x] Gio immediate-mode 不是 retained UI，缓存 op 时必须保证宏生命周期和输入事件注册仍正确。
- [x] 不能缓存会注册 pointer/key handler 的节点，否则会丢事件或命中错误。
- [x] 先缓存纯绘制和纯文本，再扩展到复杂组件。

验收：

- [x] 静态区域在鼠标移动时不重复做昂贵文本和 decoration 计算。
- [x] clickable 区域仍正确注册事件。
- [x] `BenchmarkTextHeavyList` 和 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 的 `text-ns/frame` 有可测下降。
- [x] profile 能显示缓存命中率。

完成记录：

- 新增 runtime-scoped render cache：每个 cache entry 持有独立 `op.Ops`，避免复用当前 frame `Ops` 后被 reset 导致 `CallOp` 失效；每帧按 active key sweep，避免静态 text/surface cache 只增不减。
- `Context.LayoutText` 对静态文本记录 text layout 宏，key 包含内容、字体族、字体样式、字重、字号、line-height、颜色、对齐、constraints、metric、locale 和文本方向。cache hit 时直接 replay 宏并返回尺寸，不再进入 `PerfText` 计时段。
- `LayoutSurface` 的纯 surface paint 路径缓存背景、渐变、圆角/圆形 clip 和 border shape/path 宏；`Icon` 当前走文本图标占位，因此非交互 icon 由 text cache 覆盖。带 image 或 shadow 的复杂 surface 暂不缓存，避免资源和层级语义风险。
- 轻量 dirty 先落在 Text/Surface 级：key 未变化即复用静态宏，key 变化则 miss 后重建。带 `Clickable` / pointer/key handler 的组件仍每帧执行 handler 注册，只缓存其内部纯文本绘制，因此不会丢事件或命中错误。
- `FrameStats.Cache` 和 `FormatFrameStats` 已输出 `text_cache=hits/total`、`static_paint_cache=hits/total`；`internal/perf` benchmark 也输出 `text-cache-hits/frame`、`text-cache-misses/frame`、`static-paint-cache-hits/frame`、`static-paint-cache-misses/frame`。
- 新增验证：`TestLayoutTextUsesStaticCacheAcrossFrames`、`TestLayoutSurfaceUsesStaticPaintCacheAcrossFrames` 和 `BenchmarkStaticSurfaceCache`。`examples/component_lab` / `examples/docs_browser` 性能 smoke 已改为接受 cache hit 后 `text_ops=0` 的稳定状态，并断言存在 text cache hit。
- 可重复对比命令：
  ```sh
  go test ./internal/perf -run ^$ -bench BenchmarkTextHeavyList -benchtime=3s -benchmem -count=1
  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=3s -benchmem -count=1
  go test ./internal/perf -run ^$ -bench BenchmarkStaticSurfaceCache -benchtime=3s -benchmem -count=1
  go test ./examples/component_lab ./examples/docs_browser -run TestPerfScenario -count=1
  ```
- 2026-07-02 Windows / i5-1135G7 上的 `BenchmarkTextHeavyList` 验证结果：`text-cache-hits/frame=750`、`text-cache-misses/frame=0`、`text-ns/frame=0`、`text-ops/frame=0`。这说明静态文本在 warm frame 后不再重复 text shaping/layout。
- 2026-07-02 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 验证结果：`text-cache-hits/frame=1000`、`text-cache-misses/frame=0`、`text-ns/frame=0`、`text-ops/frame=0`、`input-ops/frame=3000`、`hover-changes/frame=1`。这说明按钮文本 cache 生效，同时 clickable/input handler 仍正常注册。
- 2026-07-02 `BenchmarkStaticSurfaceCache` 验证结果：`static-paint-cache-hits/frame=1000`、`static-paint-cache-misses/frame=0`、`draw-ns/frame=0`、`draw-ops/frame=0`。这说明静态 decoration surface 的纯绘制宏已复用。
- 已通过验证：`go test ./internal -run "TestLayoutTextUsesStaticCacheAcrossFrames|TestLayoutSurfaceUsesStaticPaintCacheAcrossFrames" -count=1`、`go test ./internal ./widget ./event ./layout ./ui -count=1`、`go test ./examples/component_lab ./examples/docs_browser -count=1`。

### Phase 5: 全局交互质量策略（已完成：2026-07-02）

目标：允许大型业务应用主动降低交互视觉成本，作为性能兜底开关，而不是替代事件降噪和虚拟化。

状态：已完成。Phase 5 已提供全局和子树级交互质量策略；默认 `Full` 保持现有 MD3 质感，`Balanced` 将 hover 动画降为瞬切，`LowCPU` 禁用 hover state layer 动画并保留 pressed/focus/selected 的必要反馈。

新增策略建议：

```go
type InteractionQuality int

const (
    InteractionQualityFull InteractionQuality = iota
    InteractionQualityBalanced
    InteractionQualityLowCPU
)
```

策略影响：

- [x] `Full`: 保持完整 hover、pressed、focus、selected 动画。
- [x] `Balanced`: hover 动画缩短或改为瞬切，保留 pressed/focus。
- [x] `LowCPU`: 禁用 hover 动画，只保留 pressed/focus 必要反馈。

配置入口：

- [x] App option。
- [x] Theme option。
- [x] Context provider。
- [x] 测试用环境变量，例如 `FLUXUI_INTERACTION_QUALITY=low_cpu`。

验收：

- [x] 大项目可以一键降低 hover CPU。
- [x] 默认仍保持当前 MD3 质感。
- [x] low CPU 模式不破坏可访问性：focus ring、pressed feedback 仍保留。

完成记录：

- 新增 `theme.InteractionQuality`：`InteractionQualityFull`、`InteractionQualityBalanced`、`InteractionQualityLowCPU`。`theme.Default()` / `theme.New()` 默认仍是 `Full`，可通过 `theme.WithInteractionQuality(...)` 覆盖；测试环境变量 `FLUXUI_INTERACTION_QUALITY=low_cpu` / `balanced` / `full` 会在创建默认主题时生效。
- 新增 App 入口 `app.WithInteractionQuality(...)` 和 `ui.WithInteractionQuality(...)`，大型业务应用可以在启动参数中一键切换全局策略，不需要改组件代码。
- 新增子树入口 `internal.Context.WithInteractionQuality(...)`、`widget.WithInteractionQuality(...)`、`ui.InteractionQualityProviderElement(...)` 和 `ui.UseInteractionQuality(...)`，可对局部页面或重型区域单独降级。
- MD3 交互 helper 已接入策略：`Full` 保持原 hover/pressed/focus/selected 动画；`Balanced` 的 hover state layer 直接返回目标值，不创建动画 state；`LowCPU` 的 hover state layer 目标为 0，不创建 hover 动画 state。pressed、focus ring 和 selected 进度仍走原动画路径，避免破坏键盘可访问性和按压反馈。
- 可重复对比命令：
  ```sh
  FLUXUI_INTERACTION_QUALITY=full go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=2s -benchmem -count=1
  FLUXUI_INTERACTION_QUALITY=low_cpu go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=2s -benchmem -count=1
  ```
- 2026-07-02 Windows / i5-1135G7 上的 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 对比：`full` 为 `animations/frame=4`、`frame-ns/frame=6240396`、`layout-ns/frame=5898473`、`input-ns/frame=3855518`、`text-ns/frame=0`、`allocs/op=23841`；`low_cpu` 为 `animations/frame=0`、`frame-ns/frame=5940911`、`layout-ns/frame=5572109`、`input-ns/frame=3825188`、`text-ns/frame=0`、`allocs/op=20837`。这说明一键 low CPU 模式可以去掉 hover 动画 bookkeeping，并减少相关分配。
- 完成原因：Phase 5 的要求是提供性能兜底策略，而不是替代 Phase 2 的事件降噪或 Phase 3 的虚拟化。本轮已经覆盖 App option、Theme option、Context provider 和测试环境变量四个入口；默认策略仍保持 Full；low CPU 模式只降级 hover 动画/state layer，不移除 focus ring 和 pressed feedback，因此满足“大项目可一键降低 hover CPU”和“不破坏可访问性”的验收。
- 已通过验证：`go test ./theme ./app ./internal ./ui ./widget -count=1`、`go test ./examples/component_lab ./examples/docs_browser -count=1`。

### Phase 6: Runtime path、hook key 和组件级 dirty 深化（已完成：2026-07-02）

目标：在 Phase 1/4 的低风险降本之后，系统性降低深层树、组件 identity 和非 dirty 子树成本。

状态：已完成。Phase 6 已把 Runtime 热路径身份从长字符串迁到结构化 path id / `MemoryKey`，并提供显式静态子树 dirty/cache 入口；debug/panic 仍可按需恢复 readable path。交互节点仍不缓存 handler 注册，避免 Gio immediate-mode 输入注册丢失。

任务：

- [x] 从字符串 tree path 迁移到结构化 path id。
- [x] Context child/scope 不再每次拼接长字符串。
- [x] Hook key 使用整数 slot id 或 interned path。
- [x] `md3MotionKey`、state key、effect key 减少临时字符串。
- [x] 引入组件级 stable identity。
- [x] 为 host widget 保存上一帧 layout/draw 结果摘要。
- [x] 根据 props、theme、constraints、interaction state 判断是否 dirty。
- [x] 对静态 subtree 复用 op macro，但需要小心 Gio `op.Ops` 生命周期。
- [x] 保留 debug 时可还原的 readable path。

验收：

- [x] 大型树 allocations/op 明显下降。
- [x] hook/state 身份稳定性不变。
- [x] panic/debug 信息仍可读。
- [x] 非 dirty 静态 subtree 在鼠标移动时少做 layout/draw。

完成记录：

- 新增结构化 `PathID` 和 runtime path intern 表。`Context.Child(index)` 不再 `strconv.Itoa` 后拼接完整 `root/...` 字符串，`Context.Scope(name)` 也不再每次复制长 path；只有 `TreePath()`、panic 和 diagnostics 需要 readable path 时才通过 runtime debug 表恢复。
- 新增 `internal.MemoryKey`，Runtime memory/effect map 从 `map[string]...` 改为结构化 key。legacy `Persistent(string)` / `UseEffect(string)` 仍保留兼容，但 `state.Use*`、`UseAsync`、legacy `UseMemo` / `UseRef` / `UseAnimated*`、`md3MotionKey` 已走 `PathID + namespace + slot` 或 scoped key，减少 state/effect/animation/key 临时字符串。
- `event.Clickable` 的 hover target 从 readable path 字符串改为 `PathID`；`FrameStats.Interaction.HoverTarget` 输出时再恢复字符串，因此性能诊断仍可读，同时大量 clickable snapshot 不再频繁调用 `TreePath()`。
- React-style component hook store 已继续保留组件级 stable identity 和整数 hook slot；Phase 6 在 legacy fallback 中补齐整数 slot key，保证 hook/state 身份稳定性不变。
- 新增显式静态子树 cache：`widget.Static(child, deps...)`、`ui.Static(child, deps...)`、`ui.StaticElement(child, deps...)`。cache key 包含 path、deps hash、theme 指针、foreground、font、constraints、metric、locale、direction；命中后 replay 独立 `op.Ops` macro 和上一帧尺寸，跳过子树 layout/draw。该入口只用于无 pointer/key handler 的静态子树，交互节点仍每帧注册 handler。
- `FrameStats.Cache` / `FormatFrameStats` / `internal/perf` benchmark 已新增 `static_tree_cache=hits/total`、`static-tree-cache-hits/frame`、`static-tree-cache-misses/frame`。
- 可重复对比命令：
  ```sh
  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=2s -benchmem -count=1
  FLUXUI_INTERACTION_QUALITY=low_cpu go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=2s -benchmem -count=1
  go test ./internal/perf -run ^$ -bench BenchmarkStaticSubtreeCache -benchtime=2s -benchmem -count=1
  go test ./examples/component_lab ./examples/docs_browser -run TestPerfScenario -count=1
  ```
- 2026-07-02 Windows / i5-1135G7 上的 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 默认 Full：`frame-ns/frame=7327254`、`layout-ns/frame=6906468`、`input-ns/frame=4410403`、`text-ns/frame=0`、`animations/frame=4`、`allocs/op=22946`。相对 Phase 5 Full 记录的 `allocs/op=23841` 有可测下降，原因是 path/key/motion/effect/state 字符串临时对象减少。
- 2026-07-02 同机 LowCPU：`frame-ns/frame=6576353`、`layout-ns/frame=6212591`、`input-ns/frame=3933090`、`animations/frame=0`、`allocs/op=19943`。相对 Phase 5 LowCPU 记录的 `allocs/op=20837` 继续下降。
- 2026-07-02 `BenchmarkStaticSubtreeCache`：`static-tree-cache-hits/frame=1`、`static-tree-cache-misses/frame=0`、`text-ns/frame=0`、`draw-ns/frame=0`、`layout-ns/frame=427.9`、`allocs/op=7`。这证明非 dirty 静态子树在 warm frame 后不再执行内部 text/layout/draw。
- 新增验证：`TestStructuredPathIDStableAndReadable`、`TestLayoutStaticSubtreeCacheSkipsChildAcrossFrames`，覆盖结构化 path id 稳定、readable path 可恢复、静态子树 cache hit 跳过 child layout。
- 已通过验证：`go test ./internal ./state ./event ./ui ./widget -count=1`、`go test ./examples/component_lab ./examples/docs_browser -run TestPerfScenario -count=1`。
- 完成原因：Phase 6 的核心不是完全 retained UI 化，而是在不破坏 Gio immediate-mode 输入语义的前提下，先把高频 identity/key/path 成本降下来，并给明确无 input handler 的静态子树提供可控 dirty/cache。当前实现满足 allocations 可测下降、hook/state 身份稳定、debug 可读和非 dirty 静态子树少做 layout/draw 的验收。

### Phase 7: Gio 底层评估（已评估：2026-07-02）

状态：已评估，当前不需要定制 Gio，也不建议 fork Gio。P0-P6 完成后的 profile 仍显示主要成本来自“仍然需要 layout/register 的可见节点数量”和 Flux/Gio immediate-mode layout 链路；Gio input router 是可见成本的一部分，但没有成为无法绕开的主导瓶颈。真实大列表、Menu/Select 和文档浏览场景已经通过 Flux 层虚拟化、text/static cache、事件降噪和交互质量策略把成本收口到可见项，因此当前收益最高的方向仍是 Flux 层 API 约束、诊断 warning、组件级 dirty/cache 深化和减少可见节点 handler 数，而不是承担 Gio fork 维护成本。

只有满足以下条件，才进入定制 Gio 或 fork Gio 评估：

- Flux 层 Phase 0 到 Phase 6 已完成或明确不适用。
- Phase 1 到 Phase 3 已经证明无法继续减少 handler 数、frame 数、可见节点数或 text/layout 调用次数。
- pprof 显示主要 CPU 热点在 Gio input router、op reader、GPU backend 或 text shaper 内部。
- 该热点无法通过 Flux 层减少调用次数绕开。
- 有最小 patch 可以证明收益。
- 团队能承担长期跟进 Gio upstream 的成本。

当前复核结果：

- [x] Flux 层 Phase 0 到 Phase 6 已完成。
- [ ] 已证明无法继续减少 handler 数、frame 数、可见节点数或 text/layout 调用次数。当前仍可以通过 Flux 层继续减少误用 `ScrollView(Column(...))`、继续推广 `ListView/GridView`、增加大静态区 `Static` 包裹、减少可见交互节点 handler 和降低 allocations。
- [ ] pprof 显示 Gio input router、op reader、GPU backend 或 text shaper 已经成为主导且无法绕开。当前不是：全可见 1k 压力场景仍以 layout/button/clickable 注册链路为主；真实虚拟化场景成本随可见项增长。
- [ ] 有最小 Gio patch 能证明收益。当前没有足够证据证明 Gio patch 的收益会超过继续减少 Flux 层调用次数。
- [ ] 团队必须承担 fork Gio 长期维护成本。当前没有必要。

可能评估点：

- input router handler 扫描优化。
- pointer hit-test 缓存。
- op reader 局部处理优化。
- text shaper cache 暴露或扩展。
- backend frame scheduling / invalidate coalescing。

不建议一开始 fork Gio 的原因：

- 维护成本高。
- 容易偏离 upstream。
- 真实瓶颈可能仍在 Flux 层全树重算。
- 修改底层后仍需要 Flux 层 dirty/virtualization 才能解决规模问题。

决策结论：

| 条件 | 决策 |
| --- | --- |
| P0-P6 后当前 profile | 不 fork Gio，不定制 Gio |
| Flux 层 profile 显示全树 layout/animation 为主 | 优化 Flux |
| Gio input router 明确占主导且 Flux 无法减少 handler | 只做最小 Gio patch 实验，不直接 fork |
| text shaping 占主导 | 先加/扩大 Flux text cache，再评估 shaper |
| GPU paint 占主导 | 先减少 op 和静态缓存，再评估 backend |
| 只是任务管理器 CPU 高但无 profile | 不 fork Gio |

评估记录：

- 复核命令：
  ```sh
  go test ./internal/perf -run ^$ -bench "Benchmark(MouseMoveInteractiveTree_1k_5k/1k|ListVirtualized_100_10k|TextHeavyList|SelectMenuOpenLargeOptions|StaticSubtreeCache)$" -benchtime=2s -benchmem -count=1 -cpuprofile out/perf/phase6_current_cpu.pprof
  go test ./internal/perf -run ^$ -bench BenchmarkMouseMoveInteractiveTree_1k_5k/1k -benchtime=2s -benchmem -count=1 -cpuprofile out/perf/phase6_mousemove_1k_cpu.pprof
  go tool pprof -top -cum out/perf/phase6_current_cpu.pprof
  go tool pprof -top -cum out/perf/phase6_mousemove_1k_cpu.pprof
  ```
- 2026-07-02 当前同机 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k`：`frame-ns/frame=7749710`、`layout-ns/frame=7268121`、`input-ns/frame=4628146`、`text-ns/frame=0`、`draw-ns/frame=630021`、`animations/frame=4`、`input-ops/frame=3000`、`text-cache-hits/frame=1000`、`allocs/op=22946`。该场景是 1k 全可见可交互按钮压力测试，仍需要每帧 layout 和注册 1000 个 clickable，因此不代表真实长列表产品预算。
- `phase6_mousemove_1k_cpu.pprof` 累计热点显示：`widget.(*flexWidget).Layout` / `Context.LayoutFlex` / `gioui.org/layout.Flex.Layout` 约 54.96%，`buttonWidget.Layout` 约 46.28%，`Context.LayoutButton` 约 29.75%，`gioui.org/widget.(*Clickable).Layout` 约 29.13%，`gioui.org/io/input.(*Router).Frame` 约 16.32%，`Router.collect` 约 12.19%。这说明 Gio input router 仍重要，但不是唯一或压倒性主因；继续减少可见 handler 数和全可见交互节点数量仍能从 Flux 层绕开。
- `BenchmarkListVirtualized_100_10k` 当前 100 行和 10000 行都只有 `virtual-visible/frame=13`，10000 行 `virtual-culled/frame=9987`、`text-ns/frame=0`、`input-ops/frame=65`，证明真实大列表成本已经随可见项增长，而不是随总项增长。
- `BenchmarkSelectMenuOpenLargeOptions` 当前 `virtual-total/frame=1000`、`virtual-visible/frame=13`、`virtual-culled/frame=987`、`text-ns/frame≈536ns`，说明大 options overlay 也没有全量 layout。
- `BenchmarkTextHeavyList` 当前 `text-cache-hits/frame=750`、`text-ns/frame=0`；`BenchmarkStaticSubtreeCache` 当前 `static-tree-cache-hits/frame=1`、`layout-ns/frame≈565ns`、`allocs/op=7`。这说明 text shaper 和静态绘制不是当前必须通过 Gio 底层解决的瓶颈。
- 评估结论：当前不满足“Gio 底层主导且 Flux 层无法绕开”的条件。短期保持 Gio upstream 依赖，不 fork；如果未来真实应用 profile 显示 `io/input.Router`、op reader、GPU backend 或 text shaper 在已经虚拟化/缓存/降噪后仍占主导，再以最小 patch 方式实验，不直接长期 fork。

## P0-P6 优化收益总结

2026-07-02 完成 Phase 0 到 Phase 6 后，性能优化的核心收益是把“大量组件下鼠标移动 CPU 高”的主要问题从总组件数线性爆炸，压缩到可见节点数、可见交互 handler 数和仍需即时注册的 input 区域数量。真实长列表、Grid、Menu/Select 大 options 已经随可见项增长；1k 全可见可交互按钮仍是压力测试，主要用于观察可见 layout/clickable 注册成本。

### 1k 全可见可交互按钮 mouse move 压力场景

Phase 0 基线：

```text
frame-ns/frame   = 10,043,854
layout-ns/frame  = 9,347,792
input-ns/frame   = 6,054,720
text-ns/frame    = 2,205,134
animations/frame = 5,001
allocs/op        = 30,837
```

P6 后当前结果：

```text
frame-ns/frame   = 7,749,710
layout-ns/frame  = 7,268,121
input-ns/frame   = 4,628,146
text-ns/frame    = 0
animations/frame = 4
allocs/op        = 22,946
```

相对 Phase 0：

- `frame-ns/frame` 下降约 22.8%。
- `layout-ns/frame` 下降约 22.2%。
- `input-ns/frame` 下降约 23.6%。
- `text-ns/frame` 从毫秒级降到 0，静态 text layout/shaping cache 在 warm frame 后命中。
- `animations/frame` 从 5001 降到 4，减少约 99.9%。
- `allocs/op` 从 30837 降到 22946，下降约 25.6%。

如果使用 Phase 1 开始前本轮更差基线 `frame-ns/frame=12877589` 对比，P6 后 `frame-ns/frame=7749710`，下降约 39.8%。这说明 Phase 1 到 Phase 6 的热路径降本、事件降噪、text/static cache、交互质量策略和结构化 key/path 都有累计收益。

### 主要收益来源

- 静态 text layout/shaping cache 生效，`BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 当前 `text-cache-hits/frame=1000`、`text-ns/frame=0`。
- hover/idle 动画状态从每帧几千个推进降到个位数，idle 组件不再创建和推进无意义 hover animation state。
- interaction snapshot 把每个 Button/Container 每帧重复 `Hovered/Pressed/Focused` 查询收敛为一次性读取，`input-ops/frame` 从 8000 降到 3000。
- pointer move 降噪和 coalescing 已区分原始 move 与 hover/pressed/focus change，同一 target 内移动不再扩散成业务层状态更新。
- Runtime path/key 从长字符串热路径迁移到结构化 `PathID` / `MemoryKey`，减少 path/key/map 相关分配。
- 静态 surface paint cache 和显式 `Static` subtree cache 让非交互静态区域在 warm frame 后跳过昂贵 text/layout/draw。

### 真实大列表和大 options 场景

`BenchmarkListVirtualized_100_10k` 当前结果显示 100 行和 10000 行都只构造/layout 可见窗口：

```text
100 rows:   virtual-visible/frame = 13
10000 rows: virtual-visible/frame = 13
```

10000 行场景当前：

```text
virtual-total/frame   = 10,000
virtual-visible/frame = 13
virtual-culled/frame  = 9,987
text-ns/frame         = 0
input-ops/frame       = 65
```

`BenchmarkSelectMenuOpenLargeOptions` 当前 1000 options popup：

```text
virtual-total/frame   = 1,000
virtual-visible/frame = 13
virtual-culled/frame  = 987
text-ns/frame         ≈ 536ns
```

这说明长列表和大 options 不再全量 layout，不可见 item 不创建 clickable、动画或 text layout 状态。真实产品预算应以这些虚拟化场景为准，`BenchmarkMouseMoveInteractiveTree_1k_5k/5k` 继续作为全可见压力测试和回归观察，不作为产品级目标。

### 当前剩余成本

P6 后的 1k 全可见可交互压力测试仍有约 `7.75ms/frame`，主要因为 1000 个可见按钮仍必须 layout 并注册 clickable/input handler。pprof 显示 `LayoutFlex/Flex.Layout`、`buttonWidget.Layout`、`Context.LayoutButton`、`Clickable.Layout` 和 `Router.Frame/collect` 仍是主链路。结论是继续优化应优先减少可见交互节点数、避免误用非虚拟化大滚动内容、扩大静态 subtree 使用范围，而不是 fork Gio。

## 当前已知问题和对应路线

| 问题 | 路线 |
| --- | --- |
| legacy runtime memory 长期增长 | Phase 1，已完成按 frame 活跃 key 清理 |
| hover state-layer idle 组件也创建动画状态 | Phase 1，已完成按需创建和 idle 释放 |
| 鼠标 move 导致大树重算 | Phase 2 已完成事件降噪，Phase 4 已完成静态 text/paint cache，Phase 6 已完成显式静态 subtree dirty/cache |
| 长列表组件多时 CPU 高 | Phase 3，已完成默认虚拟化、可见区裁剪和大 options lazy layout |
| 大型业务界面需要主动降低 hover 动画 CPU | Phase 5，已完成全局/主题/子树级 `Full/Balanced/LowCPU` 策略 |
| 深层 Context path 字符串成本 | Phase 6，已完成结构化 `PathID` / `MemoryKey` |
| 文档示例中系统调用进入 render 路径 | 已在性能审查中归类，继续按 Phase 1 清理 |
| 浏览器 CPU 更低 | 通过 Phase 2/3/4/5/6 模拟浏览器的局部失效、虚拟化、缓存和交互质量降级能力 |

## API 设计建议

### 性能配置

```go
app.Run(root,
    app.WithInteractionQuality(theme.InteractionQualityBalanced),
    app.EnablePerfDiagnostics(true),
)
```

### Redraw reason

```go
ctx.RequestFrameRedrawReason("animation:button-hover")
ctx.RequestRedrawReason("state:set")
```

### 静态区域

```go
ui.Static(ui.Column(...), "props-version")
ui.StaticElement(ui.ColumnElement(...), "props-version")
```

语义：

- deps/theme/constraints 不变时复用布局和绘制结果；业务 props 应放入 deps。
- 内部不能含有需要每帧更新的 pointer/key handler；交互节点必须留在静态 cache 外。

### 虚拟化列表

```go
ui.VirtualList(count, func(ctx *ui.Context, index int) ui.Element {
    return row(index)
})
```

语义：

- 默认只构建可见项。
- 默认 overscan 可配置。
- item key 必须稳定。

## 测试和 CI

新增性能测试分层：

- 单元 benchmark：小范围函数和组件。
- 组件 benchmark：button/list/select/text 等高频组件。
- 场景 benchmark：component_lab/docs_browser/large_tree。
- 视觉回归：确保性能优化不破坏样式。
- 空闲重绘测试：确保没有 idle redraw。

CI 建议：

- 普通 CI 只跑轻量性能 smoke，避免不稳定。
- nightly 或手动 workflow 跑完整 benchmark。
- 记录历史结果，发现明显回归时报警。

## 里程碑

### M1: 可测量（已完成：2026-07-01）

- [x] 完成 benchmark 和 redraw reason。
- [x] 输出第一份大型组件 profile。
- [x] 明确鼠标 CPU 的前 5 个热点：Flux layout、Gio clickable/input router、text layout/shaping、paint/op、widget animation/state bookkeeping。

### M2: 低风险热路径降本（已完成：2026-07-01）

- [x] 降低 `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` 的 frame/layout/input/text 成本。
- [x] 完成 idle 动画状态按需创建，disabled/idle 组件跳过无意义交互查询和动画。
- [x] 减少高频 map/slice/string/key/path 分配。

### M3: 鼠标事件降噪（已完成：2026-07-01）

- [x] hover target 不变不扩散 redraw。
- [x] pointer move coalescing。
- [x] hover enter/leave change-only。

### M4: 可见区裁剪和虚拟化（已完成：2026-07-02）

- [x] List/Grid/Menu/Select 虚拟化收口。
- [x] docs_browser/component_lab 大页面性能稳定。
- [x] 大型页面 CPU 随可见节点增长，而不是随总节点增长。

### M5: Text/静态绘制 cache 原型（已完成：2026-07-02）

- [x] 静态 text/icon/surface cache。
- [x] 非交互静态 subtree 原型。
- [x] 明确哪些节点可缓存，哪些节点必须每帧注册事件。

### M6: 交互质量和 Runtime 深化

- [x] 提供 `Full/Balanced/LowCPU` 交互质量策略，作为大型应用兜底开关。（已完成：2026-07-02）
- [x] 结构化 path/hook key，减少深层树字符串成本。（已完成：2026-07-02）
- [x] 深化组件级 dirty/cache，确保 debug 信息仍可读。（已完成：2026-07-02）

### M7: Gio 底层评估（已评估：2026-07-02）

- [x] 基于 P0-P6 后 profile 决定是否需要 Gio patch。
- [x] 当前结论：不需要定制 Gio，不 fork Gio。
- [x] 后续若真实应用 profile 证明 Gio 底层不可绕开，再做最小 patch 实验，不直接长期 fork。

## 风险

- 缓存过度可能导致事件 handler 丢失。
- hover 降噪可能导致视觉状态不同步。
- 虚拟化可能改变滚动测量和 reach-end 语义。
- 禁用 hover 动画可能影响视觉质感，需要配置分层。
- fork Gio 会带来长期维护成本，应作为最后手段。

## 总结

FluxUI 的性能路线应先补齐 retained UI 常见能力中的关键部分：测量、热路径降本、事件降噪、虚拟化、text/static cache、dirty/cache、transient 状态生命周期。2026-07-01 的 pprof 显示主因是全树 layout/input/text 随组件数线性放大，而不是 Gio 底层或 widget animation 单独主导。2026-07-02 完成 P0-P6 后再次复核，真实虚拟化场景已经随可见项增长，text/static cache 生效；1k 全可见压力测试仍主要受可见 layout/clickable 注册数量影响，Gio input router 是重要成本但不是无法绕开的主导瓶颈。因此当前不定制 Gio，也不 fork Gio。只有当 Flux 层减少 frame 数、可见节点数、handler 数和重复 text/layout 调用后，profiling 仍明确指向 Gio 底层，才值得评估最小 Gio patch。

当前推荐策略：

1. 维持 Phase 0 的 benchmark、pprof 和 redraw reason 作为所有优化的对比基线。
2. 优先降低 Flux layout、Gio clickable/input 查询、text layout 和 key/map/path 热路径成本。
3. 解决 pointer move/hover 降噪，减少无变化鼠标移动扩散成全树重绘。
4. 提前建设可见区裁剪和虚拟化，让大型页面成本随可见节点增长。
5. 建设 text/static paint cache，再逐步推进更深的 dirty/cache。
6. 维持交互质量策略作为大型应用兜底开关，默认保持 Full，重型区域可降级到 Balanced/LowCPU。
7. 当前不 fork Gio；保留最小 Gio patch 作为未来真实 profile 证明必要时的最后手段。
