# A12.2 大组件树 benchmark 设计

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A12.2 大组件树 benchmark 设计。

## 事实结论

A12.2 的目标是为大列表、大表格、多卡片、多交互控件建立 layout、paint、event registration 成本测量方案，并能把 layout 成本和 event 成本拆开观察。当前仓库已有 `internal/perf/bench_test.go` 作为主要入口：它通过 `frameHarness` 在固定 `gioLayout.Context`、`op.Ops` 和 `input.Router` 下反复布局组件树，并通过 `FrameStats` 汇总 `layout-ns/frame`、`draw-ns/frame`、`input-ns/frame`、`layout-ops/frame`、`draw-ops/frame`、`input-ops/frame`、虚拟化、文本缓存和静态缓存指标。

### 现有 benchmark 覆盖

| 场景 | 当前 benchmark | 组件模型 | 是否 route input | 可区分成本 |
| --- | --- | --- | --- | --- |
| 大静态布局树 | `BenchmarkLayoutStaticTree_1k_5k_10k` | 1k/5k/10k 个 `Padding(Row(Text, Spacer, Text))` | 否 | layout/text/cache 成本；无 input/event registration |
| 多交互控件 | `BenchmarkMouseMoveInteractiveTree_1k_5k` | 1k/5k 个 `Button(Text)` | 是 | layout、draw、input routing、hover change、event target registration 混合成本 |
| 虚拟大列表 | `BenchmarkListVirtualized_10k`、`BenchmarkListVirtualized_100_10k` | 100/10k 项 `ListView(ListItem)`，viewport 约 13 项 | 是 | 可见项、裁剪项、虚拟容器数、input routing 和 layout 成本 |
| 文本密集列表 | `BenchmarkTextHeavyList` | 750 行长文本 | 否 | layout 与 text cache 命中成本 |
| 多卡片/装饰面 | `BenchmarkStaticSurfaceCache` | 1000 个带背景、padding、radius、border 的 `ContainerDecoration` | 否 | decoration layout 与 static paint cache 命中成本 |
| 静态子树缓存 | `BenchmarkStaticSubtreeCache` | 1000 项静态树包在 `Static` 中 | 否 | `Static` subtree cache 命中后 layout 退化成本 |
| 大菜单/选择项 | `BenchmarkMenuOpenLargeOptions`、`BenchmarkSelectMenuOpenLargeOptions` | 1000 项 menu/select popup | 是 | overlay 内列表、option row、input routing 和 hover 成本 |

### 成本拆分方案

| 成本 | 主要指标 | 设计口径 | 需要避免的误判 |
| --- | --- | --- | --- |
| layout 成本 | `layout-ns/frame`、`layout-ops/frame`、`ns/op`、`allocs/op` | 用 `routeInput=false` 的静态树/文本树/装饰树作为纯 layout 近似基线；对比 1k/5k/10k 增长曲线 | `FrameStats.Layout` 当前由 harness 包住整棵 `w.Layout`，部分 draw/text 发生在 layout 调用内，不能当作严格 CPU profiler 分区 |
| paint/draw 成本 | `draw-ns/frame`、`draw-ops/frame`、`static-paint-cache-hits/frame`、`static-paint-cache-misses/frame` | 用 `StaticSurfaceCache` 和未来多卡片 benchmark 观察装饰绘制、缓存命中和 misses | 许多 Gio paint op 在组件 `Layout` 中发出，若没有显式 `StartFrameSection(PerfDraw)`，`draw-ns/frame` 可能低估实际绘制准备成本 |
| event registration 成本 | `input-ns/frame`、`input-ops/frame`、`pointer-moves/frame`、`hover-changes/frame`、`allocs/op` | 用同结构的静态树和交互树对照；`routeInput=true` 时 `input-ops/frame` 可观察 clickable/pointer target 数量级 | input 成本包含 Gio router、hover 状态和部分事件处理，不等于纯 listener 注册；纯 event dispatch 仍需 A12.1/HF-08 补测 |
| 虚拟化成本 | `virtual-visible/frame`、`virtual-culled/frame`、`virtual-total/frame`、`virtual-warnings/frame` | 大列表必须证明 visible 项接近 viewport 数量，不随 total 线性布局 | `ListView` 当前覆盖列表，不覆盖大表格多列布局、动态列宽和 cell 级事件 target |
| 缓存收益 | `text-cache-hits/frame`、`text-cache-misses/frame`、`static-tree-cache-hits/frame`、`static-paint-cache-hits/frame` | 对比 `LayoutStaticTree`、`TextHeavyList`、`StaticSubtreeCache`、`StaticSurfaceCache` | 短跑 warmup 可能影响 cache hit/miss；正式结果应固定预热帧 |

### 建议 benchmark 场景

| 编号 | 建议名称 | 包 | 场景 | 关键指标 | 验收口径 |
| --- | --- | --- | --- | --- | --- |
| LT-01 | `BenchmarkLayoutStaticTree_1k_5k_10k` | 已有 `./internal/perf` | 纯静态大树，无输入路由 | `layout-ns/frame`、`layout-ops/frame`、`allocs/op` | layout ops 与节点数线性对应；无 `input-ops/frame` |
| LT-02 | `BenchmarkMouseMoveInteractiveTree_1k_5k` | 已有 `./internal/perf` | 大量 Button，pointer move 触发 hover | `layout-ns/frame`、`draw-ns/frame`、`input-ns/frame`、`input-ops/frame` | 与 LT-01 对比可观察交互注册和输入路由成本 |
| LT-03 | `BenchmarkListVirtualized_100_10k` | 已有 `./internal/perf` | 100 与 10k 项同 viewport 列表 | `virtual-visible/frame`、`virtual-culled/frame`、`layout-ops/frame` | total 增长时 visible/layout ops 不应按 total 线性增长 |
| LT-04 | `BenchmarkStaticSurfaceCache` | 已有 `./internal/perf` | 1000 个装饰面/卡片近似场景 | `static-paint-cache-hits/frame`、`layout-ns/frame`、`allocs/op` | 多卡片装饰成本和缓存命中可观测 |
| LT-05 | `BenchmarkStaticSubtreeCache` | 已有 `./internal/perf` | 大静态子树缓存命中 | `static-tree-cache-hits/frame`、`layout-ops/frame`、`allocs/op` | cache hit 后 layout ops 应接近 1 |
| LT-06 | `BenchmarkLargeTable_100x50_1000x20` | 待新增 | 大表格，固定列宽/动态列宽、可横向滚动 | `layout-ns/frame`、`layout-ops/frame`、可见 cell 数、`allocs/op` | 能区分全量 cell 布局和 viewport cell 布局 |
| LT-07 | `BenchmarkCardGrid_1k_5k` | 待新增 | 多卡片网格，含 text、decoration、image placeholder | `layout-ns/frame`、`draw-ops/frame`、cache hit/miss | 多卡片不应因装饰/阴影导致全树重绘预算失控 |
| LT-08 | `BenchmarkInteractiveRegistrationOnly_1k_5k` | 待新增 | 与 `interactiveTree` 同结构，但不投递 pointer move | `input-ops/frame`、`layout-ns/frame`、`allocs/op` | 隔离“注册大量交互 target”与“处理 pointer move”的成本 |

### 当前短跑结果

本轮用短 benchtime 验证现有大树 benchmark 可执行：

```powershell
go test ./internal/perf -run '^$' -bench 'Benchmark(LayoutStaticTree|ListVirtualized|StaticSurfaceCache|StaticSubtreeCache|TextHeavyList|MouseMoveInteractiveTree)' -benchmem -benchtime=5x
```

| benchmark | 结果摘要 |
| --- | --- |
| `BenchmarkLayoutStaticTree_1k_5k_10k/1k` | 约 `16.7 ms/op`，`6.21 MB/op`，`35909 allocs/op`；`layout-ops/frame=2002`，`input-ops/frame=0` |
| `BenchmarkLayoutStaticTree_1k_5k_10k/5k` | 约 `122.6 ms/op`，`31.0 MB/op`，`179909 allocs/op`；`layout-ops/frame=10002`，`input-ops/frame=0` |
| `BenchmarkLayoutStaticTree_1k_5k_10k/10k` | 约 `149.1 ms/op`，`62.1 MB/op`，`359909 allocs/op`；`layout-ops/frame=20002`，`input-ops/frame=0` |
| `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` | 约 `16.7 ms/op`，`3.33 MB/op`，`25950 allocs/op`；`layout-ops/frame=2002`，`input-ops/frame=3000` |
| `BenchmarkMouseMoveInteractiveTree_1k_5k/5k` | 约 `101.2 ms/op`，`16.6 MB/op`，`129956 allocs/op`；`layout-ops/frame=10002`，`input-ops/frame=15000` |
| `BenchmarkListVirtualized_10k` | 约 `0.64 ms/op`，`106 KB/op`，`915 allocs/op`；`virtual-visible/frame=13`，`virtual-culled/frame=9987` |
| `BenchmarkListVirtualized_100_10k/100` | 约 `0.70 ms/op`，`106 KB/op`，`915 allocs/op`；`virtual-visible/frame=13`，`virtual-culled/frame=87` |
| `BenchmarkListVirtualized_100_10k/10000` | 约 `1.44 ms/op`，`106 KB/op`，`915 allocs/op`；`virtual-visible/frame=13`，`virtual-culled/frame=9987` |
| `BenchmarkTextHeavyList` | 约 `3.75 ms/op`，`1.69 MB/op`，`8909 allocs/op`；`text-cache-hits/frame=750` |
| `BenchmarkStaticSurfaceCache` | 约 `5.12 ms/op`，`2.45 MB/op`，`11909 allocs/op`；`static-paint-cache-hits/frame=1000` |
| `BenchmarkStaticSubtreeCache` | 约 `0.096 ms/op`，`1472 B/op`，`7 allocs/op`；`static-tree-cache-hits/frame=1`，`layout-ops/frame=1` |

这些结果是本机短跑 smoke，仅用于确认入口和指标可用。正式预算应固定 Go/Gio 版本、CPU、电源策略、`-benchtime`、预热帧和 `benchstat` 比较方式。

## 风险

- 当前大树 benchmark 已能覆盖静态树、交互树、虚拟列表、文本密集、装饰面和静态缓存，但没有真实“大表格”模型；表格的列宽计算、横向滚动、cell 级事件 target 和 sticky header 仍是缺口。
- `FrameStats.Layout` 由 harness 包住整棵 `w.Layout`，可用于同口径对比，但不能完全等同于严格的 layout-only CPU 时间；绘制 op、文本处理和部分状态动画可能发生在 layout 调用内部。
- `input-ops/frame` 能反映大量交互控件注册量级，但它混合了 Gio input ops、clickable/hover state 和 pointer routing，不是纯 listener registry 计数。
- `BenchmarkMouseMoveInteractiveTree` 同时包含 event registration 和 pointer move 处理；还需要 LT-08 这类“不投递输入但保留交互 target”的对照组。
- 多卡片目前用 `StaticSurfaceCache` 近似，缺少真实 card grid 中 image placeholder、不同高度、阴影、ripple 和 focus ring 的组合成本。
- 短跑结果中静态 10k 树未严格按 5k 比例增长，说明短 benchtime/缓存/机器调度会带来波动；不能据此直接判断优化或退化。

## 验收

- 已输出大列表、大表格、多卡片、多交互控件的 benchmark 场景设计和预算口径。
- 已确认现有 `internal/perf` benchmark 可以区分 `routeInput=false` 的 layout 近似基线与 `routeInput=true` 的交互/input routing 成本。
- 已记录 `layout-ns/frame`、`draw-ns/frame`、`input-ns/frame`、`layout-ops/frame`、`draw-ops/frame`、`input-ops/frame`、虚拟化和缓存指标的使用方式。
- 已用短 benchtime 执行现有大树 benchmark，确认入口可运行并能输出可用于拆分成本的指标。
- 已标出大表格、真实 card grid、registration-only 对照组和严格 paint 分区的缺口。
- 已确认本轮只记录 Perf/Layout benchmark 设计，不新增 benchmark 代码、不修改 runtime/widget 行为。

## 后续依赖

- A12.1：高频事件 benchmark 的 pointer/wheel/hover 指标应继续与 A12.2 的大树交互场景共用 `frameStatsTotal` 口径。
- A12.3：diagnostics 能力审查需要补齐 event registration、event path、redraw reason 与 frame stats 的关联字段。
- A3/A6：布局、滚动、横向溢出或虚拟化策略变更后，应优先回归 LT-01、LT-03、LT-06。
- A4/A10：样式、state layer、ripple、Card、Button、Select、Menu 等组件变更后，应回归 LT-02、LT-04、LT-07、`BenchmarkMenuOpenLargeOptions` 和 `BenchmarkSelectMenuOpenLargeOptions`。
- A13.x：冗余收敛或缓存策略调整后，应以 LT-01/LT-04/LT-05 的同机 `benchstat` 结果作为回归依据。
