# 生产治理：API、平台与发布事实来源

本文件是 FluxUI R0 的治理索引。它规定公共 API snapshot、支持矩阵、弃用 ledger、质量门与发布记录之间的权威关系；它不替代组件行为文档或路线图。

## 单一事实来源

| 主题 | 权威文件 | 使用规则 |
| --- | --- | --- |
| 导出的 `ui` 公共 API | [`api/ui.snapshot`](../api/ui.snapshot) | 由 snapshot 流程生成并在 PR 中比较；不得手工把未生成的差异伪装成稳定 API。 |
| 弃用状态、替代项和迁移窗口 | [`docs/deprecation-ledger.md`](deprecation-ledger.md) | 新增/修改 `Deprecated:`、兼容 no-op 或移除计划时必须同一 PR 更新。 |
| 工具链、平台、显示服务器与 native smoke 边界 | 本文件的“支持矩阵” | 版本或验证层级变化必须更新本表和 release checklist。 |
| 面向用户的版本摘要 | [`CHANGELOG.md`](../CHANGELOG.md) | 只作发布说明；术语、状态和替代项必须与 ledger 一致，冲突时以 ledger 为准并在同一变更中修正。 |
| 历史策略说明 | [`docs/deprecation-and-versioning.md`](deprecation-and-versioning.md) | 解释背景，不覆盖 ledger 的当前条目。 |
| 发布执行记录 | [`docs/release-checklist.md`](release-checklist.md) | 每次候选发布必须填充；未填不表示门已通过。 |

所有新增公开符号、签名变更、删除、弃用、兼容 shim、平台承诺或 release gate 变更，都必须在同一 PR 更新对应权威文件。小版本不得静默破坏 `Widget`/`Run` 等既有入口。

## API snapshot 规则

`api/ui.snapshot` 是当前导出 `ui` surface 的可比较清单。它至少需要稳定记录导出名称、类型/函数/方法签名和 Go deprecation 标记；生成器版本与生成 Go 版本应可追溯。

- PR 改变 snapshot 时，必须说明它是新增、兼容变更、弃用还是破坏性变更。
- 新增公开符号必须同步有组件/指南文档、可构建 Docs Browser demo（适用时）、行为测试和 API snapshot。
- 删除或破坏签名需要版本与弃用评审、ledger 条目、迁移说明和 CHANGELOG；仅更新 snapshot 不能视为批准。
- snapshot 差异为零不豁免行为、视觉、race 或平台验证。

## SelectSearchable 的 R0 决定

`SelectSearchable[T](bool)` 在 R0 是**保留编译兼容的 no-op**：它不展示搜索输入、不筛选选项、不提供 query、IME、typeahead 或异步建议能力。不得在 Docs Browser、示例、API 列表或 release note 中把它描述为“可搜索 Select”。

它的当前替代路径是普通 `Select` 的既有选项选择；真实搜索选择能力留给 R1 的 Combobox/Autocomplete，并在其公开 API、测试、Overlay/roving focus 合同和迁移方案准备好后再宣布。该条目与 Go deprecation 标记、snapshot 和 [`deprecation ledger`](deprecation-ledger.md) 必须同步。

相同规则适用于任何预留、实验或尚未生效的 option：在它具备可验证行为前，必须标为 no-op/实验性并进入 ledger，不能借由名称、示例或 release note 暗示能力已可用。

## 支持矩阵

### 固定依赖基线

| 项目 | R0 基线 | 规则 |
| --- | --- | --- |
| Go | `1.25.12` | CI、benchmark 基线和发布制品使用该精确版本；升级须重新跑 snapshot、race、coverage、benchstat 与三平台验证。 |
| Gio | `v0.9.0` | 与 `go.mod` 保持一致；升级须重新审查输入、窗口、GPU/视觉与平台 smoke。 |
| 模块依赖 | 已锁定的 `go.mod`/`go.sum` | 依赖更新必须经过漏洞、许可证/来源与可复现构建检查。 |

README 中的 Go `1.25+` 是开发环境的最低要求；本表的 `1.25.12` 是 CI、benchmark 和发布候选的精确验证基线，不表示任意 Go patch 版本都已完成同等发布验证。

### 目标平台与验证层级

R0 的“支持”指某个 release candidate 已按本表完成对应验证；不是仅在开发者机器上能编译。最低 OS 版本暂不在 R0 单独承诺，发布记录必须注明实际 runner/设备和显示服务器版本。

| 平台 | 持续门（每个受影响 PR） | 发布候选的实际 smoke | 备注 |
| --- | --- | --- | --- |
| Windows | build、test、vet；适用时 race | 窗口启动/关闭、输入/focus、视觉入口及使用到的原生能力 | 原生文件/托盘/通知等按 capability 单独记录。 |
| macOS | build、test、vet；适用时 race | 窗口启动/关闭、输入/focus、视觉入口及使用到的原生能力 | 签名/notarization 按发布渠道启用，不由 mock 代替。 |
| Linux | build、test、vet；适用时 race | 至少一组 X11 与一组 Wayland 的窗口、输入与视觉 smoke | Linux 行不因其中一个 display server 通过而自动覆盖另一个。 |
| Linux / X11 | Linux 基础门 | X11 会话中的实际窗口/输入/视觉 smoke | 将会话、显示服务器和驱动信息写入 release record。 |
| Linux / Wayland | Linux 基础门 | Wayland compositor 中的实际窗口/输入/视觉 smoke | 不能用 X11/XWayland 结果替代原生 Wayland 验证。 |

默认 CI 继续保留 Linux、Windows、macOS 的 build、test、vet 和现有视觉 artifact；R0 新增的格式、lint、核心 race、coverage、依赖/漏洞扫描、snapshot 与 benchmark 门在各自触发条件下记录结果。缺少实际 GUI 环境时，release owner 必须记录为未覆盖风险，而不是把 mock 标记为平台通过。

## Mock、headless 与实际 smoke 的边界

| 验证方式 | 可以证明 | 不能证明 | 使用规则 |
| --- | --- | --- | --- |
| 单元/集成 mock | capability 分支、错误映射、回调顺序、受控状态和无平台逻辑 | OS 窗口、原生输入/IME、系统对话框、拖放、托盘、GPU 或 compositor 行为 | 必须用于可确定逻辑；不能宣称 native support。 |
| headless layout/visual | 约束、绘制入口、固定帧视觉 smoke、Docs Browser 可构建性 | 真正窗口生命周期、显示服务器、GPU 驱动和系统 UI 互操作 | 可作为 CI artifact，不替代实际 smoke。 |
| 实际 native smoke | 对应 OS/display server 的窗口、focus、输入及被触及的系统能力 | 未覆盖的平台或未运行的能力 | 发行候选的必需证据；可自动化，无法自动化时使用明确人工步骤和记录。 |

任何 native capability（文件对话框、剪贴板、拖放、通知、托盘、窗口 chrome、IME 等）必须同时说明：(1) mock 测了什么，(2) 实机/会话 smoke 测了什么，(3) 不可用时向调用方报告的 capability 或错误。mock 失败/通过都不得改变实际平台能力声明。

## R0 质量与性能门

- 格式：所有 Go 改动通过 `gofmt` 检查。
- 静态检查：`go vet` 必须通过；`golangci-lint` 以既定配置对新增/修改 Go 代码执行防回退检查，存量发现不能被新改动扩大。
- 并发：触及 runtime、并发、动画、Hook 或 Ref 的变更运行核心包 `go test -race`。
- 覆盖率：以核心包为统计对象，不用 examples 的零覆盖率稀释；先防回退，新增/修改代码的有效行为分支目标至少 80%，核心包总线先达 60%、v1.0 前 70%。
- 性能：固定 Go `1.25.12`、Gio `v0.9.0`、机器/电源配置、`-benchtime` 和重复次数，使用 `benchstat` 比较基线与候选。热点回归超过 10%，或超过已声明的 8 ms/16 ms 帧预算，必须阻断或取得明确书面批准。
- 供应链：依赖/漏洞扫描、制品 checksum 和 release checklist 是发布门的一部分；发现既有漏洞不能被静默忽略，必须记录风险处置。

执行细节和签字项见 [`docs/release-checklist.md`](release-checklist.md)，高级组件的 demo、行为、视觉和 benchmark 起点见 [`docs/templates/`](templates/)。

## R0 文档完成记录

- 2026-07-12：建立 API snapshot、弃用 ledger、支持矩阵、mock/native 边界和发布检查的权威关系。
- 2026-07-12：固定 Go `1.25.12` 与 Gio `v0.9.0` 基线，定义 Windows/macOS/Linux/X11/Wayland 的 release smoke 要求。
- 本记录仅确认治理文档已建立；每次 release 的实际通过证据必须填入 release checklist，不能由本页替代。
