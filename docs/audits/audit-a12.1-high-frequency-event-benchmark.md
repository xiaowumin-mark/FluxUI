# A12.1 高频事件 benchmark 设计

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，记录 A12.1 高频事件 benchmark 设计。

## 事实结论

A12.1 的目标是为 pointer move、wheel、hover、scroll 建立可复现的 benchmark 场景和预算口径，用来衡量高频输入是否产生明显分配热点。当前仓库已经有两类可复用入口：`widget/pointer_area_bench_test.go` 负责低层 `PointerArea` 的 Gio 输入到 FluxUI 事件桥接；`internal/perf/bench_test.go` 负责大组件树、hover target 变化、pointer move 合并和 frame diagnostics 指标聚合。

| 高频输入 | 当前可用 benchmark | 覆盖链路 | 当前缺口 |
| --- | --- | --- | --- |
| pointer move | `BenchmarkPointerAreaHighFrequencyMove`、`BenchmarkMouseMoveInteractiveTree_1k_5k`、`BenchmarkPointerMoveSameTargetCoalesced`、`BenchmarkPointerMoveBlankAreaCoalesced` | Gio `pointer.Move`/`pointer.Drag`、`PointerEventFromGio`、`DispatchPointerEvent`、runtime event path、interaction diagnostics | 缺少按 listener 数、path 深度、capture/bubble listener 组合拆分的纯 event dispatch benchmark |
| wheel | `BenchmarkPointerAreaHighFrequencyWheel` | Gio `pointer.Scroll`、`WheelEventFromGio`、`DispatchWheelEvent`、`PointerOnWheel` | 缺少 `ScrollView`/`ListView` offset 更新、父子滚动、剩余 delta 和横向 delta 的专项 benchmark |
| hover | `BenchmarkHoverTargetChange`、`BenchmarkMouseMoveInteractiveTree_1k_5k` | pointer move 后 hover target 变化、`InteractionFrameStats.HoverChanged`、button/tree hit routing | 缺少“同一目标内移动不变 hover”和“空白区移动不应产生 hover storm”的预算断言 |
| scroll | `BenchmarkListVirtualized_10k`、`BenchmarkListVirtualized_100_10k` 间接覆盖滚动集合；`BenchmarkPointerAreaHighFrequencyWheel` 覆盖 wheel 事件桥接 | 虚拟列表可见窗口、pointer input routing、frame diagnostics | 缺少真实 wheel 驱动 `ScrollView`/`ListView` offset 变化的端到端 benchmark |

### 建议 benchmark 场景

| 编号 | 建议名称 | 包 | 场景 | 测量指标 | 预算口径 |
| --- | --- | --- | --- | --- | --- |
| HF-01 | `BenchmarkPointerAreaHighFrequencyMove` | `./widget` | 单个 `PointerArea` 每帧 8 个 move 合并为一次 `PointerEvent` | `ns/op`、`B/op`、`allocs/op`、handler 看到的 `Coalesced` 数 | 桥接层不应随事件批量线性增加 handler 调用；单帧分配应主要来自 Gio/router 与 coalesced copy |
| HF-02 | `BenchmarkPointerAreaHighFrequencyWheel` | `./widget` | 单个 `PointerArea` 连续 wheel 输入 | `ns/op`、`B/op`、`allocs/op` | wheel 事件转换和分发应低于 pointer move 合并场景；不应触发布局树级放大 |
| HF-03 | `BenchmarkPointerMoveSameTargetCoalesced` | `./internal/perf` | 大交互树内同一 target 连续 pointer move | `pointer-moves/frame`、`hover-changes/frame`、`input-ns/frame`、`allocs/op` | `pointer-moves/frame` 应稳定为 1；同 target 移动的 `hover-changes/frame` 应接近 0 |
| HF-04 | `BenchmarkHoverTargetChange` | `./internal/perf` | 1024 个交互 target 之间跳转 hover | `hover-changes/frame`、`input-ns/frame`、`layout-ns/frame`、`allocs/op` | 每次跨 target 移动最多记录一次 hover change；分配随可见/注册 target 增长可观察 |
| HF-05 | `BenchmarkWheelScrollViewVertical` | 待新增 `./widget` 或 `./internal/perf` | wheel 驱动纵向 `ScrollView` offset 更新 | `offset-delta/frame`、`ScrollOnChange` 次数、`allocs/op` | wheel 可消费时 offset 必须变化；callback 只在真实 offset 变化后触发 |
| HF-06 | `BenchmarkWheelNestedScrollRemainingDelta` | 待新增 `./widget` 或 `./internal/perf` | 子滚动区到边界后向父级传递剩余 delta | 父/子 offset、defaultPrevented、`allocs/op` | 子级不可滚动时不应无条件截停父级滚动；剩余 delta 需要可观测 |
| HF-07 | `BenchmarkHorizontalWheelDelta` | 待新增 `./widget` 或 `./internal/perf` | 横向 `ScrollView` 处理 `DeltaX`，纵向 wheel 不误触发横向滚动 | x/y offset、callback 次数、`allocs/op` | 纵向 delta 不应改变 x offset；真实 `DeltaX` 才改变横向 offset |
| HF-08 | `BenchmarkPointerDispatchPathDepth` | 待新增 `./event` 或 `./internal` | 纯 event target path 深度和 listener 数扩展 | listener calls、path len、`ns/op`、`allocs/op` | path/listener 扩展成本应可线性解释；无 listener 时不应产生显著分配热点 |

### 当前短跑结果

本轮用短 benchtime 验证现有入口可执行：

```powershell
go test ./widget ./internal/perf -run '^$' -bench 'Benchmark(PointerAreaHighFrequency|PointerMove|HoverTarget|MouseMoveInteractiveTree)' -benchmem -benchtime=20x
```

| 包 | benchmark | 结果摘要 |
| --- | --- | --- |
| `./widget` | `BenchmarkPointerAreaHighFrequencyMove` | 约 `465210 ns/op`，`6341 B/op`，`67 allocs/op` |
| `./widget` | `BenchmarkPointerAreaHighFrequencyWheel` | 约 `13275 ns/op`，`2975 B/op`，`23 allocs/op` |
| `./internal/perf` | `BenchmarkMouseMoveInteractiveTree_1k_5k/1k` | 约 `60.2 ms/op`，`3139273 B/op`，`25947 allocs/op`；`pointer-moves/frame=1`，`hover-changes/frame=1` |
| `./internal/perf` | `BenchmarkMouseMoveInteractiveTree_1k_5k/5k` | 约 `118.3 ms/op`，`15693536 B/op`，`129948 allocs/op`；`pointer-moves/frame=1`，`hover-changes/frame=1` |
| `./internal/perf` | `BenchmarkHoverTargetChange` | 约 `23.0 ms/op`，`3213512 B/op`，`26571 allocs/op`；`pointer-moves/frame=1`，`hover-changes/frame=1` |
| `./internal/perf` | `BenchmarkPointerMoveSameTargetCoalesced` | 约 `20.7 ms/op`，`2994185 B/op`，`24931 allocs/op`；`pointer-moves/frame=1`，`hover-changes/frame=0.05` |
| `./internal/perf` | `BenchmarkPointerMoveBlankAreaCoalesced` | 约 `19.6 ms/op`，`2994198 B/op`，`24931 allocs/op`；`pointer-moves/frame=1`，`hover-changes/frame=0.05` |

这些数值是本机短跑 smoke，不作为跨机器绝对性能基线；正式预算应使用固定 Go/Gio 版本、固定 `-benchtime`、固定 CPU governor 和多轮 `benchstat` 比较。

### 建议预算

| 层级 | 预算项 | 建议阈值/判定 |
| --- | --- | --- |
| 桥接层 | `PointerArea` move/wheel `allocs/op` | 任何优化或修复不应让 `allocs/op` 较当前同机基线增长超过 20%；新增字段或 diagnostics 需要单独解释 |
| 事件层 | 纯 event dispatch | 应新增无 Gio/router 的 `event`/`internal` benchmark，目标是区分 runtime dispatch 自身分配与 Gio router 分配 |
| 交互树 | 1k/5k 交互 target | `pointer-moves/frame` 应稳定为 1；hover 变化应与 target 变化次数一致，不随 move 样本数放大 |
| 滚动层 | `ScrollView`/`ListView` wheel | wheel 可消费时 offset 必须变化；不可消费时应能观察父级传递或未消费状态 |
| 横向滚动 | `DeltaX`/`DeltaY` 轴策略 | 横向滚动 benchmark 必须同时报告 x/y offset，避免纵向 wheel 被误计为横向消费 |
| 诊断层 | `FrameStats` 指标 | 每个 benchmark 应至少报告 `allocs/op`，Perf benchmark 还应报告 `input-ns/frame`、`layout-ns/frame`、`pointer-moves/frame`、`hover-changes/frame` |

## 风险

- 现有 `BenchmarkPointerAreaHighFrequencyMove` 的 move 合并会执行 `append([]PointerSample(nil), coalesced...)`，它能表达合并语义，但也会把 coalesced copy 计入分配；后续优化需要区分“语义必要 copy”和“意外临时分配”。
- `internal/perf` 的大树 benchmark 同时包含 layout、draw、text cache、Gio router 和 event routing，适合观察用户级成本，但不适合作为纯事件分发热点归因。
- wheel 当前只有 `PointerArea` 低层桥接 benchmark，缺少 `ScrollView` offset、`ListView` 虚拟化滚动、父子滚动和横向滚动策略的端到端基准。
- hover 当前通过 pointer move 间接测量，缺少直接断言“同 target 高频 move 不反复触发 hover change”的稳定预算。
- 短 benchtime 结果波动较大，不应直接写入 release gate；需要后续用 `benchstat` 比较多轮结果。
- diagnostics 打开后会改变测量成本；需要保留 diagnostics on/off 两组 benchmark，避免把观测开销误判为业务路径开销。

## 验收

- 已列出 pointer move、wheel、hover、scroll 四类高频输入的 benchmark 场景设计和预算口径。
- 已确认当前已有 `./widget` 和 `./internal/perf` 两组相关 benchmark 入口，可作为 A12.1 初始 smoke。
- 已用短 benchtime 执行现有高频事件 benchmark，证明入口可运行并能输出 `allocs/op` 与 frame diagnostics 指标。
- 已标出 scroll 端到端、父子滚动、横向 delta、纯 event dispatch path/listener 扩展等缺口。
- 已明确本轮仅记录 benchmark 设计和预算，不新增 benchmark 代码、不修改 runtime/widget 行为。

## 后续依赖

- A6.2/A6.3/A6.4：wheel 分发、横向滚动策略、滚动后命中刷新变更后，需要补齐 HF-05/HF-06/HF-07。
- A5.2/A5.5：EventTarget 分发和 default action 可取消性变更后，需要补齐 HF-08，并分别覆盖 capture、target、bubble、passive、preventDefault。
- A12.2：大组件树 benchmark 应复用 `internal/perf` 的 `frameHarness` 和 `frameStatsTotal`，避免 A12.1/A12.2 指标口径分裂。
- A12.3：diagnostics 字段审查需要确认 benchmark 能记录 event path、target、defaultPrevented、redraw reason，同时保留 diagnostics off 的对照组。
- A13.x：若后续收敛事件、hover、scroll 冗余逻辑，应以 A12.1 benchmark 作为分配和高频输入回归入口。
