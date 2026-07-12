# 高级组件 Benchmark 与 benchstat 模板

高级组件只在能说明工作量与可见项/用户路径的关系时才有生产级 benchmark。不要用一次短跑替代固定环境的趋势比较。

## 基线环境

| 项目 | 记录值 |
| --- | --- |
| 组件 / benchmark 名 | `Benchmark<Component><Scenario>` |
| commit（baseline / candidate） |  |
| Go / Gio | `1.25.12` / `v0.9.0` |
| OS / CPU / 核数 / 内存 |  |
| 电源策略 / governor / 后台负载 |  |
| `-count` / `-benchtime` / `-benchmem` |  |
| 数据规模 / 可见项 / overscan |  |
| 帧预算 | `8 ms` 或 `16 ms`，并说明场景 |

## 场景矩阵

| 场景 | 输入规模 | 必测指标 | 通过条件 |
| --- | --- | --- | --- |
| cold layout | 大集合/复杂表单首次构建 | ns/op、allocs/op、B/op | 不与总数据量无界线性退化。 |
| warm layout/paint | 缓存稳定后的连续 frame | ns/op、allocs/op、idle redraw | 无无意义持续 redraw。 |
| 输入热路径 | pointer move、wheel、keyboard/筛选 | ns/op、allocs/op、帧预算 | 不越过 8/16 ms。 |
| virtualized range | 大总数、固定 visible + overscan | 构建项数量、ns/op | 与 visible 项相关。 |
| Overlay/scroll（适用时） | placement、滚动后命中 | ns/op、allocs/op | 不复制全量布局或注册。 |

## 命令与比较

```powershell
# 用相同机器和参数分别采集 baseline / candidate，次数与 benchtime 必须记录。
go test ./<package> -run '^$' -bench 'Benchmark<Component>' -benchmem -count=<n> -benchtime=<duration> | Tee-Object baseline.txt
go test ./<package> -run '^$' -bench 'Benchmark<Component>' -benchmem -count=<n> -benchtime=<duration> | Tee-Object candidate.txt
benchstat baseline.txt candidate.txt
```

## 结果与决策

| 指标 | Baseline | Candidate | 变化 | 门槛 | 结论 |
| --- | --- | --- | --- | --- | --- |
| ns/op |  |  |  | 不得恶化 >10% |  |
| allocs/op |  |  |  | 说明新增分配 |  |
| B/op |  |  |  | 说明新增分配 |  |
| 帧时间 |  |  |  | ≤8 ms / ≤16 ms |  |

热点回归超过 10% 或越过场景帧预算必须阻断；若确需例外，记录批准人、理由、影响、到期日、补救计划和可回滚 commit。将 benchstat 输出或 artifact 链接附在 PR/release checklist。
