# Release Checklist

本清单用于 FluxUI 的 tag/release candidate。所有复选项都应在 release record 中写入命令、运行环境、artifact 链接或明确的批准人；未填写即未通过。

## 0. 发布身份与范围

- [ ] 版本/tag：`v...`
- [ ] 候选 commit 与干净工作树：`...`
- [ ] release owner / 审核人：`...`
- [ ] 变更范围、兼容性级别和回滚 tag：`...`
- [ ] 固定工具链：Go `1.25.12`、Gio `v0.9.0`，实际 `go version`/`go env` 已记录。

## 1. API、文档与弃用

- [ ] `api/ui.snapshot` 已重新生成并与候选比较；所有差异有分类（新增、兼容、弃用、破坏）。
- [ ] 新增公开 API 有文档、可构建 Docs Browser demo（适用时）、行为测试、视觉状态矩阵和 benchmark/豁免说明。
- [ ] [`docs/deprecation-ledger.md`](deprecation-ledger.md) 与 Go `Deprecated:` 标记一致。
- [ ] `CHANGELOG.md` 与 ledger 的术语、替代项和兼容窗口一致。
- [ ] `SelectSearchable` 未被描述为搜索能力；若 R1 要改变它，已完成独立 API/迁移评审。

## 2. 代码质量与测试

- [ ] `gofmt` 检查通过，且没有未格式化的 Go 文件。
- [ ] `go vet ./...` 通过。
- [ ] `golangci-lint` 使用仓库锁定配置通过。
- [ ] Linux、Windows、macOS 的 build、test、vet 结果均已记录。
- [ ] `go test ./...` 通过；受影响包的定向测试结果已记录。
- [ ] 触及 runtime、并发、动画、Hook 或 Ref 时，核心包 `go test -race` 通过；未运行必须给出风险与批准。
- [ ] 覆盖率报告已生成：核心包总线不低于当前防回退基线；新增/修改有效行为分支目标 ≥80%，examples 不计入稀释分母。

## 3. 视觉与交互

- [ ] 现有 Docs Browser/视觉 smoke 已运行，失败截图或结果 artifact 已保留。
- [ ] 每个受影响高级组件有 Light/Dark、窄/宽、100/125/150/200% DPI、hover/focus/disabled/error/loading、长文本/空状态的矩阵结果或明确豁免。
- [ ] 键盘-only 与 pointer 两条路径均已验证；复合组件包含重排、卸载、scroll hit refresh、Overlay 和同帧受控冲突（适用时）。
- [ ] 视觉基线更新有显式审批，不以“看起来正常”代替断言。

## 4. 性能

- [ ] benchmark 在固定 Go `1.25.12`、Gio `v0.9.0`、相同机器/CPU 电源策略、固定 `-benchtime` 与重复次数下运行。
- [ ] 既有热点至少包括：

  ```powershell
  go test ./widget -run '^$' -bench 'Benchmark(WheelScrollViewVertical|HorizontalWheelDelta)' -benchmem
  go test ./internal/perf -run '^$' -bench 'Benchmark(LayoutStaticTree|MouseMoveInteractiveTree|ListVirtualized|StaticSurfaceCache)' -benchmem
  ```

- [ ] 基线与候选输出已用 `benchstat` 比较，链接/摘要：`...`
- [ ] 热点回归未超过 10%，且没有超过组件声明的 8 ms/16 ms 帧预算；例外有书面批准、owner、期限和回滚条件。
- [ ] 新增的高级组件 benchmark 覆盖可见项规模、alloc、idle redraw 与其最重交互路径。

## 5. 平台与 native smoke

- [ ] Windows：实际窗口启动/关闭、输入/focus、视觉入口及所用 native capability 已 smoke。
- [ ] macOS：实际窗口启动/关闭、输入/focus、视觉入口及所用 native capability 已 smoke。
- [ ] Linux/X11：实际 X11 会话的窗口、输入、视觉 smoke 已记录。
- [ ] Linux/Wayland：实际 Wayland 会话的窗口、输入、视觉 smoke 已记录，未用 XWayland 结果替代。
- [ ] 每项原生能力都记录 mock 覆盖、实际 smoke 覆盖和不可用时 capability/error 行为。
- [ ] 未覆盖的平台/设备/显示服务器已列为 release risk，并有明确批准或阻断结论。

## 6. 供应链与制品

- [ ] 依赖与漏洞扫描（含 `govulncheck` 或仓库规定的等价工具）已运行；发现项有处置、豁免期限或阻断结论。
- [ ] 依赖变更已审查来源、许可证与 `go.sum`；未引入未说明的可执行下载步骤。
- [ ] 三平台可复现构建产物已生成并验证可启动（适用时）。
- [ ] 每个发布制品已发布 SHA-256 checksum；生成命令、文件清单和校验结果已保留。
- [ ] 适用渠道的 provenance/SBOM、Windows 签名、macOS notarization 已完成或有批准的豁免。

## 7. 发布、观测与回滚

- [ ] Tag、release note、制品、checksum 和文档链接指向同一 commit。
- [ ] 崩溃/日志/隐私规则、已知限制和支持矩阵链接已在 release note 中可访问。
- [ ] 回滚方式（撤回制品、回退 tag、兼容 shim）和负责人已记录。
- [ ] 发布后 smoke、问题入口和首次响应 owner 已指定。

## Release completion record

```md
## Release completion record

- 版本 / commit：
- 执行日期与 owner：
- Go / Gio / runner / display server：
- API snapshot 差异：
- 覆盖率与 race 结果：
- benchmark / benchstat 结果：
- Windows / macOS / Linux-X11 / Linux-Wayland smoke：
- 漏洞、依赖、checksum、SBOM/签名结果：
- 已批准的例外与到期日：
- 回滚方式：
```

## R0 文档完成记录

- 2026-07-12：将 API、质量、视觉、性能、平台、供应链和回滚门收敛为可审计的发布清单。
- 实际 release 的“完成”只能由上方记录和签字证明；模板本身不代表任何候选版本已通过。
