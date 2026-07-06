# A0.3 核心测试基线

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 0：基线冻结。

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

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A0.3 执行前工作区干净，没有待解释的代码或文档污染。 |

## 测试命令

```powershell
go test ./event ./internal ./ui ./widget
```

## 测试输出

```text
ok  	github.com/xiaowumin-mark/FluxUI/event	0.467s
ok  	github.com/xiaowumin-mark/FluxUI/internal	0.450s
ok  	github.com/xiaowumin-mark/FluxUI/ui	2.607s
ok  	github.com/xiaowumin-mark/FluxUI/widget	1.054s
```

## 通过/失败清单

| 包 | 结果 | 耗时 | 既有问题判断 |
| --- | --- | ---: | --- |
| `github.com/xiaowumin-mark/FluxUI/event` | 通过 | `0.467s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/internal` | 通过 | `0.450s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/ui` | 通过 | `2.607s` | 无失败项，不涉及既有失败归类。 |
| `github.com/xiaowumin-mark/FluxUI/widget` | 通过 | `1.054s` | 无失败项，不涉及既有失败归类。 |

## 事实结论

- 指定核心测试命令 `go test ./event ./internal ./ui ./widget` 执行成功，退出码为 0。
- `event`、`internal`、`ui`、`widget` 四个核心包全部通过。
- 本次 A0.3 未发现失败项，因此没有需要标注为既有问题的测试失败。
- A0.3 执行前工作区干净；本次审查只追加基线文档，不改动运行时代码或测试代码。

## 风险

- 本任务只覆盖文档指定的核心包测试命令，不代表全仓库 `go test ./...` 已通过。
- 当前结果只能证明 A0.3 时点核心包在 Go `1.25.1 windows/amd64` 和 A0.2 记录的模块集合下通过。
- 后续若出现这些包的失败，应先与本条基线对比，再判断是否由后续审查或修复引入。

## 验收

- 通过/失败清单已记录：四个核心包全部通过。
- 失败项标注已记录：无失败项，因此没有既有失败被误判为审查引入问题。
- 后续问题可以回到 A0.2 的版本快照和本条测试输出复现核心测试基线。

## 后续依赖

- A0.4 示例 smoke 基线可以在核心测试通过的前提下继续执行。
- 若后续审查修改 `event`、`internal`、`ui` 或 `widget`，应至少重跑受影响包测试，并与本基线对照。
