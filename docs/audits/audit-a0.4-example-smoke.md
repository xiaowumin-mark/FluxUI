# A0.4 示例 smoke 基线

> 本文是 `docs/audits/project-audit-baseline.md` 的子文件，归属 Batch 0：基线冻结。

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

## 执行前工作区状态

| 项目 | 结果 |
| --- | --- |
| `git status --short --untracked-files=all` | 无输出 |
| 判断 | A0.4 执行前工作区干净，没有待解释的代码或文档污染。 |

## 示例入口确认

| 示例 | 路径 | README 状态 | 运行入口 |
| --- | --- | --- | --- |
| docs browser | `examples/docs_browser` | 有 | `go run ./examples/docs_browser` |
| component lab | `examples/component_lab` | 有 | `go run ./examples/component_lab` |
| event system testbench | `examples/event_system_testbench` | 有 | `go run ./examples/event_system_testbench` |
| drag drop showcase | `examples/drag_drop_showcase` | 有 | `go run ./examples/drag_drop_showcase` |

## Smoke 命令

```powershell
go test -run '^$' ./examples/docs_browser ./examples/component_lab ./examples/event_system_testbench ./examples/drag_drop_showcase
```

本次使用 `go test -run '^$'` 做空跑编译 smoke：编译包和测试文件，但不运行测试用例、不启动 GUI 窗口。

## Smoke 输出

```text
ok  	github.com/xiaowumin-mark/FluxUI/examples/docs_browser	0.170s [no tests to run]
ok  	github.com/xiaowumin-mark/FluxUI/examples/component_lab	0.737s [no tests to run]
?   	github.com/xiaowumin-mark/FluxUI/examples/event_system_testbench	[no test files]
?   	github.com/xiaowumin-mark/FluxUI/examples/drag_drop_showcase	[no test files]
```

## 示例可运行状态表

| 示例 | Smoke 状态 | 测试文件状态 | 后续手动验收入口 | 主要覆盖 |
| --- | --- | --- | --- | --- |
| `examples/docs_browser` | 通过 | 有测试文件；本次空跑未执行测试用例 | 是 | 交互式文档主入口，Markdown 文档、示例注册、右侧预览、搜索、分类、API 索引、控件/布局/系统 API demo。 |
| `examples/component_lab` | 通过 | 有测试文件；本次空跑未执行测试用例 | 是 | 组件视觉和交互实验室，覆盖主题、密度、表单、选择、导航、overlay、list、grid、drag/drop、router、Ref 控制。 |
| `examples/event_system_testbench` | 通过 | 无测试文件 | 是 | 事件系统手动测试台，覆盖 P0-P7：旧回调、新事件分发、pointer、focus/keyboard、text input、默认行为、portal、诊断。 |
| `examples/drag_drop_showcase` | 通过 | 无测试文件 | 是 | 生产拖放 API 展示，覆盖 DragSource、DropTarget、payload、生命周期、系统能力探测。 |

## 后续手动验收入口

| 优先级 | 入口 | 手动验收用途 |
| --- | --- | --- |
| P0 | `go run ./examples/docs_browser` | 后续 A1-A13 的文档和示例总验收入口，尤其用于确认文档页滚动、代码块、表格、tabs、chips、示例弹窗和搜索交互。 |
| P0 | `go run ./examples/component_lab` | 后续布局、样式、hover、cursor、组件状态、Ref 和 overlay 组合行为的视觉/交互 smoke 入口。 |
| P0 | `go run ./examples/event_system_testbench` | 后续 A5-A7 事件、焦点、键盘、文本输入、默认行为和诊断链路的手动验收入口。 |
| P1 | `go run ./examples/drag_drop_showcase` | 后续 drag/drop、pointer 冲突、scroll 冲突和系统后端能力的专项验收入口。 |

## 事实结论

- 四个指定示例包均通过空跑编译 smoke，命令退出码为 0。
- 本次未发现示例编译失败，因此没有需要标注为既有问题的失败项。
- `docs_browser` 和 `component_lab` 的测试文件能参与编译；测试用例被 `-run '^$'` 显式跳过，符合本任务的"构建或空跑测试"范围。
- `event_system_testbench` 和 `drag_drop_showcase` 当前没有测试文件；空跑结果证明包可编译，但不等价于 GUI 手动流程已验证。
- 四个示例都应作为后续手动验收入口保留，其中 `docs_browser`、`component_lab`、`event_system_testbench` 为 P0，`drag_drop_showcase` 为 P1。

## 风险

- 空跑编译 smoke 不启动 GUI，不能验证窗口创建、真实鼠标键盘输入、拖放后端、系统通知或平台特性。
- `drag_drop_showcase` 的外部拖入/拖出行为依赖操作系统和桌面后端，本次仅确认包可编译。
- `event_system_testbench` 的 P0-P7 行为需要人工操作观察日志，本次仅冻结可编译入口。
- 后续若示例 smoke 失败，应先与本条基线对比，避免把平台手动行为差异误判为编译回归。

## 验收

- 示例可运行状态表已记录：四个指定示例 smoke 均通过。
- 后续手动验收入口已明确：`docs_browser`、`component_lab`、`event_system_testbench`、`drag_drop_showcase`。
- 本次未发现失败项，没有既有失败被误判为审查引入问题。
- A0.4 未修改示例实现、测试代码或运行时代码。

## 后续依赖

- Batch 1 以后如修改文档示例注册、事件系统、组件状态、overlay、scroll 或 drag/drop，应重跑对应示例 smoke，并按需启动对应手动验收入口。
- G0 关卡可以基于 A0.1-A0.4 判断：工作区、版本、核心测试和示例 smoke 均已记录。
