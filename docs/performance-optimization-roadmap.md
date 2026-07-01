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
| 5k 可见轻量组件鼠标移动 | 单帧 CPU layout+draw 低于 16ms，允许降级 hover 动画 |
| 长列表 10k 项 | 只 layout 可见项和少量 overscan |
| hover target 未变化 | 不触发业务层状态更新，不启动新动画 |
| 大量 disabled 组件 | 不注册无意义 hover/pressed/ripple 动画 |
| overlay/menu/select | 只对打开 overlay 的可见项做交互动画 |

## 测量体系

性能优化的第一阶段必须先补测量工具。没有 profile 的优化只能作为低风险局部清理，不能进入架构重构。

### 必需基准

新增 `benchmarks` 或 `internal/perf` 类测试场景：

- `BenchmarkLayoutStaticTree_1k_5k_10k`
- `BenchmarkMouseMoveInteractiveTree_1k_5k`
- `BenchmarkHoverTargetChange`
- `BenchmarkListVirtualized_10k`
- `BenchmarkTextHeavyList`
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
- pointer hover change
- pointer move
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

### Phase 0: 建立基线

目标：知道 CPU 高在哪里。

任务：

- 新增大型组件 benchmark。
- 新增 pointer move / hover profile 场景。
- 新增 redraw reason 日志。
- 新增 frame 统计结构：layout、draw、animation、state、text、input。
- 在 `examples/component_lab` 和 `examples/docs_browser` 增加性能 smoke 场景。

验收：

- 可以稳定复现“大量组件下鼠标移动 CPU 高”的场景。
- 可以判断热点来自 Flux layout、widget animation、Gio input router、text shaping 还是 paint/op。
- 每次性能修改可以用同一命令对比前后结果。

### Phase 1: 低风险热路径清理

目标：不改架构，先降低明显浪费。

已开始的方向：

- Runtime legacy memory 按 frame 活跃 key 清理，避免组件和 transient 状态只增不减。
- MD3 state layer 动画按需创建，idle 组件不创建 hover 动画状态，退出后释放。

后续任务：

- disabled 组件跳过 hover/pressed/ripple 动画路径。
- target 为 0 且没有历史动画状态时，所有轻量动画直接返回。
- 减少 `ctx.TreePath()` 字符串拼接和 `md3MotionKey` 构造次数。
- 对 `ctx.Theme().Colors`、Shapes、Density 等高频读取做局部变量缓存。
- 组件内部避免每帧创建临时 slice/map，优先复用 backing array。
- 高频 row/item 组件避免在未 hover/selected/focused 时创建 decoration animation state。

验收：

- 大型 idle 页面鼠标移动 CPU 有可测下降。
- allocations/op 明显下降。
- 所有现有视觉和组件测试通过。

### Phase 2: 交互事件降噪

目标：鼠标 move 不等于一定重绘。

任务：

- 建立 hover target 跟踪。
- 当 pointer move 后 hover target 不变时，不触发业务层 redraw。
- hover enter/leave 只影响旧 target 和新 target。
- 为组件提供 `OnHover` 的 change-only 语义，避免每帧重复回调。
- 区分 pointer move、hover changed、pressed changed、focus changed。
- 对高频 pointer move 做 coalescing，只处理最新位置。

需要注意：

- Gio 本身仍需要 input router 处理事件，但 Flux 层可以避免把无变化 move 扩散成全树状态变化。
- 如果动画仍在运行，仍需 frame redraw，但只应由动画驱动，而不是 move 本身驱动。

验收：

- 鼠标在同一组件内部移动，不应触发业务状态更新。
- 鼠标在空白区域移动，不应触发昂贵 hover 动画。
- hover target 改变时，只启动相关 target 的视觉变化。

### Phase 3: 全局交互质量策略

目标：允许大型业务应用主动降低交互视觉成本。

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

- `Full`: 保持完整 hover、pressed、focus、selected 动画。
- `Balanced`: hover 动画缩短或改为瞬切，保留 pressed/focus。
- `LowCPU`: 禁用 hover 动画，只保留 pressed/focus 必要反馈。

配置入口：

- App option。
- Theme option。
- Context provider。
- 测试用环境变量，例如 `FLUXUI_INTERACTION_QUALITY=low_cpu`。

验收：

- 大项目可以一键降低 hover CPU。
- 默认仍保持当前 MD3 质感。
- low CPU 模式不破坏可访问性：focus ring、pressed feedback 仍保留。

### Phase 4: 可见区虚拟化和布局裁剪

目标：组件数量大时，CPU 随可见项增长，而不是随总项增长。

任务：

- List/Grid 默认使用虚拟化。
- 对所有滚动容器暴露 viewport。
- 长列表只构造、layout、paint 可见项和 overscan 项。
- overlay menu/select 对超大 options 做虚拟化。
- docs/browser 示例中所有长内容区域强制虚拟化。
- 为业务方提供“非虚拟化大列表”诊断 warning。

验收：

- 10k 行列表滚动和 hover 的 CPU 与 100 行同量级。
- Select/Menu 大量选项打开时不全量 layout。
- 不可见 item 不创建 clickable、动画和 text layout 状态。

### Phase 5: 组件级 dirty/cache

目标：非 dirty 子树少做工作。

任务：

- 引入组件级 stable identity。
- 为 host widget 保存上一帧 layout/draw 结果摘要。
- 根据 props、theme、constraints、interaction state 判断是否 dirty。
- 对静态 text、icon、decoration、surface 做缓存。
- 对文本 shaping 和 layout 做 scoped cache。
- 对静态 subtree 复用 op macro，但需要小心 Gio `op.Ops` 生命周期。

风险：

- Gio immediate-mode 不是 retained UI，缓存 op 时必须保证宏生命周期和输入事件注册仍正确。
- 不能缓存会注册 pointer/key handler 的节点，否则会丢事件或命中错误。
- 先缓存纯绘制和纯文本，再扩展到复杂组件。

验收：

- 静态区域在鼠标移动时不重复做昂贵文本和 decoration 计算。
- clickable 区域仍正确注册事件。
- profile 能显示缓存命中率。

### Phase 6: Runtime path 和 hook key 优化

目标：减少深层树和大量组件下的字符串 key 成本。

任务：

- 从字符串 tree path 迁移到结构化 path id。
- Context child/scope 不再每次拼接长字符串。
- Hook key 使用整数 slot id 或 interned path。
- `md3MotionKey`、state key、effect key 减少临时字符串。
- 保留 debug 时可还原的 readable path。

验收：

- 大型树 allocations/op 明显下降。
- hook/state 身份稳定性不变。
- panic/debug 信息仍可读。

### Phase 7: Gio 底层评估

只有满足以下条件，才进入定制 Gio 或 fork Gio 评估：

- Flux 层 Phase 0 到 Phase 4 已完成或明确不适用。
- pprof 显示主要 CPU 热点在 Gio input router、op reader、GPU backend 或 text shaper 内部。
- 该热点无法通过 Flux 层减少调用次数绕开。
- 有最小 patch 可以证明收益。
- 团队能承担长期跟进 Gio upstream 的成本。

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
| Flux 层 profile 显示全树 layout/animation 为主 | 优化 Flux |
| Gio input router 明确占主导且 Flux 无法减少 handler | 评估 Gio patch |
| text shaping 占主导 | 先加 Flux text cache，再评估 shaper |
| GPU paint 占主导 | 先减少 op 和静态缓存，再评估 backend |
| 只是任务管理器 CPU 高但无 profile | 不 fork Gio |

## 当前已知问题和对应路线

| 问题 | 路线 |
| --- | --- |
| legacy runtime memory 长期增长 | Phase 1，已开始修复 |
| hover state-layer idle 组件也创建动画状态 | Phase 1，已开始修复 |
| 鼠标 move 导致大树重算 | Phase 2 + Phase 5 |
| 长列表组件多时 CPU 高 | Phase 4 |
| 深层 Context path 字符串成本 | Phase 6 |
| 文档示例中系统调用进入 render 路径 | 已在性能审查中归类，继续按 Phase 1 清理 |
| 浏览器 CPU 更低 | 通过 Phase 2/4/5 模拟浏览器的局部失效和虚拟化能力 |

## API 设计建议

### 性能配置

```go
app.Run(root,
    app.InteractionQuality(app.InteractionQualityBalanced),
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
ui.Static(ui.ColumnElement(...))
```

语义：

- props/theme/constraints 不变时尽量复用布局和绘制结果。
- 内部不能含有需要每帧更新的 pointer handler，或必须显式声明。

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

### M1: 可测量

- 完成 benchmark 和 redraw reason。
- 输出第一份大型组件 profile。
- 明确鼠标 CPU 的前 5 个热点。

### M2: 低风险降本

- 完成 idle 动画状态按需创建。
- disabled/idle 组件跳过无意义交互动画。
- 减少高频 map/slice/string 分配。

### M3: 鼠标事件降噪

- hover target 不变不扩散 redraw。
- pointer move coalescing。
- hover enter/leave change-only。

### M4: 大列表和大型页面

- List/Grid/Menu/Select 虚拟化收口。
- docs_browser/component_lab 大页面性能稳定。

### M5: dirty/cache 原型

- 静态 text/icon/surface cache。
- 非交互静态 subtree 原型。
- 明确哪些节点可缓存，哪些节点必须每帧注册事件。

### M6: Gio 底层评估

- 基于 profile 决定是否需要 Gio patch。
- 如果需要，先做最小实验，不直接长期 fork。

## 风险

- 缓存过度可能导致事件 handler 丢失。
- hover 降噪可能导致视觉状态不同步。
- 虚拟化可能改变滚动测量和 reach-end 语义。
- 禁用 hover 动画可能影响视觉质感，需要配置分层。
- fork Gio 会带来长期维护成本，应作为最后手段。

## 总结

FluxUI 的性能路线应先补齐 retained UI 常见能力中的关键部分：测量、事件降噪、虚拟化、dirty/cache、transient 状态生命周期。只有当这些工作完成后，profiling 仍明确指向 Gio 底层，才值得定制 Gio。

当前推荐策略：

1. 立即推进 Flux 层性能专项。
2. 建立大型 UI benchmark 和 redraw reason。
3. 优先解决鼠标 move/hover 高频路径。
4. 中期建设虚拟化和 dirty/cache。
5. 最后再评估 Gio patch，不提前 fork。
