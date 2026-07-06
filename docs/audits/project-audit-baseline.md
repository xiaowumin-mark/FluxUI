<!-- fluxui-doc-meta
{
  "id": "project_audit_baseline",
  "title": "项目逻辑审查基线记录",
  "category": "工程审查",
  "order": 47,
  "summary": "记录 FluxUI 大项目逻辑审查 Batch 0 的工作区、版本、测试和示例基线，保证后续审查从可复现状态开始。",
  "example": { "id": "event_system_basic" },
  "apis": [
    "internal.Context",
    "internal.Runtime",
    "ui.Element",
    "widget.Widget",
    "event.Event"
  ]
}
-->

# FluxUI 项目逻辑审查基线记录

本文件用于记录 `docs/project-audit-roadmap.md` 和 `docs/project-audit-task-breakdown.md` 中 Batch 0 的基线事实。审查阶段只记录事实、风险和证据，不直接修改运行时代码。

## Batch 0：基线冻结

### 任务编号：A0.1 工作区状态冻结

- 状态：Done
- 日期：2026-07-06 14:22:45 +08:00
- 负责人：Codex
- 关注：Runtime、Docs/Test
- 输入命令：
  - `git branch --show-current`
  - `git rev-parse --short HEAD`
  - `git status --short`
  - `git status --porcelain=v1 -uall`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
- 关联能力：
  - 工作区状态冻结
  - 审查产物记录
  - 后续 A0.2-A0.4 基线记录入口

#### 状态快照

| 项目 | 结果 |
| --- | --- |
| 当前分支 | `main` |
| 当前 HEAD | `6cc4a2c` |
| `git status --short` | 无输出 |
| `git status --porcelain=v1 -uall` | 无输出 |
| 审查前已修改文件 | 无 |
| 审查前已暂存文件 | 无 |
| 审查前未跟踪文件 | 无 |

#### 事实结论

- A0.1 开始前工作区是干净状态，没有已修改、已暂存或未跟踪文件。
- 没有审查前遗留脏文件需要解释或保留。
- 没有审查前遗留新增文件需要解释或保留。
- 本次 A0.1 执行过程中新增 `docs/audits/project-audit-baseline.md`，用途是记录 Batch 0 基线事实，允许继续保留。
- `docs/audits/project-audit-baseline.md` 可作为后续 A0.2 版本和依赖快照、A0.3 核心测试基线、A0.4 示例 smoke 基线的追加记录入口。

#### 允许继续保留的文件

| 文件 | 来源 | 允许保留原因 |
| --- | --- | --- |
| `docs/audits/project-audit-baseline.md` | A0.1 新增审查产物 | 这是工作区状态冻结记录本身，也是 Batch 0 后续基线记录入口。 |

#### 风险

- 工作区状态本身无污染风险：A0.1 前 `git status --short` 和 `git status --porcelain=v1 -uall` 均无输出。
- A0.1 完成后，工作区会出现 `docs/audits/project-audit-baseline.md` 这个新增文件；该文件是预期审查产物，不应被误判为审查前遗留脏文件。
- 由于当前仓库是 Go module，本次会话按 Go 工作区规则执行过初始 `go_vulncheck ./...`。该检查发现多项 Go 标准库及间接依赖相关公告，属于环境安全检查结果，不改变 A0.1 的工作区冻结结论。

#### 验收

- 审查前脏文件：无；解释为工作区初始状态干净。
- 审查前新增文件：无；解释为没有未跟踪文件。
- A0.1 新增文件：`docs/audits/project-audit-baseline.md`；解释为本任务要求的工作区状态记录，允许保留。
- 后续进入 A0.2 前，应将 `docs/audits/project-audit-baseline.md` 视为已知审查产物，而不是新的未知污染源。

#### 后续依赖

- A0.2 版本和依赖快照应继续追加到本文档。
- A0.3 核心测试基线应基于此状态说明区分“审查前已有失败”和“后续引入失败”。
- A0.4 示例 smoke 基线应记录构建或空跑状态，不直接修改示例实现。

### 任务编号：A0.2 版本和依赖快照

- 状态：Done
- 日期：2026-07-06 14:28:06 +08:00
- 负责人：Codex
- 关注：Runtime
- 输入命令：
  - `go version`
  - `go env GOPATH GOMOD`
  - `go list -m all`
- 输入文件：
  - `go.mod`
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
- 关联能力：
  - Go 工具链复现
  - Gio 运行时依赖复现
  - 核心依赖版本冻结

#### 工具链快照

| 项目 | 结果 |
| --- | --- |
| `go version` | `go version go1.25.1 windows/amd64` |
| `go.mod` Go 指令 | `go 1.25.0` |
| `GOPATH` | `E:\go` |
| `GOMOD` | `E:\projects\FluxUI\go.mod` |
| 主模块 | `github.com/xiaowumin-mark/FluxUI` |

#### Gio 和核心依赖版本

| 依赖 | 版本 | 角色 |
| --- | --- | --- |
| `gioui.org` | `v0.9.0` | Gio 主依赖，窗口、布局、绘制和输入运行时基础。 |
| `gioui.org/shader` | `v1.0.8` | Gio shader 间接依赖。 |
| `gioui.org/cpu` | `v0.0.0-20210808092351-bfe733dd3334` | Gio CPU 后端间接依赖。 |
| `github.com/tdewolff/font` | `v0.0.0-20260527091451-1663e68cb8a4` | 字体处理直接依赖。 |
| `github.com/tdewolff/parse/v2` | `v2.8.13` | 解析相关间接依赖。 |
| `github.com/go-text/typesetting` | `v0.3.4` | 文本排版间接依赖。 |
| `github.com/worldiety/material-color-utilities` | `v0.0.0-20250324124753-a84b74640c16` | Material 颜色工具依赖。 |
| `golang.org/x/exp/shiny` | `v0.0.0-20250408133849-7e4ce0ab07d0` | Shiny 后端相关间接依赖。 |
| `golang.org/x/image` | `v0.41.0` | 图像处理依赖。 |
| `golang.org/x/sys` | `v0.33.0` | 系统调用依赖。 |
| `golang.org/x/text` | `v0.37.0` | 文本处理间接依赖。 |
| `github.com/andybalholm/brotli` | `v1.2.1` | 压缩间接依赖。 |
| `github.com/golang/freetype` | `v0.0.0-20170609003504-e2365dfdc4a0` | 字体渲染间接依赖。 |

#### 完整模块清单

以下为 `go list -m all` 的完整输出，作为后续问题复现的模块版本证据。

```text
github.com/xiaowumin-mark/FluxUI
codeberg.org/go-latex/latex v0.3.0
codeberg.org/go-pdf/fpdf v0.12.0
dmitri.shuralyov.com/gpu/mtl v0.0.0-20221208032759-85de2813cf6b
eliasnaur.com/font v0.0.0-20230308162249-dd43949cb42d
gioui.org v0.9.0
gioui.org/cpu v0.0.0-20210808092351-bfe733dd3334
gioui.org/shader v1.0.8
github.com/BurntSushi/freetype-go v0.0.0-20160129220410-b763ddbfe298
github.com/BurntSushi/graphics-go v0.0.0-20160129215708-b43f31a4a966
github.com/BurntSushi/xgb v0.0.0-20210121224620-deaf085860bc
github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046
github.com/ByteArena/poly2tri-go v0.0.0-20170716161910-d102ad91854f
github.com/Kagami/go-avif v0.1.0
github.com/andybalholm/brotli v1.2.1
github.com/benoitkugler/textlayout v0.3.2
github.com/benoitkugler/textprocessing v0.0.6
github.com/go-fonts/latin-modern v0.3.3
github.com/go-gl/glfw/v3.3/glfw v0.0.0-20231223183121-56fa3ac82ce7
github.com/go-text/typesetting v0.3.4
github.com/go-text/typesetting-utils v0.0.0-20260223113751-2d88ac90dae3
github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
github.com/jezek/xgb v1.1.1
github.com/kolesa-team/go-webp v1.0.5
github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef
github.com/srwiley/scanx v0.0.0-20190309010443-e94503791388
github.com/tdewolff/canvas v0.0.0-20260508100355-63a7228e682d
github.com/tdewolff/font v0.0.0-20260527091451-1663e68cb8a4
github.com/tdewolff/minify/v2 v2.24.13
github.com/tdewolff/parse/v2 v2.8.13
github.com/tdewolff/test v1.0.12
github.com/wcharczuk/go-chart/v2 v2.1.2
github.com/worldiety/material-color-utilities v0.0.0-20250324124753-a84b74640c16
github.com/xyproto/randomstring v1.0.5
github.com/yuin/goldmark v1.8.2
golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0
golang.org/x/image v0.41.0
golang.org/x/mobile v0.0.0-20231127183840-76ac6878050a
golang.org/x/mod v0.35.0
golang.org/x/net v0.55.0
golang.org/x/sync v0.20.0
golang.org/x/sys v0.33.0
golang.org/x/text v0.37.0
golang.org/x/tools v0.44.0
gonum.org/v1/plot v0.17.0
modernc.org/knuth v0.5.5
modernc.org/token v1.1.0
star-tex.org/x/tex v0.7.1
```

#### 事实结论

- 当前复现环境使用 Go `1.25.1`，目标平台为 `windows/amd64`。
- 仓库 `go.mod` 声明 Go 语言版本为 `1.25.0`，实际工具链比模块声明高一个补丁版本。
- 当前模块文件为 `E:\projects\FluxUI\go.mod`，后续测试和示例 smoke 应以该模块根目录为复现入口。
- Gio 主版本冻结为 `gioui.org v0.9.0`；涉及 Gio 后端和 shader 的间接版本也已记录。
- 与 Runtime、绘制、文本、样式和文档示例相关的核心依赖已记录到表格和完整模块清单中。

#### 风险

- Go 工具链补丁版本与 `go.mod` 指令不同：`go.mod` 为 `1.25.0`，本机为 `1.25.1`。后续复现问题时应同时记录实际 `go version`，避免把补丁版本差异误判为代码行为变化。
- `go list -m all` 包含由示例、文档、绘图、字体和 Gio 后端拉入的间接依赖；后续审查若只看 `go.mod` 直接依赖，可能遗漏运行时或示例构建差异。
- A0.2 只冻结版本事实，不判断依赖是否应该升级，也不修改 `go.mod` 或 `go.sum`。

#### 验收

- Go 工具链版本已记录：`go version go1.25.1 windows/amd64`。
- GOPATH 和 GOMOD 已记录：`E:\go`、`E:\projects\FluxUI\go.mod`。
- Gio 主依赖版本已记录：`gioui.org v0.9.0`。
- 完整 `go list -m all` 输出已记录，后续问题可以按同一模块版本集合复现。

#### 后续依赖

- A0.3 核心测试基线应使用本文记录的 Go 工具链和模块集合解释测试结果。
- A0.4 示例 smoke 基线应在同一 GOMOD 环境下执行，避免跨模块或外部 workspace 污染。
- 后续任何依赖升级建议应作为单独修复候选登记，不应混入审查基线任务。

### 任务编号：A0.3 核心测试基线

- 状态：Done
- 日期：2026-07-06 14:31:38 +08:00
- 负责人：Codex
- 关注：Docs/Test
- 输入命令：
  - `git status --short --untracked-files=all`
  - `go test ./event ./internal ./ui ./widget`
- 输入文件：
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
  - `docs/audits/project-audit-baseline.md`
- 关联能力：
  - 核心包测试基线
  - 既有失败识别
  - 后续审查回归参照

#### 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A0.3 执行前工作区干净，没有待解释的代码或文档污染。 |

#### 测试命令

```powershell
go test ./event ./internal ./ui ./widget
```

#### 测试输出

```text
ok  	github.com/xiaowumin-mark/FluxUI/event	0.467s
ok  	github.com/xiaowumin-mark/FluxUI/internal	0.450s
ok  	github.com/xiaowumin-mark/FluxUI/ui	2.607s
ok  	github.com/xiaowumin-mark/FluxUI/widget	1.054s
```

#### 通过/失败清单

| 包 | 结果 | 耗时 | 既有问题判断 |
| --- | --- | ---: | --- |
| `github.com/xiaowumin-mark/FluxUI/event` | 通过 | `0.467s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/internal` | 通过 | `0.450s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/ui` | 通过 | `2.607s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/widget` | 通过 | `1.054s` | 无失败项，不涉及既有失败归类。 |

#### 事实结论

- 指定核心测试命令 `go test ./event ./internal ./ui ./widget` 执行成功，退出码为 0。
- `event`、`internal`、`ui`、`widget` 四个核心包全部通过。
- 本次 A0.3 未发现失败项，因此没有需要标注为既有问题的测试失败。
- A0.3 执行前工作区干净；本次审查只追加基线文档，不改动运行时代码或测试代码。

#### 风险

- 本任务只覆盖文档指定的核心包测试命令，不代表全仓库 `go test ./...` 已通过。
- 当前结果只能证明 A0.3 时点核心包在 Go `1.25.1 windows/amd64` 和 A0.2 记录的模块集合下通过。
- 后续若出现这些包的失败，应先与本条基线对比，再判断是否由后续审查或修复引入。

#### 验收

- 通过/失败清单已记录：四个核心包全部通过。
- 失败项标注已记录：无失败项，因此没有既有失败被误判为审查引入问题。
- 后续问题可以回到 A0.2 的版本快照和本条测试输出复现核心测试基线。

#### 后续依赖

- A0.4 示例 smoke 基线可以在核心测试通过的前提下继续执行。
- 若后续审查修改 `event`、`internal`、`ui` 或 `widget`，应至少重跑受影响包测试，并与本基线对照。
