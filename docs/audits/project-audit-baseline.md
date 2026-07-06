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

### 任务编号：A0.4 示例 smoke 基线

- 状态：Done
- 日期：2026-07-06 14:36:13 +08:00
- 负责人：Codex
- 关注：Docs/Test
- 输入命令：
  - `git status --short --untracked-files=all`
  - `rg --files examples`
  - `go test -run '^$' ./examples/docs_browser ./examples/component_lab ./examples/event_system_testbench ./examples/drag_drop_showcase`
- 输入文件：
  - `examples/docs_browser/README.md`
  - `examples/component_lab/README.md`
  - `examples/event_system_testbench/README.md`
  - `examples/drag_drop_showcase/README.md`
  - `docs/project-audit-roadmap.md`
  - `docs/project-audit-task-breakdown.md`
- 关联能力：
  - 示例 smoke 编译基线
  - docs browser 手动验收入口
  - component lab 手动验收入口
  - event system testbench 手动验收入口
  - drag/drop showcase 手动验收入口

#### 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A0.4 执行前工作区干净，没有待解释的代码或文档污染。 |

#### 示例入口确认

| 示例 | 路径 | README 状态 | 运行入口 |
| --- | --- | --- | --- |
| docs browser | `examples/docs_browser` | 有 | `go run ./examples/docs_browser` |
| component lab | `examples/component_lab` | 有 | `go run ./examples/component_lab` |
| event system testbench | `examples/event_system_testbench` | 有 | `go run ./examples/event_system_testbench` |
| drag drop showcase | `examples/drag_drop_showcase` | 有 | `go run ./examples/drag_drop_showcase` |

#### Smoke 命令

```powershell
go test -run '^$' ./examples/docs_browser ./examples/component_lab ./examples/event_system_testbench ./examples/drag_drop_showcase
```

本次使用 `go test -run '^$'` 做空跑编译 smoke：编译包和测试文件，但不运行测试用例、不启动 GUI 窗口。

#### Smoke 输出

```text
ok  	github.com/xiaowumin-mark/FluxUI/examples/docs_browser	0.170s [no tests to run]
ok  	github.com/xiaowumin-mark/FluxUI/examples/component_lab	0.737s [no tests to run]
?   	github.com/xiaowumin-mark/FluxUI/examples/event_system_testbench	[no test files]
?   	github.com/xiaowumin-mark/FluxUI/examples/drag_drop_showcase	[no test files]
```

#### 示例可运行状态表

| 示例 | Smoke 状态 | 测试文件状态 | 后续手动验收入口 | 主要覆盖 |
| --- | --- | --- | --- | --- |
| `examples/docs_browser` | 通过 | 有测试文件；本次空跑未执行测试用例 | 是 | 交互式文档主入口，Markdown 文档、示例注册、右侧预览、搜索、分类、API 索引、控件/布局/系统 API demo。 |
| `examples/component_lab` | 通过 | 有测试文件；本次空跑未执行测试用例 | 是 | 组件视觉和交互实验室，覆盖主题、密度、表单、选择、导航、overlay、list、grid、drag/drop、router、Ref 控制。 |
| `examples/event_system_testbench` | 通过 | 无测试文件 | 是 | 事件系统手动测试台，覆盖 P0-P7：旧回调、新事件分发、pointer、focus/keyboard、text input、默认行为、portal、诊断。 |
| `examples/drag_drop_showcase` | 通过 | 无测试文件 | 是 | 生产拖放 API 展示，覆盖 DragSource、DropTarget、payload、生命周期、系统能力探测。 |

#### 后续手动验收入口

| 优先级 | 入口 | 手动验收用途 |
| --- | --- | --- |
| P0 | `go run ./examples/docs_browser` | 后续 A1-A13 的文档和示例总验收入口，尤其用于确认文档页滚动、代码块、表格、tabs、chips、示例弹窗和搜索交互。 |
| P0 | `go run ./examples/component_lab` | 后续布局、样式、hover、cursor、组件状态、Ref 和 overlay 组合行为的视觉/交互 smoke 入口。 |
| P0 | `go run ./examples/event_system_testbench` | 后续 A5-A7 事件、焦点、键盘、文本输入、默认行为和诊断链路的手动验收入口。 |
| P1 | `go run ./examples/drag_drop_showcase` | 后续 drag/drop、pointer 冲突、scroll 冲突和系统后端能力的专项验收入口。 |

#### 事实结论

- 四个指定示例包均通过空跑编译 smoke，命令退出码为 0。
- 本次未发现示例编译失败，因此没有需要标注为既有问题的失败项。
- `docs_browser` 和 `component_lab` 的测试文件能参与编译；测试用例被 `-run '^$'` 显式跳过，符合本任务的“构建或空跑测试”范围。
- `event_system_testbench` 和 `drag_drop_showcase` 当前没有测试文件；空跑结果证明包可编译，但不等价于 GUI 手动流程已验证。
- 四个示例都应作为后续手动验收入口保留，其中 `docs_browser`、`component_lab`、`event_system_testbench` 为 P0，`drag_drop_showcase` 为 P1。

#### 风险

- 空跑编译 smoke 不启动 GUI，不能验证窗口创建、真实鼠标键盘输入、拖放后端、系统通知或平台特性。
- `drag_drop_showcase` 的外部拖入/拖出行为依赖操作系统和桌面后端，本次仅确认包可编译。
- `event_system_testbench` 的 P0-P7 行为需要人工操作观察日志，本次仅冻结可编译入口。
- 后续若示例 smoke 失败，应先与本条基线对比，避免把平台手动行为差异误判为编译回归。

#### 验收

- 示例可运行状态表已记录：四个指定示例 smoke 均通过。
- 后续手动验收入口已明确：`docs_browser`、`component_lab`、`event_system_testbench`、`drag_drop_showcase`。
- 本次未发现失败项，没有既有失败被误判为审查引入问题。
- A0.4 未修改示例实现、测试代码或运行时代码。

#### 后续依赖

- Batch 1 以后如修改文档示例注册、事件系统、组件状态、overlay、scroll 或 drag/drop，应重跑对应示例 smoke，并按需启动对应手动验收入口。
- G0 关卡可以基于 A0.1-A0.4 判断：工作区、版本、核心测试和示例 smoke 均已记录。

## Batch 1：包边界和 runtime 基础

### 任务编号：A1.1 包依赖方向图

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

#### 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A1.1 执行前工作区干净，没有待解释的代码或文档污染。 |

#### 目标包确认

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

#### 生产直接依赖方向图

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

#### 直接 import 证据摘录

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

#### 允许依赖

| 方向 | 判断 | 理由 |
| --- | --- | --- |
| `event -> internal` | 允许 | 事件类型和分发需要复用 runtime path、target 或基础状态能力。 |
| `layout -> internal` | 允许但需保持窄接口 | 布局组件当前依赖 `internal.Context` 或尺寸相关基础类型；后续 A3 应确认是否只依赖最小 runtime 能力。 |
| `style -> theme` | 允许 | 样式层可使用 theme 的颜色、字体和 Material 语义。 |
| `widget -> event/internal/layout/style/theme` | 允许 | widget 是组合层，需要 runtime、事件、布局、样式和主题能力。 |
| `ui -> event/internal/widget/style/theme/system` | 允许 | `ui` 是公开 Element/Builder 聚合层，当前承担对 widget、runtime、事件和系统 API 的门面职责。 |
| `theme ->` 无目标包依赖 | 允许 | theme 当前保持基础语义层，不反向依赖 runtime/widget/ui。 |
| `system ->` 无目标包依赖 | 允许 | system 当前独立于 UI/runtime 包边界，适合作为平台能力层。 |

#### 禁止依赖

| 方向 | 当前是否观察到 | 说明 |
| --- | --- | --- |
| `event -> widget/ui/system` | 未观察到 | event 不应调用控件层、公开 UI 门面或系统能力。 |
| `layout -> event/widget/ui/style/theme/system` | 未观察到 | layout 不应知道事件、控件、样式主题或系统 API；当前只直接依赖 `internal`。 |
| `style -> internal/event/widget/ui/system` | 未观察到 | style 不应依赖 runtime、事件、控件或公开 UI 门面；当前只直接依赖 `theme`。 |
| `theme -> internal/event/ui/widget/layout/style/system` | 未观察到 | theme 应保持最底层语义资源，不应回依赖任何 FluxUI 上层包。 |
| `system -> internal/event/ui/widget/layout/style/theme` | 未观察到 | system 应保持平台能力边界，不应反向持有 UI/runtime 语义。 |
| `widget -> ui/system` | 未观察到 | widget 不应依赖公开 UI 门面或平台系统包，避免组件实现反向污染门面层或系统层。 |
| `internal -> event/ui/widget/layout/system` | 未观察到 | runtime 不应反向调用事件公开层、控件层、布局层、UI 门面或系统能力。 |

#### 疑似反向依赖

| 方向 | 状态 | 证据 | 风险 |
| --- | --- | --- | --- |
| `internal -> style` | 已观察到，疑似反向依赖 | `internal/ripple.go`、`internal/render.go`、`internal/render_cache.go` import `github.com/xiaowumin-mark/FluxUI/style` | `internal` 是 runtime 基础层，依赖 style 会把渲染/状态基础设施和高层样式结构耦合；后续 A2/A4 需要确认 style 数据是否应下沉为基础类型或改由上层传入。 |
| `internal -> theme` | 已观察到，疑似反向依赖 | `internal/context.go`、`internal/runtime.go`、`internal/render.go`、`internal/render_cache.go` import `github.com/xiaowumin-mark/FluxUI/theme` | runtime 直接持有 theme 语义可能让 frame/context 生命周期受主题系统影响；后续 A2 应确认 theme 是否是 Context 公共能力，A4 应确认默认样式来源。 |
| `internal/perf` test -> `widget/style` | 测试专用，暂不计入生产反向依赖 | `internal/perf/bench_test.go` import `internal`、`style`、`widget` | 当前不污染生产图；但性能基准若作为 runtime 指标，应避免长期依赖高层 widget 作为唯一场景。 |

#### 传递依赖观察

- `go list -deps` 显示 `internal` 的传递依赖中包含 `style`、`theme`，这来自生产直接依赖。
- `event` 和 `layout` 虽然只直接依赖 `internal`，但会因 `internal -> style/theme` 间接获得 `style/theme` 传递依赖。
- `ui` 作为聚合层传递依赖覆盖 `event`、`internal`、`layout`、`style`、`system`、`theme`、`widget`，符合公开门面层特征，但后续 A1.2 仍需明确 public API 所有权。
- 生产依赖未形成 Go import cycle；当前风险不是编译循环，而是底层 runtime 语义和高层 style/theme 语义之间的所有权耦合。

#### 事实结论

- 目标包的生产依赖方向总体呈现：`theme/style/internal/event/layout/widget/ui` 分层向上组合，`system` 独立，`ui` 作为公开聚合层。
- 当前最明显的边界风险是 `internal -> style/theme`，与“runtime 基础层应避免被样式/主题层反向约束”的审查主旨存在张力，已标为疑似反向依赖。
- 未观察到 `event -> widget/ui`、`widget -> ui/system`、`style/theme/system -> runtime/widget/ui` 等更严重反向依赖。
- A1.1 只记录依赖事实和风险，不调整 import、不拆包、不修改运行时代码。

#### 验收

- 已输出 `internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system` 的依赖方向图。
- 已标出允许依赖、禁止依赖和疑似反向依赖。
- 已区分生产 import 和测试 import，避免把 `internal/perf` 的 benchmark 依赖误判为 runtime 生产依赖。

#### 后续依赖

- A1.2 public API 所有权清单应重点确认 `internal.Context`、`style.Decoration`、`theme.Theme` 等跨层类型的所有权。
- A2 runtime 生命周期审查应确认 `internal.Context` 和 `internal.Runtime` 是否应该直接持有 theme/style 语义。
- A4 渲染、样式和动画审查应确认 `internal` 中 ripple/render/cache 对 `style/theme` 的依赖是否需要收敛到单向边界。

### 任务编号：A1.2 public API 所有权清单

- 状态：Done
- 日期：2026-07-06 14:42:48 +08:00
- 负责人：Codex
- 关注：Runtime、Event、Widget
- 输入命令：
  - `gopls go_package_api` for `internal`、`event`、`ui`、`widget`、`layout`、`style`、`theme`、`system`
  - `go doc github.com/xiaowumin-mark/FluxUI/internal`
  - `go doc github.com/xiaowumin-mark/FluxUI/event`
  - `go doc github.com/xiaowumin-mark/FluxUI/ui`
  - `go doc github.com/xiaowumin-mark/FluxUI/widget`
  - `go doc github.com/xiaowumin-mark/FluxUI/layout`
  - `go doc github.com/xiaowumin-mark/FluxUI/style`
  - `go doc github.com/xiaowumin-mark/FluxUI/theme`
  - `go doc github.com/xiaowumin-mark/FluxUI/system`
- 输入文件：
  - `internal/*.go`
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
  - public API 所有权
  - 旧 API 兼容边界
  - runtime/event/widget 语义归属

#### 所有权规则

- `internal` 导出符号只在 Go `internal` 可见边界内公开；语义由 runtime 基础层维护，不应被外部用户直接视为稳定公共 API。
- `event` 维护事件语义、事件类型转换、listener 选项、focus/keyboard/pointer/input/drag 事件门面；其中别名到 `internal` 的基础类型应保持行为与 runtime 分发一致。
- `widget` 维护组件实现语义、组件 option、Ref、默认行为、状态更新和组合控件行为。
- `layout` 维护基础布局尺寸语义，不维护 widget 默认行为。
- `style` 维护样式数据结构、装饰、状态层、阴影、图片填充、transform 和交互动画参数。
- `theme` 维护颜色、字体、typography、density、shape、icon font 和 interaction quality 语义。
- `system` 维护平台能力、剪贴板、文件对话框、系统通知、消息框、注册、tray、single instance 和系统事件语义。
- `ui` 维护 Element/Component/Hook/Router/RunElement 等公开门面语义；对 `widget`、`event`、`style`、`theme`、`system`、`app`、`router`、`state`、`anim` 的 re-export 不改变原始语义所有权。

#### 按包归属清单

| 包 | 公开 API 范围 | 语义维护者 |
| --- | --- | --- |
| `internal` | `Runtime`、`Context`、`PathID`、`MemoryKey`、`WindowID`、`WindowController`、window command methods、frame lifecycle、persistent memory、provider、hook store、diagnostics、perf stats、render/cache helpers、runtime event registry、focus/shortcut/pointer capture primitives、`ClickableState`、foundation specs such as `ButtonSpec`/`InputSpec`/`SurfaceSpec` | `internal`。这是 runtime 基础层；其他包可以适配它，但不应重新定义 runtime 生命周期、path identity、memory 或 dispatch 规则。 |
| `event` | `Type`、`Event`、`TargetID`、`Phase`、`ListenerOption`、`On*` listener helpers、`Dispatch*` helpers、`PointerEvent`、`WheelEvent`、`KeyboardEvent`、`InputEvent`、`CompositionEvent`、`DragEvent`、`ActivationEvent`、focus target/boundary/portal/shortcut helpers、Gio event conversion helpers | `event`。浏览器式事件 API、事件字段映射、listener 选项、默认可取消语义的 public facade 归 `event`；底层 path/dispatch 状态仍由 `internal` 执行。 |
| `widget` | `Widget` interface and widget constructors, component option types, Ref types, controlled/uncontrolled component options, layout wrapper widgets, Material 3 widgets, overlay widgets, list/grid/scroll widgets, pointer/keyboard/event boundary widgets, drag/drop widgets, media/image/icon/text helpers | `widget`。组件默认行为、内部状态、Ref 命令消费、callback 触发条件、组件级 layout/render/event 组合语义归 `widget`。 |
| `layout` | `Axis`、`Dimensions`、`FlexChild`、`StackChild`、`Flex`、`Stack`、`Rigid`、`Flexed`、`Stacked` | `layout`。只维护基础约束输入/输出、axis、flex/stack 尺寸语义。 |
| `style` | `Style`、`Decoration`、`Border`、`Insets`、`CornerShape(s)`、`BoxShadow`、`ShadowLayer`、`ImageFill`、`ImageFillFit`、`LinearGradient`、`Transform2D`、`TransformOrigin`、state layer/elevation/image/color/interaction easing helpers | `style`。只维护样式数据和渲染参数语义，不维护 widget 状态或 runtime 生命周期。 |
| `theme` | `Theme`、`ThemeOption`、`ColorScheme`、`ColorOption`、`TextStyle`、`TypeScale`、`DensityScale`、`ShapeScale`、`FontFace`、`FontSpec`、`IconFont`、`InteractionQuality` and font/icon/color discovery/registration helpers | `theme`。维护设计 token、字体、颜色、density、shape 和主题质量等级语义。 |
| `system` | `Capability`/probe APIs, standard system errors, clipboard, file dialog, drag/drop probe, global shortcut, notification, message box, registration, shell open/reveal, single instance, system events, Toast activator, tray APIs | `system`。维护平台能力和 OS 边界语义；不维护 FluxUI widget 或 runtime 交互规则。 |
| `ui` | `Context` alias, `Component`/`Element`/`Widget` facade, Element builders, hooks, router element APIs, app/window run APIs, system convenience wrappers, and re-exported widget/event/style/theme/system/app/router/state/anim APIs | `ui` for Element/hooks/router/run facade only. Re-exported aliases and forwarding constructors keep semantic ownership in their source package. |

#### `internal` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Runtime/context/window | `Runtime`、`NewRuntime`、`Context`、`NewContext`、`Frame`、`WindowID`、`WindowController`、`WindowHiddenMemoryPolicy`、`Context.Window*` methods | `internal` 维护 frame、window command 和 runtime 生命周期语义。A2 必须以这些 API 为事实来源。 |
| Path/memory/provider | `PathID`、`MemoryKey`、`Context.NextKey`、`NextMemoryKey`、`ScopeMemoryKey`、`Persistent*`、`Memo`、`ProviderKey`、`ProviderKeyFor`、`WithProvider*`、`Provider*Value` | `internal` 维护组件身份、路径内存和 provider 作用域语义。 |
| Hooks | `HookKind`、`HookSlot`、`ComponentIdentity`、`ComponentInstance`、`HookStore`、`NewHookStore`、`ShouldRunHookEffect`、`DepsEqual`、`CloneDeps`、`EffectSetup` | `internal` 维护 hook slot 生命周期；`ui` hooks 只是公开 facade。 |
| Events/focus/keyboard | `EventType`、`Event`、`EventPhase`、`EventHandler`、`EventListenerOptions`、`FocusEvent`、`KeyboardEvent`、`Modifiers`、`Shortcut`、`FocusTargetOptions`、`Runtime.Register*`、`Dispatch*`、`RequestFocus`/`MoveFocus` primitives | `internal` 维护真实注册表和 dispatch 执行；`event` 维护 public event facade。 |
| Interaction/perf/diagnostics | `InteractionChangeKind`、`InteractionFrameStats`、`PerfDiagnostics`、`PerfSection`、`FrameStats`、`FrameSectionStats`、`EventDiagnosticsStats`、`VirtualizationStats`、`RenderCacheStats`、`FormatFrameStats` | `internal` 维护 runtime 诊断字段和统计口径。 |
| Render/style bridge | `RippleSpec`、`ShadowSpec`、`SurfaceSpec`、`TextSpec`、`ButtonSpec`、`CheckboxSpec`、`SwitchSpec`、`InputSpec`、`SliderSpec`、`FocusIndicatorSpec`、`DrawCheckMark`、`MixNRGBA` | 目前由 `internal` 维护，但与 A1.1 的 `internal -> style/theme` 风险相关；A4 需确认这些 spec 是否属于 runtime 还是 style/widget。 |
| Foundation layout specs | `Axis`、`Alignment`、`Insets`、`FlexChild`、`StackChild` | 当前由 `internal` 暴露基础表示；`layout`/`widget` 负责公开布局组合语义。 |

#### `event` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Event identity and phases | `Type`、`TargetID`、`Event`、`Phase`、`BoundaryMode`、`BoundaryPolicy`、`ListenerOptions` aliases and constants | `event` 维护 public 命名、兼容和 facade；实际 path/phase 状态来自 `internal`。 |
| Listener registration | `On`、`OnActivate`、`OnFocus`、`OnKeyboard`、`OnKeyDown`、`OnKeyUp`、`OnShortcut`、`OnPointer`、`OnWheel`、`OnInput`、`OnComposition`、`OnDrag`、`Capture`、`Once`、`Passive`、`Priority` | `event` 维护事件注册 API 和 listener option 语义。 |
| Dispatch helpers | `Dispatch`、`DispatchEvent`、`DispatchCustomEvent`、`DispatchKeyboardEvent`、`DispatchPointerEvent`、`DispatchWheelEvent`、`DispatchInputEvent`、`DispatchCompositionEvent`、`DispatchDragEvent`、`DispatchActivationEvent` | `event` 维护 public dispatch facade；runtime dispatch correctness 归 `internal`。 |
| Focus/boundary/portal | `RegisterFocusTarget`、`RequestFocus`、`BlurFocus`、`MoveFocus`、`Focused`、`FocusedTarget`、`FocusManagerFor`、`RegisterBoundary`、`RegisterPortal`、`Focus*` options | `event` 维护 public focus/boundary API；focus registry 归 `internal`。 |
| Pointer/wheel/keyboard/text/drag payloads | `PointerEvent`、`WheelEvent`、`KeyboardEvent`、`InputEvent`、`CompositionEvent`、`DragEvent`、`ActivationEvent`、`Button(s)`、`Modifiers`、`Shortcut`、conversion helpers from Gio | `event` 维护字段映射和 public payload shape。 |
| Legacy compatibility | `Clickable`、`UseClickable`、`ClickHandler`、`HoverHandler` | `event` 维护旧 click/hover bridge 的 public 兼容语义；组件默认行为归 `widget`。 |

#### `widget` 公开 API 所有权

| API 组 | 符号 | 所有权结论 |
| --- | --- | --- |
| Core widget model | `Widget`、`Text`、`Static`、`RenderElement` bridge consumers | `widget` 维护 widget contract 和 component-to-widget rendering behavior。 |
| Basic interaction | `Button` variants、`IconButton` variants、`ClickArea`、`Pressable`、`Card` variants、`FloatingActionButton` variants and their option/ref types | `widget` 维护 click/hover/pressed/disabled/loading/ripple/default action and Ref semantics。 |
| Form controls | `Checkbox`、`Switch`、`RadioGroup`、`Select` variants、`TextField`/`FilledTextField`/`OutlinedTextField`、`SearchBar` and option/ref types | `widget` 维护 controlled value、internal state、OnChange、input/focus/validation/default behavior。 |
| Layout and container widgets | `Row`、`Column`、`Stack`、`Center`、`Padding`、`Fixed*`、`Fill*`、`Flexed`、`Container`、`ContainerDecoration`、`Divider`、`Spacer` | `widget` 维护 widget-facing layout composition; raw constraints rules remain shared with `layout`/`internal`。 |
| Scroll/list/grid/navigation | `ScrollView`、`ListView`、`GridView`/`Grid`、`Tabs`、`NavigationDrawer`、`NavigationRail`、`BottomNavigation` and refs/options | `widget` 维护 offset、virtualization hooks, OnChange and interactive state semantics。 |
| Overlay/feedback | `Dialog`、`Popup`、`DropdownMenu`、`Menu`、`Tooltip`、`Toast`、`Snackbar` and option/ref types | `widget` 维护 overlay mount, outside click, focus, animation and callback semantics。 |
| Event/focus wrappers | `PointerArea`、`KeyboardScope`、`EventBoundary`、`EventPortal` and option types | `widget` 维护 component-level registration boundaries; event dispatch semantics remain in `event/internal`。 |
| Drag/drop | `DragSource`、`DropTarget`、`DragPayload`、`DragOperation`、`DragSourceEvent`、`DropEvent` and option types | `widget` 维护 widget drag/drop lifecycle; platform capability probing remains in `system`。 |
| Media/progress/text/icon | `Image`、`Icon`、`Progress*`、`LoadingIndicator`、`Text` and option/source types | `widget` 维护 visual component behavior; styling primitives remain in `style/theme`。 |

#### `layout`、`style`、`theme`、`system` 所有权摘要

| 包 | 公开 API | 所有权结论 |
| --- | --- | --- |
| `layout` | `Axis`、`Dimensions`、`FlexChild`、`StackChild`、`Flex`、`Stack`、`Rigid`、`Flexed`、`Stacked` | `layout` owns low-level layout measurement semantics and should not define widget-specific behavior。 |
| `style` | `Style`、`Decoration`、`Insets`、`Border`、`CornerShape(s)`、`BoxShadow`、`ImageFill`、`LinearGradient`、`Transform2D`、state/elevation/color/image/easing helpers | `style` owns visual data and transformation semantics, not event state or runtime lifecycle。 |
| `theme` | `Theme`、`ColorScheme`、`DensityScale`、`ShapeScale`、`TypeScale`、`TextStyle`、`FontFace`、`FontSpec`、`IconFont`、`InteractionQuality` and constructors/options | `theme` owns design token and font/color discovery semantics。 |
| `system` | capability/error APIs, clipboard, file dialogs, drag/drop probe, global shortcut, notification, message box, registration, shell, single instance, system events, toast activator, tray | `system` owns OS/platform integration semantics and must stay independent from widget/runtime behavior。 |

#### `ui` facade 所有权矩阵

| `ui` API 来源 | Examples from `go doc ui` | Semantic owner |
| --- | --- | --- |
| Native `ui` Element facade | `Component`、`Element`、`ElementRootBuilder`、`RenderElement`、`ElementKey`、Element builder functions and `RunElement*` | `ui` |
| Hooks/context facade | `Context` alias、`ContextKey`、`NewContextKey`、`UseState`、`UseMemo`、`UseCallback`、`UseEffect*`、`UseMount`、`UseLifecycle`、`UseInterval`、`UseContext`、`UseRef`、`UseAnimatedValue` | `ui` owns facade semantics; state storage and runtime slots are owned by `state/internal/anim` as applicable。 |
| App/window facade | `Run`、`RunMulti`、`App`、`Window`、`WindowElement`、`Window*` functions/options/types | `app` and `internal` own window/runtime execution; `ui` owns public convenience surface。 |
| Widget re-exports | `Widget`、`Button`、`TextField`、`Dialog`、`Popup`、`ScrollView`、`Tabs`、`Select`、`DragSource`、all widget option/ref aliases | `widget` |
| Event re-exports | `Event`、`EventType`/`Type`、`TargetID`、`OnEvent`/`OnActivate`/`OnDrag`、`Dispatch*`、`BoundaryOption`、`Shortcut`、event payload aliases | `event` |
| Style re-exports | `Decoration`、`Style`、`Insets`、`Border`、`CornerShape`、`Bg`、`Pad`、`Shadow`、`TransformDeco`、image/color helpers | `style` |
| Theme re-exports | `Theme`、`ColorScheme`、`TextStyle`、`FontSpec`、`IconFont`、`DensityScale`、theme/color/font constructors and constants | `theme` |
| System wrappers | `OpenFileDialog*`、`SaveFileDialog*`、`ShowMessageBox*`、`CurrentWindowNativeHandle` and system option/result aliases | `system` owns platform semantics; `ui` owns context-aware wrapper convenience。 |
| Router/state/anim re-exports | `RouterElement`、`RouteElement`、`Navigate*`、`UseRoute`、`UseParams`、`State`、`AsyncHandle`、`Animate`、easing/options | `router`、`state`、`anim` respectively; `ui` owns integration surface。 |

#### 风险

- `ui` exposes a very broad facade. Without the ownership matrix above, later fixes may accidentally change `widget`/`event`/`theme` semantics through `ui` wrappers instead of editing the owning package。
- `internal` exports many symbols that are only module-internal by Go rules. They should not be treated as end-user public API, but they are still cross-package contracts for `event`、`layout`、`widget`、`ui` and tests。
- A1.1 identified `internal -> style/theme`; A1.2 confirms several `internal` exported specs are style-like. Ownership of those specs remains a follow-up question for A2/A4, not a fix in this task。
- `event` aliases many `internal` types. Changes to internal event fields can become public event API changes through aliases, so compatibility must be checked in both packages。

#### 事实结论

- Every exported API surfaced by `go doc` can be assigned to an owning package by the table above。
- `ui` is primarily a public facade and re-export layer; most of its widget/event/style/theme/system symbols should not be considered semantically owned by `ui`。
- Runtime lifecycle, path identity, memory, provider, event registry, focus registry and diagnostics are owned by `internal`。
- Event payload and listener facade semantics are owned by `event`; component default actions and Ref behavior are owned by `widget`。

#### 验收

- 已按包记录公开 API 所有权清单。
- 已明确 `ui` re-export 不改变原始包的语义所有权。
- 已标注 `internal` exported API 的特殊边界：模块内部可见，但不是面向最终用户的稳定公共 API。
- 后续修改任一 public API 时，可以从本清单判断应由哪个包维护语义。

#### 后续依赖

- A1.3 escape hatch 边界审查应重点查看 `Context.Gtx`、Gio 原始 event、`system` wrapper 和 `ui` facade 的越界入口。
- A1.4 旧 API 兼容矩阵应基于 `event`/`widget`/`ui` 的所有权划分，尤其是 `OnClick`、`OnHover`、`InputOnChange`、`ScrollOnChange`。
- A2/A4 需要继续处理 `internal` 中 style/theme-like exported specs 的所有权归属风险。
