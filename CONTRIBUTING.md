# 贡献指南

欢迎为 FluxUI 提交 Issue 或 Pull Request。

## 开发流程

```bash
git clone https://github.com/xiaowumin-mark/FluxUI
cd FluxUI
go mod tidy
make test
```

## 提交前检查

| 命令 | 用途 |
|------|------|
| `make test` | 运行全部测试 |
| `make vet` | 静态分析 |
| `make fmt-check` | 检查 `gofmt`，不改写文件 |
| `make fmt` | 格式化代码 |
| `make lint` | 运行仓库锁定的 `golangci-lint` |
| `make race` | 运行核心包 race 测试 |
| `make coverage` | 生成核心包 coverage 报告并执行防回退检查 |
| `make api-check` | 验证 `api/ui.snapshot` |
| `make deps-verify` | 校验模块完整性 |

没有 GNU Make 的环境（例如普通 Windows PowerShell）可直接运行：

```powershell
go run ./tools/gofmtcheck
go run ./tools/api-snapshot -check
go test ./...
go vet ./...
```

## 设计约束

- `ui` 是唯一对外入口，不破坏模块边界
- 新技术能力先在 `ui` 层暴露，底层按依赖方向放置
- 新增能力需补充示例和文档
- 组件实现需遵循 `COMPONENT_DEVELOPMENT_GUIDE.md`
- 代码风格需遵循 `CODE_STYLE_GUIDE.md`
- 功能完成后需执行 `FEATURE_INTEGRATION_CHECKLIST.md` 的关联性检查

## Pull Request 要求

1. 保持模块边界清晰，不跨层依赖
2. 添加对应测试
3. 补充示例（`examples/`）或文档（`docs/widgets/`）
4. 运行受影响的格式、测试、vet、lint、coverage/race/benchmark 门
5. 新增公开 API 时更新 `api/ui.snapshot`、弃用 ledger 和 CHANGELOG
6. 在 PR 描述中说明关联审查、自动测试、手动 smoke、未覆盖风险和局部回滚策略
