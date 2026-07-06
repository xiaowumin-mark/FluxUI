# A1.1 包依赖方向图

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 1：包边界和 runtime 基础。

- 状态：Done
- 日期：2026-07-06 14:39:19 +08:00
- 负责人：Codex
- 关注：Runtime
- 输入命令：
  - `git status --short --untracked-files=all`
  - `go list ./...`
  - `go list -deps ./internal ./internal/perf ./event ./ui ./widget ./layout ./style ./theme ./system`
  - `go list -f '{{.ImportPath}}|{{range .Imports}}{{.}} {{end}}' ./internal ./internal/perf ./event ./ui ./widget ./layout ./style ./theme ./system`
  - `rg "github\\.com/xiaowumin-mark/FluxUI/(internal|event|ui|widget|layout|style|theme|system)" --glob "*.go" internal event ui widget layout style theme system`
- 输入文件：
  - `internal/*.go`
  - `internal/perf/*.go`
  - `event/*.go`
  - `ui/*.go`
  - `widget/*.go`
  - `layout/*.go`
  - `style/*.go`
  - `theme/*.go`
  - `system/*.go`
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
- 关联能力：
  - 包依赖方向
  - runtime 所有权边界
  - 反向依赖识别

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A1.1 执行前工作区干净，没有待解释的代码或文档污染。 |

## 目标包确认

| 审查维度 | Go 包 |
| --- | --- |
| `internal` | `github.com/xiaowumin-mark/FluxUI/internal`、`github.com/xiaowumin-mark/FluxUI/internal/perf` |
| `event` | `github.com/xiaowumin-mark/FluxUI/event` |
| `ui` | `github.com/xiaowumin-mark/FluxUI/ui` |
| `widget` | `github.com/xiaowumin-mark/FluxUI/widget` |
| `layout` | `github.com/xiaowumin-mark/FluxUI/layout` |
| `style` | `github.com/xiaowumin-mark/FluxUI/style` |
| `theme` | `github.com/xiaowumin-mark/FluxUI/theme` |
| `system` | `github.com/xiaowumin-mark/FluxUI/system` |

## 生产直接依赖方向图

以下只记录 `internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system` 之间的生产 import；测试 import 单独列为备注。

```text
theme
system
style   -> theme
internal -> style, theme
layout  -> internal
event   -> internal
widget  -> event, internal, layout, style, theme
ui      -> event, internal, widget, style, theme, system
```

## 直接 import 证据摘录

| 包 | 目标范围内的生产直接 import |
| --- | --- |
| `internal` | `style`、`theme` |
| `internal/perf` | 无生产 import |
| `event` | `internal` |
| `ui` | `event`、`internal`、`style`、`system`、`theme`、`widget` |
| `widget` | `event`、`internal`、`layout`、`style`、`theme` |
| `layout` | `internal` |
| `style` | `theme` |
| `theme` | 无 |
| `system` | 无 |

## 允许依赖

| 方向 | 判断 | 理由 |
| --- | --- | --- |
| `event -> internal` | 允许 | 事件类型和分发需要复用 runtime path、target 或基础状态能力。 |
| `layout -> internal` | 允许但需保持窄接口 | 布局组件当前依赖 `internal.Context` 或尺寸相关基础类型；后续 A3 应确认是否只依赖最小 runtime 能力。 |
| `style -> theme` | 允许 | 样式层可使用 theme 的颜色、字体和 Material 语义。 |
| `widget -> event/internal/layout/style/theme` | 允许 | widget 是组合层，需要 runtime、事件、布局、样式和主题能力。 |
| `ui -> event/internal/widget/style/theme/system` | 允许 | `ui` 是公开 Element/Builder 聚合层，当前承担对 widget、runtime、事件和系统 API 的门面职责。 |
| `theme ->` 无目标包依赖 | 允许 | theme 当前保持基础语义层，不反向依赖 runtime/widget/ui。 |
| `system ->` 无目标包依赖 | 允许 | system 当前独立于 UI/runtime 包边界，适合作为平台能力层。 |

## 禁止依赖

| 方向 | 当前是否观察到 | 说明 |
| --- | --- | --- |
| `event -> widget/ui/system` | 未观察到 | event 不应调用控件层、公开 UI 门面或系统能力。 |
| `layout -> event/widget/ui/style/theme/system` | 未观察到 | layout 不应知道事件、控件、样式主题或系统 API；当前只直接依赖 `internal`。 |
| `style -> internal/event/widget/ui/system` | 未观察到 | style 不应依赖 runtime、事件、控件或公开 UI 门面；当前只直接依赖 `theme`。 |
| `theme -> internal/event/ui/widget/layout/style/system` | 未观察到 | theme 应保持最底层语义资源，不应回依赖任何 FluxUI 上层包。 |
| `system -> internal/event/ui/widget/layout/style/theme` | 未观察到 | system 应保持平台能力边界，不应反向持有 UI/runtime 语义。 |
| `widget -> ui/system` | 未观察到 | widget 不应依赖公开 UI 门面或平台系统包，避免组件实现反向污染门面层或系统层。 |
| `internal -> event/ui/widget/layout/system` | 未观察到 | runtime 不应反向调用事件公开层、控件层、布局层、UI 门面或系统能力。 |

## 疑似反向依赖

| 方向 | 状态 | 证据 | 风险 |
| --- | --- | --- | --- |
| `internal -> style` | 已观察到，疑似反向依赖 | `internal/ripple.go`、`internal/render.go`、`internal/render_cache.go` import `github.com/xiaowumin-mark/FluxUI/style` | `internal` 是 runtime 基础层，依赖 style 会把渲染/状态基础设施和高层样式结构耦合；后续 A2/A4 需要确认 style 数据是否应下沉为基础类型或改由上层传入。 |
| `internal -> theme` | 已观察到，疑似反向依赖 | `internal/context.go`、`internal/runtime.go`、`internal/render.go`、`internal/render_cache.go` import `github.com/xiaowumin-mark/FluxUI/theme` | runtime 直接持有 theme 语义可能让 frame/context 生命周期受主题系统影响；后续 A2 应确认 theme 是否是 Context 公共能力，A4 应确认默认样式来源。 |
| `internal/perf` test -> `widget/style` | 测试专用，暂不计入生产反向依赖 | `internal/perf/bench_test.go` import `internal`、`style`、`widget` | 当前不污染生产图；但性能基准若作为 runtime 指标，应避免长期依赖高层 widget 作为唯一场景。 |

## 传递依赖观察

- `go list -deps` 显示 `internal` 的传递依赖中包含 `style`、`theme`，这来自生产直接依赖。
- `event` 和 `layout` 虽然只直接依赖 `internal`，但会因 `internal -> style/theme` 间接获得 `style/theme` 传递依赖。
- `ui` 作为聚合层传递依赖覆盖 `event`、`internal`、`layout`、`style`、`system`、`theme`、`widget`，符合公开门面层特征，但后续 A1.2 仍需明确 public API 所有权。
- 生产依赖未形成 Go import cycle；当前风险不是编译循环，而是底层 runtime 语义和高层 style/theme 语义之间的所有权耦合。

## 事实结论

- 目标包的生产依赖方向总体呈现：`theme/style/internal/event/layout/widget/ui` 分层向上组合，`system` 独立，`ui` 作为公开聚合层。
- 当前最明显的边界风险是 `internal -> style/theme`，与"runtime 基础层应避免被样式/主题层反向约束"的审查主旨存在张力，已标为疑似反向依赖。
- 未观察到 `event -> widget/ui`、`widget -> ui/system`、`style/theme/system -> runtime/widget/ui` 等更严重反向依赖。
- A1.1 只记录依赖事实和风险，不调整 import、不拆包、不修改运行时代码。

## 验收

- 已输出 `internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system` 的依赖方向图。
- 已标出允许依赖、禁止依赖和疑似反向依赖。
- 已区分生产 import 和测试 import，避免把 `internal/perf` 的 benchmark 依赖误判为 runtime 生产依赖。

## 后续依赖

- A1.2 public API 所有权清单应重点确认 `internal.Context`、`style.Decoration`、`theme.Theme` 等跨层类型的所有权。
- A2 runtime 生命周期审查应确认 `internal.Context` 和 `internal.Runtime` 是否应该直接持有 theme/style 语义。
- A4 渲染、样式和动画审查应确认 `internal` 中 ripple/render/cache 对 `style/theme` 的依赖是否需要收敛到单向边界。
